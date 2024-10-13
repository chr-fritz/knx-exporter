// Copyright © 2020-2024 Christian Fritz <mail@chr-fritz.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package knx

import (
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"
)

type Listener interface {
	Run()
	IsActive() bool
}

type listener struct {
	config         *Config
	inbound        <-chan knx.GroupEvent
	metricsChan    chan *Snapshot
	messageCounter *prometheus.CounterVec
	active         bool
	logger         *slog.Logger
}

func NewListener(config *Config, inbound <-chan knx.GroupEvent, metricsChan chan *Snapshot, messageCounter *prometheus.CounterVec) Listener {
	return &listener{
		config:         config,
		inbound:        inbound,
		metricsChan:    metricsChan,
		messageCounter: messageCounter,
		active:         true,
		logger: slog.With(
			"connectionType", config.Connection.Type,
			"endpoint", config.Connection.Endpoint,
		),
	}
}

func (l *listener) Run() {
	l.logger.Info("Waiting for incoming knx telegrams...")
	defer func() {
		l.active = false
	}()
	for msg := range l.inbound {
		l.handleEvent(msg)
	}
	l.logger.Warn("Finished listening for incoming knx telegrams")
}

func (l *listener) IsActive() bool {
	return l.active
}

func (l *listener) handleEvent(event knx.GroupEvent) {
	l.messageCounter.WithLabelValues("received", "false").Inc()
	destination := GroupAddress(event.Destination)
	logger := l.logger.With(
		"command", event.Command.String(),
		"source", event.Source.String(),
		"destination", event.Destination.String(),
	)

	addr, ok := l.config.AddressConfigs[destination]
	if !ok {
		logger.Debug("Received event but ignore them due to missing configuration")
		return
	}

	if event.Command == knx.GroupRead {
		logger.Debug("Skip group event as it is a GroupRead message.")
		return
	}

	value, err := unpackEvent(event, addr)
	logger = logger.With("dpt", addr.DPT)

	if err != nil {
		logger.Warn(err.Error())
		return
	}

	floatValue, err := extractAsFloat64(value)
	if err != nil {
		logger.Warn(err.Error())
		return
	}
	metricName := l.config.NameFor(addr)
	logger.With(
		"metricName", metricName,
		"value", value,
	).Log(nil, slog.LevelDebug-2, "Processed received group address value")
	l.metricsChan <- &Snapshot{
		name:        metricName,
		value:       floatValue,
		source:      PhysicalAddress(event.Source),
		timestamp:   time.Now(),
		config:      addr,
		destination: destination,
	}
	l.messageCounter.WithLabelValues("received", "true").Inc()
}

func unpackEvent(event knx.GroupEvent, addr *GroupAddressConfig) (DPT, error) {
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
