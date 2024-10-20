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

//go:generate mockgen -destination=snapshotMocks_test.go -package=knx -self_package=github.com/chr-fritz/knx-exporter/pkg/knx -source=snapshot.go

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricSnapshotHandler holds and manages all the snapshots of metrics.
type MetricSnapshotHandler interface {
	// AddSnapshot adds a new snapshot that should be exported as metric.
	AddSnapshot(snapshot *Snapshot)
	// FindSnapshot finds a given snapshot by the snapshots key.
	FindSnapshot(key SnapshotKey) (*Snapshot, error)
	// FindYoungestSnapshot finds the youngest snapshot with the given metric name.
	// It don't matter from which device the snapshot was received.
	FindYoungestSnapshot(name string) *Snapshot
	// GetValueFunc returns a function that returns the current value for the given snapshot key.
	GetValueFunc(key SnapshotKey) func() float64
	// Run let the MetricSnapshotHandler listen for new snapshots on the Snapshot channel.
	Run()
	// GetMetricsChannel returns the channel to send new snapshots to this MetricSnapshotHandler.
	GetMetricsChannel() chan *Snapshot
	// Close stops listening for new Snapshots and closes the Snapshot channel.
	Close()
	// IsActive indicates that this handler is active and waits for new metric snapshots
	IsActive() bool
}

type snapshotKeyLabels string

// SnapshotKey identifies all the snapshots that were received from a specific device and exported with the specific name.
type SnapshotKey struct {
	name   string
	source PhysicalAddress
	labels snapshotKeyLabels
}

// Snapshot stores all information about a single metric snapshot.
type Snapshot struct {
	name        string
	source      PhysicalAddress
	destination GroupAddress
	value       float64
	timestamp   time.Time
	config      *GroupAddressConfig
}

type metricSnapshots struct {
	lock        sync.RWMutex
	snapshots   map[SnapshotKey]snapshot
	registerer  prometheus.Registerer
	metricsChan chan *Snapshot
	active      bool
}

type snapshot struct {
	snapshot *Snapshot
	metric   prometheus.Collector
}

func NewMetricsSnapshotHandler(registerer prometheus.Registerer) MetricSnapshotHandler {
	return &metricSnapshots{
		lock:        sync.RWMutex{},
		snapshots:   make(map[SnapshotKey]snapshot),
		registerer:  registerer,
		metricsChan: make(chan *Snapshot),
		active:      true,
	}
}

func (m *metricSnapshots) AddSnapshot(s *Snapshot) {
	key := s.GetKey()
	m.lock.Lock()
	defer m.lock.Unlock()
	meta, ok := m.snapshots[key]
	logger := slog.With(
		"metricName", s.name,
		"source", s.source,
	)

	if ok {
		meta.snapshot = s
	} else {
		metric, err := createMetric(s, m.GetValueFunc(key))
		if err != nil {
			logger.Warn(err.Error())
			return
		}
		meta = snapshot{snapshot: s, metric: metric}
		err = m.registerer.Register(meta.metric)
		if err != nil && !errors.Is(err, prometheus.AlreadyRegisteredError{}) {
			logger.Warn("Can not register new metric: " + err.Error())
		}
	}
	m.snapshots[key] = meta
}

func (m *metricSnapshots) FindSnapshot(key SnapshotKey) (*Snapshot, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	meta, ok := m.snapshots[key]
	if !ok {
		return nil, fmt.Errorf("no snapshot for %s from %s found", key.name, key.source.String())
	}
	return meta.snapshot, nil
}

func (m *metricSnapshots) FindYoungestSnapshot(name string) *Snapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var youngest *Snapshot
	for _, s := range m.snapshots {
		if s.snapshot.name != name {
			continue
		}
		if youngest == nil {
			youngest = s.snapshot
			continue
		}
		if youngest.timestamp.Before(s.snapshot.timestamp) {
			youngest = s.snapshot
		}
	}
	return youngest
}

func (m *metricSnapshots) GetValueFunc(key SnapshotKey) func() float64 {
	return func() float64 {
		s, err := m.FindSnapshot(key)
		if err != nil {
			return math.NaN()
		}
		return s.value
	}
}

func (m *metricSnapshots) Run() {
	defer func() { m.active = false }()
	for snap := range m.metricsChan {
		m.AddSnapshot(snap)
	}
}

func (m *metricSnapshots) IsActive() bool {
	return m.active
}

func (m *metricSnapshots) Close() {
	close(m.metricsChan)
}

func (m *metricSnapshots) GetMetricsChannel() chan *Snapshot {
	return m.metricsChan
}

func (s *Snapshot) GetKey() SnapshotKey {
	var labels []byte = nil
	if s.config.Labels != nil {
		var e error
		labels, e = json.Marshal(s.config.Labels)
		if e != nil {
			panic("can not create labels key")
		}
	}
	return SnapshotKey{
		name:   s.name,
		source: s.source,
		labels: snapshotKeyLabels(labels),
	}
}

func createMetric(s *Snapshot, getter func() float64) (prometheus.Collector, error) {
	var metric prometheus.Collector

	if strings.ToLower(s.config.MetricType) == "counter" {
		metric = prometheus.NewCounterFunc(
			prometheus.CounterOpts{
				Name:        s.name,
				ConstLabels: getSnapshotLabels(s),
			},
			getter,
		)
	} else if strings.ToLower(s.config.MetricType) == "gauge" {
		metric = prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name:        s.name,
				ConstLabels: getSnapshotLabels(s),
			},
			getter,
		)
	} else {
		return nil, fmt.Errorf("can not create metric '%s' for metric typ '%s'", s.name, s.config.MetricType)
	}
	return metric, nil
}

// getSnapshotLabels returns a full list of all labels that should be added to the given metric.
func getSnapshotLabels(s *Snapshot) map[string]string {
	var labels = map[string]string{
		"physicalAddress": s.source.String(),
	}
	if s.config.Labels != nil {
		for name, value := range s.config.Labels {
			labels[name] = value
		}
	}
	return labels
}
