package main

import (
	"encoding/json"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	namespace = "http_json"
)

type Exporter struct {
	URL    string
	up     prometheus.Gauge
	value  *prometheus.GaugeVec
	client *http.Client
}

func NewExporter(url string, timeout time.Duration) *Exporter {
	return &Exporter{
		URL: url,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of JSON successful",
		}),
		value: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "value",
			Help:      "JSON value",
		}, []string{"path"}),
		client: &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, timeout)
					if err != nil {
						return nil, err
					}
					if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
						return nil, err
					}
					return c, nil
				},
			},
		},
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up.Desc()
	e.value.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	resp, err := e.client.Get(e.URL)
	if err != nil {
		log.Errorf("Can't scrape Spring Actuator: %v", err)
		return
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		log.Errorf("Can't scrape Spring Actuator: StatusCode: %d", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Reading response body failed %v", err)
		return
	}

	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		log.Fatalf("JSON unmarshaling failed: %s", err)
	}

	intrChan := make(chan StringInterfacePair)
	go func() {
		FlatJson(parsed, intrChan)
	}()

	for pair := range intrChan {
		e.value.WithLabelValues(pair.Key).Set(pair.Value)
		log.Debugf("%s => %s", pair.Key, pair.Value)
	}
	e.value.Collect(ch)
}

type StringInterfacePair struct {
	Key   string
	Value float64
}

func NewStringInterfacePair(key string, value float64) StringInterfacePair {
	return StringInterfacePair{
		Key:   key,
		Value: value,
	}
}

func FlatJson(value interface{}, ch chan<- StringInterfacePair) {
	defer close(ch)
	scanJson("$", value, ch)
}

func scanJson(label string, value interface{}, ch chan<- StringInterfacePair) {
	switch value.(type) {
	case int:
		ch <- NewStringInterfacePair(label, float64(value.(int)))
	case float64:
		ch <- NewStringInterfacePair(label, value.(float64))
	case string:
		log.Debug("Ignore string value. Prometheus can't store string value(in current version@20160521): %s", label)
	case nil:
		log.Debug("Ignore nil value. Prometheus can't store nil value(in current version@20160521): %s", label)
	case map[string]interface{}:
		m := value.(map[string]interface{})
		for k, v := range m {
			if strings.Contains(k, ".") {
				scanJson(label+"['"+k+"']", v, ch)
			} else {
				scanJson(label+"."+k, v, ch)
			}
		}
	case []interface{}:
		for i, v := range value.([]interface{}) {
			scanJson(label+"["+strconv.Itoa(i)+"]", v, ch)
		}
	default:
		panic("Unsupported type in json")
	}
}

func main() {
	var (
		listenAddress     = flag.String("web.listen-address", ":9101", "Address to listen on for web interface and telemetry.")
		metricsPath       = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		actuatorScrapeURI = flag.String("actuator.scrape-uri", "http://localhost/metrics", "HTTP JSON API's URL.")
		timeout           = flag.Duration("actuator.timeout", 5*time.Second, "Timeout for trying to get stats from Spring Actuator.")
	)
	flag.Parse()

	exporter := NewExporter(*actuatorScrapeURI, *timeout)
	prometheus.MustRegister(exporter)

	log.Infof("Starting Server: %s", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>Spring Actuator Exporter</title></head>
		<body>
		<h1>Spring Actuator Exporter</h1>
		<p><a href='` + *metricsPath + `'>Metrics</a></p>
		</body>
		</html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))

}
