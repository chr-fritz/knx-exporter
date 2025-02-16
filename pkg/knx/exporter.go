// Copyright © 2020-2025 Christian Fritz <mail@chr-fritz.de>
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
	"context"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vapourismo/knx-go/knx"
)

type MetricsExporter interface {
	Run(ctx context.Context) error
	IsAlive() error
}

type metricsExporter struct {
	config *Config
	client GroupClient

	metrics        MetricSnapshotHandler
	listener       Listener
	messageCounter *prometheus.CounterVec
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
		metrics: NewMetricsSnapshotHandler(),
		messageCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      "messages",
			Namespace: "knx",
		}, []string{"direction", "processed"}),
	}
	if err = registerer.Register(m.messageCounter); err != nil {
		return nil, fmt.Errorf("can not register message counter metrics: %s", err)
	}
	if err = registerer.Register(m.metrics); err != nil {
		return nil, fmt.Errorf("can not register metrics collector: %s", err)
	}
	return m, nil
}

func (e *metricsExporter) Run(ctx context.Context) error {
	e.poller = NewPoller(e.config, e.metrics, e.messageCounter)
	e.listener = NewListener(e.config, e.metrics.GetMetricsChannel(), e.messageCounter)
	go e.metrics.Run(ctx)

	if err := e.createClient(); err != nil {
		e.health = err
		return err
	}
	context.AfterFunc(ctx, func() {
		e.client.Close()
	})

	go e.listener.Run(ctx, e.client.Inbound())
	e.poller.Run(ctx, e.client, true)

	return nil
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
		slog.With(
			"endpoint", e.config.Connection.Endpoint,
			"connection_type", "tunnel",
			"useTcp", e.config.Connection.TunnelConfig.UseTCP,
		).Info("Connecting to endpoint")
		tunnel, err := knx.NewGroupTunnel(e.config.Connection.Endpoint, e.config.Connection.TunnelConfig.toKnxTunnelConfig())
		if err != nil {
			return err
		}
		e.client = &tunnel
		return nil
	case Router:
		slog.With(
			"endpoint", e.config.Connection.Endpoint,
			"connection_type", "routing",
		).Info("Connecting to endpoint")

		config, err := e.config.Connection.RouterConfig.toKnxRouterConfig()
		if err != nil {
			return fmt.Errorf("unable to convert router config: %s", err)
		}

		router, err := knx.NewGroupRouter(e.config.Connection.Endpoint, config)
		if err != nil {
			return err
		}
		e.client = &router
		return nil
	default:
		return fmt.Errorf("invalid connection type. must be either Tunnel or Router")
	}
}
