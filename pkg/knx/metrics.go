package knx

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"
)

type MetricsExporter struct {
	config *Config
	client GroupClient

	metricsChan    chan *Snapshot
	metrics        MetricSnapshotHandler
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
	logrus.Info("Waiting for incoming knx telegrams...")
	for msg := range e.client.Inbound() {
		e.handleEvent(msg)
	}

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

func (e *MetricsExporter) handleEvent(event knx.GroupEvent) {
	e.messageCounter.WithLabelValues("received", "false").Inc()
	destination := GroupAddress(event.Destination)
	addr, ok := e.config.AddressConfigs[destination]
	if !ok {
		logrus.Tracef("Got ignored %s telegram from %s for %s.",
			event.Command.String(),
			event.Source.String(),
			event.Destination.String())
		return
	}

	v, found := dpt.Produce(addr.DPT)
	if !found {
		logrus.Warnf("Can not find dpt description for \"%s\" to unpack %s telegram from %s for %s.",
			addr.DPT,
			event.Command.String(),
			event.Source.String(),
			event.Destination.String())
		return
	}
	value := v.(DPT)

	if err := value.Unpack(event.Data); err != nil {
		logrus.Warn("Can not unpack data: ", err)
		return
	}

	floatValue, err := extractAsFloat64(value)
	if err != nil {
		logrus.Warn(err)
		return
	}
	metricName := e.config.NameFor(addr)
	logrus.Tracef("Processed value %s for %s on group address %s", value.String(), metricName, destination)
	e.metricsChan <- &Snapshot{
		name:        metricName,
		value:       floatValue,
		source:      PhysicalAddress(event.Source),
		timestamp:   time.Now(),
		config:      &addr,
		destination: destination,
	}
	e.messageCounter.WithLabelValues("received", "true").Inc()
}

func extractAsFloat64(value dpt.DatapointValue) (float64, error) {
	typedValue := reflect.ValueOf(value).Elem()
	kind := typedValue.Kind()
	if kind == reflect.Bool {
		if typedValue.Bool() {
			return 1, nil
		} else {
			return 0, nil
		}
	} else if kind >= reflect.Int && kind <= reflect.Int64 {
		return float64(typedValue.Int()), nil
	} else if kind >= reflect.Uint && kind <= reflect.Uint64 {
		return float64(typedValue.Uint()), nil
	} else if kind >= reflect.Float32 && kind <= reflect.Float64 {
		return typedValue.Float(), nil
	} else {
		return math.NaN(), fmt.Errorf("can not find appropriate type for %s", typedValue.Type().Name())
	}
}
