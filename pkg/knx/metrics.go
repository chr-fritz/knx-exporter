package knx

import (
	"fmt"
	"io/ioutil"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ghodss/yaml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"
)

type MetricsExporter struct {
	configFile string
	config     *Config
	client     GroupClient

	metricsChan  chan metricSnapshot
	snapshotLock sync.RWMutex
	metrics      map[string]metricSnapshot
}

type metricSnapshot struct {
	name       string
	value      float64
	timestamp  time.Time
	metricType string
}

func NewMetricsExporter(configFile string) (*MetricsExporter, error) {
	m := &MetricsExporter{
		configFile:   configFile,
		snapshotLock: sync.RWMutex{},
		metrics:      map[string]metricSnapshot{},
		metricsChan:  make(chan metricSnapshot),
	}
	if err := m.readConfig(); err != nil {
		return nil, err
	}

	return m, nil
}

func (e *MetricsExporter) Run() error {
	if err := e.createClient(); err != nil {
		return err
	}

	go e.storeSnapshots()
	for msg := range e.client.Inbound() {
		e.handleEvent(msg)
	}

	return nil
}

func (e *MetricsExporter) Close() {
	e.client.Close()
	close(e.metricsChan)
}

func (e *MetricsExporter) readConfig() error {
	content, err := ioutil.ReadFile(e.configFile)
	if err != nil {
		return err
	}
	config := Config{}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return fmt.Errorf("can not read config file %s: %s", e.configFile, err)
	}
	e.config = &config
	return nil
}

func (e *MetricsExporter) createClient() error {
	switch e.config.Connection.Type {
	case Tunnel:
		tunnel, err := knx.NewGroupTunnel(e.config.Connection.Endpoint, knx.DefaultTunnelConfig)
		if err != nil {
			return err
		}
		e.client = &tunnel
		return nil
	case Router:
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
	for snapshot := range e.metricsChan {
		e.snapshotLock.Lock()
		e.metrics[snapshot.name] = snapshot
		e.snapshotLock.Unlock()
	}
}

func (e *MetricsExporter) RegisterMetrics() []prometheus.Collector {
	var metrics []prometheus.Collector
	for ga, gaConfig := range e.config.AddressConfigs {
		if !gaConfig.Export {
			continue
		}
		name := e.config.MetricsPrefix + gaConfig.Name
		var metric prometheus.Collector
		if strings.ToLower(gaConfig.MetricType) == "counter" {
			metric = prometheus.NewCounterFunc(
				prometheus.CounterOpts{
					Name: name,
					Help: fmt.Sprintf("Value of %s\n%s", ga.String(), gaConfig.Comment),
				},
				e.getMetricsValue(name),
			)
		} else if strings.ToLower(gaConfig.MetricType) == "gauge" {
			metric = prometheus.NewGaugeFunc(
				prometheus.GaugeOpts{
					Name: name,
					Help: fmt.Sprintf("Value of %s\n%s", ga.String(), gaConfig.Comment),
				},
				e.getMetricsValue(name),
			)
		}

		if metric != nil {
			metrics = append(metrics, metric)
		}
	}
	return metrics
}

func (e *MetricsExporter) getMetricsValue(metric string) func() float64 {
	return func() float64 {
		e.snapshotLock.RLock()
		defer e.snapshotLock.RUnlock()
		snapshot, ok := e.metrics[metric]
		if !ok {
			return math.NaN()
		}
		return snapshot.value
	}
}

func (e *MetricsExporter) handleEvent(event knx.GroupEvent) {
	destination := GroupAddress(event.Destination)
	addr, ok := e.config.AddressConfigs[destination]
	if !ok {
		logrus.Tracef("Got ignored %s telegram from %s for %s.",
			event.Command.String(),
			event.Source.String(),
			event.Destination.String())
		return
	}

	value, found := dpt.Produce(addr.DPT)
	if !found {
		logrus.Warnf("Can not find dpt description to unpack %s telegram from %s for %s.",
			event.Command.String(),
			event.Source.String(),
			event.Destination.String())
		return
	}

	if err := value.Unpack(event.Data); err != nil {
		logrus.Warn("Can not unpack data: ", err)
		return
	}

	floatValue, err := extractAsFloat64(value)
	if err != nil {
		logrus.Warn(err)
		return
	}

	e.metricsChan <- metricSnapshot{
		name:       e.config.MetricsPrefix + addr.Name,
		value:      floatValue,
		timestamp:  time.Now(),
		metricType: addr.MetricType,
	}
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
