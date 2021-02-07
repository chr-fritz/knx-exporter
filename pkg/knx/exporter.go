package knx

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/vapourismo/knx-go/knx"
)

type MetricsExporter struct {
	config *Config
	client GroupClient

	metricsChan    chan *Snapshot
	metrics        MetricSnapshotHandler
	listener       Listener
	messageCounter *prometheus.CounterVec
	health         error
}

func NewMetricsExporter(configFile string, registerer prometheus.Registerer) (*MetricsExporter, error) {
	config, err := ReadConfig(configFile)
	if err != nil {
		return nil, err
	}
	m := &MetricsExporter{
		config:      config,
		metrics:     NewMetricsSnapshotHandler(registerer),
		metricsChan: make(chan *Snapshot),
		messageCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      "messages",
			Namespace: "knx",
		}, []string{"direction", "processed"}),
	}
	_ = registerer.Register(m.messageCounter)
	return m, nil
}

func (e *MetricsExporter) Run() error {
	if err := e.createClient(); err != nil {
		e.health = err
		return err
	}

	go e.storeSnapshots()

	e.listener = NewListener(e.config, e.client.Inbound(), e.metrics.GetMetricsChannel(), e.messageCounter)
	go e.listener.Run()
	return nil
}

func (e *MetricsExporter) Close() {
	e.client.Close()
	close(e.metricsChan)
}

func (e *MetricsExporter) IsAlive() error {
	return e.health
}

func (e *MetricsExporter) createClient() error {
	switch e.config.Connection.Type {
	case Tunnel:
		logrus.Infof("Connect to %s using tunneling", e.config.Connection.Endpoint)
		tunnel, err := knx.NewGroupTunnel(e.config.Connection.Endpoint, knx.DefaultTunnelConfig)
		if err != nil {
			return err
		}
		e.client = &tunnel
		return nil
	case Router:
		logrus.Infof("Connect to %s using multicast routing", e.config.Connection.Endpoint)
		router, err := knx.NewGroupRouter(e.config.Connection.Endpoint, knx.DefaultRouterConfig)
		if err != nil {
			return err
		}
		e.client = &router
		return nil
	default:
		return fmt.Errorf("invalid connection type. must be either Tunnel or Router")
	}
}

func (e *MetricsExporter) storeSnapshots() {
	for snap := range e.metricsChan {
		e.metrics.AddSnapshot(snap)
	}
}
