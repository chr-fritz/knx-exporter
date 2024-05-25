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

package metrics

//go:generate mockgen -destination=fake/exporterMocks.go -package=fake -source=exporter.go

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type exporter struct {
	Port          uint16
	health        healthcheck.Handler
	meterRegistry *prometheus.Registry
	server        *http.Server
}

type Exporter interface {
	Run() error
	Shutdown() error
	MustRegister(collectors ...prometheus.Collector)
	Register(collector prometheus.Collector) error
	Unregister(collector prometheus.Collector) bool
	AddLivenessCheck(name string, check healthcheck.Check)
	AddReadinessCheck(name string, check healthcheck.Check)
}

func NewExporter(port uint16, withGoMetrics bool) Exporter {
	registry := prometheus.DefaultRegisterer.(*prometheus.Registry)
	if !withGoMetrics {
		registry = prometheus.NewPedanticRegistry()
	}
	return &exporter{
		Port:          port,
		health:        healthcheck.NewHandler(),
		meterRegistry: registry,
		server:        &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", port)},
	}
}

func (e exporter) Run() error {
	server := http.NewServeMux()

	server.HandleFunc("/live", e.health.LiveEndpoint)
	server.HandleFunc("/ready", e.health.ReadyEndpoint)
	handler := promhttp.Handler()
	server.Handle("/metrics", handler)
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)

	e.server.Handler = server
	return e.server.ListenAndServe()
}
func (e exporter) Shutdown() error {
	ctx := context.Background()
	return e.server.Shutdown(ctx)
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
func (e exporter) AddLivenessCheck(name string, check healthcheck.Check) {
	e.health.AddLivenessCheck(name, check)
}
func (e exporter) AddReadinessCheck(name string, check healthcheck.Check) {
	e.health.AddReadinessCheck(name, check)
}
