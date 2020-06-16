package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/timonwong/uwsgi_exporter/exporter"
)

var (
	listenAddress    = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interfaces.").Default(":9117").String()
	metricsPath      = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	statsURI         = kingpin.Flag("stats.uri", "URI for accessing uwsgi stats.").Default("").String()
	statsTimeout     = kingpin.Flag("stats.timeout", "Timeout for trying to get stats from uwsgi.").Default("5s").Duration()
	collectCores     = kingpin.Flag("collect.cores", "Collect cores information per uwsgi worker.").Default("false").Bool()
	applicationLabel = kingpin.Flag("application.label", "Label for WSGI application.").Default("").String()
)

func init() {
	prometheus.MustRegister(version.NewCollector("uwsgi_exporter"))

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("uwsgi_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

}

func main() {

	log.Infoln("Starting uwsgi_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	uwsgiExporter := exporter.NewExporter(*statsURI, *statsTimeout, *collectCores, *applicationLabel)
	prometheus.MustRegister(uwsgiExporter)

	http.Handle(*metricsPath, promhttp.Handler()) // nolint: staticcheck
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>uWSGI Exporter</title></head>
			<body>
			<h1>uWSGI Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
            <h2>Build</h2>
            <pre>` + version.Info() + ` ` + version.BuildContext() + `</pre>
			</body>
			</html>`))
	})
	http.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	log.Infoln("Listening on", *listenAddress)
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
