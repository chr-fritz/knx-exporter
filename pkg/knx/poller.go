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
	"log/slog"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
)

// Poller defines the interface for active polling for metrics values against the knx system.
type Poller interface {
	// Run starts the polling.
	Run()
	// Close stops the polling.
	Close()
}

type poller struct {
	client          GroupClient
	config          *Config
	messageCounter  *prometheus.CounterVec
	snapshotHandler MetricSnapshotHandler
	pollingInterval time.Duration
	metricsToPoll   GroupAddressConfigSet
	ticker          *time.Ticker
}

// NewPoller creates a new Poller instance using the given MetricsExporter for connection handling and metrics observing.
func NewPoller(config *Config, client GroupClient, metricsHandler MetricSnapshotHandler, messageCounter *prometheus.CounterVec) Poller {
	metricsToPoll := getMetricsToPoll(config)
	interval := calcPollingInterval(metricsToPoll)
	return &poller{
		client:          client,
		config:          config,
		messageCounter:  messageCounter,
		pollingInterval: interval,
		snapshotHandler: metricsHandler,
		metricsToPoll:   metricsToPoll,
	}
}

func (p *poller) Run() {
	if p.pollingInterval <= 0 {
		return
	}
	slog.Log(nil, slog.LevelDebug-2, "Start polling group addresses", "pollingInterval", p.pollingInterval)
	p.ticker = time.NewTicker(p.pollingInterval)
	c := p.ticker.C
	go func() {
		for t := range c {
			p.pollAddresses(t)
		}
	}()
}

func (p *poller) Close() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}

func (p *poller) pollAddresses(t time.Time) {
	for address, config := range p.metricsToPoll {
		logger := slog.With("address", address)
		s := p.snapshotHandler.FindYoungestSnapshot(config.Name)
		if s == nil {
			logger.Log(nil, slog.LevelDebug-2, "Initial polling of address")
			p.sendReadMessage(address, config)
			continue
		}

		diff := t.Sub(s.timestamp).Round(time.Second)
		maxAge := time.Duration(config.MaxAge)
		if diff >= maxAge {
			logger.Log(nil, slog.LevelDebug-2,
				"Poll address for new value as it is to old",
				"maxAge", maxAge,
				"actualAge", diff,
			)
			p.sendReadMessage(address, config)
		}
	}
}

func (p *poller) sendReadMessage(address GroupAddress, config *GroupAddressConfig) {
	event := knx.GroupEvent{
		Command: knx.GroupRead,
		Source:  cemi.IndividualAddr(p.config.Connection.PhysicalAddress),
	}

	if config.ReadType == WriteOther {
		event.Command = knx.GroupWrite
		event.Destination = cemi.GroupAddr(config.ReadAddress)
		event.Data = config.ReadBody
	} else {
		event.Destination = cemi.GroupAddr(address)
	}

	if e := p.client.Send(event); e != nil {
		slog.Info("Can not send read request: "+e.Error(), "address", address.String())
	}
	p.messageCounter.WithLabelValues("sent", "true").Inc()
}

func getMetricsToPoll(config *Config) GroupAddressConfigSet {
	toPoll := make(GroupAddressConfigSet)
	for address, addressConfig := range config.AddressConfigs {
		interval := time.Duration(addressConfig.MaxAge).Truncate(time.Second)
		if !addressConfig.Export || !addressConfig.ReadActive || interval < time.Second {
			continue
		}

		interval = time.Duration(math.Max(float64(interval), float64(5*time.Second)))
		toPoll[address] = &GroupAddressConfig{
			Name:        config.NameFor(addressConfig),
			ReadActive:  true,
			ReadType:    addressConfig.ReadType,
			ReadAddress: addressConfig.ReadAddress,
			ReadBody:    addressConfig.ReadBody,
			MaxAge:      Duration(interval),
		}
	}
	return toPoll
}

func calcPollingInterval(config GroupAddressConfigSet) time.Duration {
	var intervals []time.Duration
	for _, ga := range config {
		intervals = append(intervals, time.Duration(ga.MaxAge))
	}
	if len(intervals) == 0 {
		return -1
	} else if len(intervals) == 1 {
		return intervals[0]
	}

	ggt := int64(intervals[0].Seconds())
	for i := 1; i < len(intervals); i++ {
		ggt = gcd(ggt, int64(intervals[i].Seconds()))
	}
	return time.Duration(ggt) * time.Second
}

// greatest common divisor (GCD) via Euclidean algorithm
func gcd(a, b int64) int64 {
	for b != 0 {
		t := b
		b = a % b
		a = t
	}
	return a
}
