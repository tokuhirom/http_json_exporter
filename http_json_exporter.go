package main

import (
	"encoding/json"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/tokuhirom/json_path_scanner"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	namespace = "http_json"
)

type Exporter struct {
	URL    string
	mutex  sync.RWMutex
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

func (e *Exporter) fetch() (*[]byte, error) {
	resp, err := e.client.Get(e.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	body, err := e.fetch()
	if err != nil {
		return err
	}

	var parsed interface{}
	if err := json.Unmarshal(*body, &parsed); err != nil {
		return err
	}

	intrChan := make(chan json_path_scanner.PathValue)
	go func() {
		json_path_scanner.Scan(parsed, intrChan)
	}()

	for pair := range intrChan {
		log.Debugf("%s => %s", pair.Path, pair.Value)

		switch pair.Value.(type) {
		case int:
			e.value.WithLabelValues(pair.Path).Set(float64(pair.Value.(int)))
		case float64:
			e.value.WithLabelValues(pair.Path).Set(pair.Value.(float64))
		case string:
			log.Debugf("Skip. Prometheus can't handle string.")
		case nil:
			log.Debugf("Skip. Prometheus can't handle nil.")
		}
	}
	e.value.Collect(ch)

	return nil
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	err := e.collect(ch)
	if err != nil {
		log.Warn(err)
		e.up.Set(0)
	} else {
		e.up.Set(1)
	}
	e.up.Collect(ch)
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
