// Copyright © 2024 Christian Fritz <mail@chr-fritz.de>
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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
)

// StartupReader defines the interface for active polling for metrics values against the knx system at startup.
type StartupReader interface {
	// Run starts the startup reading.
	Run()
	// Close stops the startup reading.
	Close()
}

type startupReader struct {
	client          GroupClient
	config          *Config
	messageCounter  *prometheus.CounterVec
	snapshotHandler MetricSnapshotHandler
	metricsToRead   GroupAddressConfigSet
	ticker          *time.Ticker
}

// NewStartupReader creates a new StartupReader instance using the given MetricsExporter for connection handling and metrics observing.
func NewStartupReader(config *Config, client GroupClient, metricsHandler MetricSnapshotHandler, messageCounter *prometheus.CounterVec) StartupReader {
	metricsToRead := getMetricsToRead(config)
	return &startupReader{
		client:          client,
		config:          config,
		messageCounter:  messageCounter,
		snapshotHandler: metricsHandler,
		metricsToRead:   metricsToRead,
	}
}

func (s *startupReader) Run() {
	readInterval := time.Duration(s.config.ReadStartupInterval)
	if readInterval.Milliseconds() <= 0 {
		readInterval = 200 * time.Millisecond
	}
	slog.Info("start reading addresses after startup.", "delay", readInterval)
	s.ticker = time.NewTicker(readInterval)
	go func() {
		for address, config := range s.metricsToRead {
			<-s.ticker.C
			s.sendReadMessage(address, config)
		}
		s.ticker.Stop()
	}()
}

func (s *startupReader) Close() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
}

func (s *startupReader) sendReadMessage(address GroupAddress, config *GroupAddressConfig) {
	event := knx.GroupEvent{
		Command: knx.GroupRead,
		Source:  cemi.IndividualAddr(s.config.Connection.PhysicalAddress),
	}

	if config.ReadType == WriteOther {
		event.Command = knx.GroupWrite
		event.Destination = cemi.GroupAddr(config.ReadAddress)
		event.Data = config.ReadBody
	} else {
		event.Destination = cemi.GroupAddr(address)
	}

	if e := s.client.Send(event); e != nil {
		slog.Error("can not send read request: "+e.Error(), "address", address)
	}
	s.messageCounter.WithLabelValues("sent", "true").Inc()
}

func getMetricsToRead(config *Config) GroupAddressConfigSet {
	toRead := make(GroupAddressConfigSet)
	for address, addressConfig := range config.AddressConfigs {
		if !addressConfig.Export || !addressConfig.ReadStartup {
			continue
		}

		toRead[address] = &GroupAddressConfig{
			Name:        config.NameFor(addressConfig),
			ReadStartup: true,
			ReadType:    addressConfig.ReadType,
			ReadAddress: addressConfig.ReadAddress,
			ReadBody:    addressConfig.ReadBody,
		}
	}
	return toRead
}
