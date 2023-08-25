// Copyright © 2020-2022 Christian Fritz <mail@chr-fritz.de>
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

//go:generate mockgen -destination=fake/exporterMocks.go -package=fake -source=exporter.go

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/vapourismo/knx-go/knx"
)

type MetricsExporter interface {
	Run() error
	Close()
	IsAlive() error
}

type metricsExporter struct {
	config *Config
	client GroupClient

	metrics        MetricSnapshotHandler
	listener       Listener
	messageCounter *prometheus.CounterVec
	startupReader  StartupReader
	poller         Poller
	health         error
}

func NewMetricsExporter(configFile string, registerer prometheus.Registerer) (MetricsExporter, error) {
	config, err := ReadConfig(configFile)
	if err != nil {
		return nil, err
	}
	m := &metricsExporter{
		config:  config,
		metrics: NewMetricsSnapshotHandler(registerer),
		messageCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      "messages",
			Namespace: "knx",
		}, []string{"direction", "processed"}),
	}
	_ = registerer.Register(m.messageCounter)
	return m, nil
}

func (e *metricsExporter) Run() error {
	if err := e.createClient(); err != nil {
		e.health = err
		return err
	}

	e.startupReader = NewStartupReader(e.config, e.client, e.metrics, e.messageCounter)
	e.startupReader.Run()

	e.poller = NewPoller(e.config, e.client, e.metrics, e.messageCounter)
	e.poller.Run()

	go e.metrics.Run()

	e.listener = NewListener(e.config, e.client.Inbound(), e.metrics.GetMetricsChannel(), e.messageCounter)
	go e.listener.Run()
	return nil
}

func (e *metricsExporter) Close() {
	if e.startupReader != nil {
		e.startupReader.Close()
	}
	if e.poller != nil {
		e.poller.Close()
	}
	if e.client != nil {
		e.client.Close()
	}
	if e.metrics != nil {
		e.metrics.Close()
	}
}

func (e *metricsExporter) IsAlive() error {
	if !e.listener.IsActive() {
		return fmt.Errorf("listener is closed")
	}
	if !e.metrics.IsActive() {
		return fmt.Errorf("metric snapshot handler is closed")
	}

	return e.health
}

func (e *metricsExporter) createClient() error {
	switch e.config.Connection.Type {
	case Tunnel:
		logrus.WithField("endpoint", e.config.Connection.Endpoint).
			WithField("connection_type", "tunnel").
			Infof("Connect to %s using tunneling", e.config.Connection.Endpoint)
		tunnel, err := knx.NewGroupTunnel(e.config.Connection.Endpoint, knx.DefaultTunnelConfig)
		if err != nil {
			return err
		}
		e.client = &tunnel
		return nil
	case Router:
		logrus.WithField("endpoint", e.config.Connection.Endpoint).
			WithField("connection_type", "routing").
			Infof("Connect to %s using multicast routing", e.config.Connection.Endpoint)
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
