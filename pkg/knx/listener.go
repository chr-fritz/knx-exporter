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

type Listener interface {
	Run()
}

type listener struct {
	config         *Config
	inbound        <-chan knx.GroupEvent
	metricsChan    chan *Snapshot
	messageCounter *prometheus.CounterVec
}

func NewListener(config *Config, inbound <-chan knx.GroupEvent, metricsChan chan *Snapshot, messageCounter *prometheus.CounterVec) Listener {
	return &listener{
		config:         config,
		inbound:        inbound,
		metricsChan:    metricsChan,
		messageCounter: messageCounter,
	}
}

func (l *listener) Run() {
	logrus.Info("Waiting for incoming knx telegrams...")
	for msg := range l.inbound {
		l.handleEvent(msg)
	}
}

func (l *listener) handleEvent(event knx.GroupEvent) {
	l.messageCounter.WithLabelValues("received", "false").Inc()
	destination := GroupAddress(event.Destination)

	addr, ok := l.config.AddressConfigs[destination]
	if !ok {
		logrus.Tracef("Got ignored %s telegram from %s for %s.",
			event.Command.String(),
			event.Source.String(),
			event.Destination.String())
		return
	}

	value, err := unpackEvent(event, addr)
	if err != nil {
		logrus.Warn(err)
		return
	}

	floatValue, err := extractAsFloat64(value)
	if err != nil {
		logrus.Warn(err)
		return
	}
	metricName := l.config.NameFor(addr)
	logrus.Tracef("Processed value %s for %s on group address %s", value.String(), metricName, destination)
	l.metricsChan <- &Snapshot{
		name:        metricName,
		value:       floatValue,
		source:      PhysicalAddress(event.Source),
		timestamp:   time.Now(),
		config:      &addr,
		destination: destination,
	}
	l.messageCounter.WithLabelValues("received", "true").Inc()
}

func unpackEvent(event knx.GroupEvent, addr GroupAddressConfig) (DPT, error) {
	v, found := dpt.Produce(addr.DPT)
	if !found {
		return nil, fmt.Errorf("can not find dpt description for \"%s\" to unpack %s telegram from %s for %s",
			addr.DPT,
			event.Command.String(),
			event.Source.String(),
			event.Destination.String())

	}
	value := v.(DPT)

	if err := value.Unpack(event.Data); err != nil {
		return nil, fmt.Errorf("can not unpack data: %s", err)
	}
	return value, nil
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
