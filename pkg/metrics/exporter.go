package metrics

import (
	"fmt"
	"net/http"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type exporter struct {
	Port          uint16
	health        healthcheck.Handler
	meterRegistry prometheus.Registerer
}

type Exporter interface {
	Run() error
	MustRegister(collectors ...prometheus.Collector)
	Register(collector prometheus.Collector) error
	Unregister(collector prometheus.Collector) bool
}

func NewExporter(port uint16) Exporter {
	return &exporter{
		Port:          port,
		health:        healthcheck.NewHandler(),
		meterRegistry: prometheus.NewRegistry(),
	}
}

func (e exporter) Run() error {
	server := http.NewServeMux()
	listenAddr := fmt.Sprintf("0.0.0.0:%d", e.Port)

	server.HandleFunc("/live", e.health.LiveEndpoint)
	server.HandleFunc("/ready", e.health.ReadyEndpoint)
	handler := promhttp.InstrumentMetricHandler(e.meterRegistry, promhttp.Handler())
	server.Handle("/metrics", handler)
	return http.ListenAndServe(listenAddr, server)
}
func (e exporter) MustRegister(collectors ...prometheus.Collector) {
	e.meterRegistry.MustRegister(collectors...)
}
func (e exporter) Register(collector prometheus.Collector) error {
	return e.meterRegistry.Register(collector)
}
func (e exporter) Unregister(collector prometheus.Collector) bool {
	return e.meterRegistry.Unregister(collector)
}
