# http_json_exporter

Export HTTP JSON API response via `http_json_exporter`!

## SYNOPSIS

    $ http_json_exporter

                -actuator.scrape-uri string
                      HTTP JSON API's URL. (default "http://localhost/metrics")
                -actuator.timeout duration
                      Timeout for trying to get stats from Spring Actuator. (default 5s)
                -log.format value
                      If set use a syslog logger or JSON logging. Example: logger:syslog?appname=bob&local=7 or logger:stdout?json=true. Defaults to stderr.
                -log.level value
                      Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]. (default info)
                -web.listen-address string
                      Address to listen on for web interface and telemetry. (default ":9101")
                -web.telemetry-path string
                      Path under which to expose metrics. (default "/metrics")

## Sample

If HTTP API returns following:

```
{
  "foo": 1,
  "bar": {
    "baz": 2
  },
  "yo": null,
  "i": "ppp"
}
```

Exporter returns values like following:

    # HELP http_json_value JSON value
    # TYPE http_json_value gauge
    http_json_value{path="$.bar.baz"} 2
    http_json_value{path="$.foo"} 1

## The spec of path label

The 'path' label contains JSON path for.

## LICENSE

    The MIT License (MIT)
    Copyright © 2016 Tokuhiro Matsuno, http://64p.org/ <tokuhirom@gmail.com>

    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the “Software”), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in
    all copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
    OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
    THE SOFTWARE.

