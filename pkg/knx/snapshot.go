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

//go:generate mockgen -destination=snapshotMocks_test.go -package=knx -self_package=github.com/chr-fritz/knx-exporter/pkg/knx -source=snapshot.go

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricSnapshotHandler holds and manages all the snapshots of metrics.
type MetricSnapshotHandler interface {
	prometheus.Collector

	// AddSnapshot adds a new snapshot that should be exported as metric.
	AddSnapshot(snapshot *Snapshot)
	// FindSnapshot finds a given snapshot by the snapshots key.
	FindSnapshot(key SnapshotKey) (*Snapshot, error)
	// FindYoungestSnapshot finds the youngest snapshot with the given metric name.
	// It don't matter from which device the snapshot was received.
	FindYoungestSnapshot(name string) *Snapshot
	// Run let the MetricSnapshotHandler listen for new snapshots on the Snapshot channel.
	Run(ctx context.Context)
	// GetMetricsChannel returns the channel to send new snapshots to this MetricSnapshotHandler.
	GetMetricsChannel() chan *Snapshot
	// IsActive indicates that this handler is active and waits for new metric snapshots
	IsActive() bool
}

// SnapshotKey identifies all the snapshots that were received from a specific device and exported with the specific name.
type SnapshotKey struct {
	source PhysicalAddress
	target GroupAddress
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
	lock         sync.RWMutex
	snapshots    map[SnapshotKey]*Snapshot
	descriptions map[SnapshotKey]*prometheus.Desc
	metricsChan  chan *Snapshot
	active       bool
}

func NewMetricsSnapshotHandler() MetricSnapshotHandler {
	return &metricSnapshots{
		lock:         sync.RWMutex{},
		snapshots:    make(map[SnapshotKey]*Snapshot),
		descriptions: make(map[SnapshotKey]*prometheus.Desc),
		metricsChan:  make(chan *Snapshot),
		active:       true,
	}
}

func (m *metricSnapshots) AddSnapshot(s *Snapshot) {
	key := s.getKey()
	m.lock.Lock()
	defer m.lock.Unlock()
	_, ok := m.descriptions[key]

	if !ok {
		m.descriptions[key] = createMetric(s)
	}
	m.snapshots[key] = s
}

func (m *metricSnapshots) FindSnapshot(key SnapshotKey) (*Snapshot, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	snapshot, ok := m.snapshots[key]
	if !ok {
		return nil, fmt.Errorf("no snapshot for %s from %s found", key.target.String(), key.source.String())
	}
	return snapshot, nil
}

func (m *metricSnapshots) FindYoungestSnapshot(name string) *Snapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var youngest *Snapshot
	for _, s := range m.snapshots {
		if s.name != name {
			continue
		}
		if youngest == nil {
			youngest = s
			continue
		}
		if youngest.timestamp.Before(s.timestamp) {
			youngest = s
		}
	}
	return youngest
}

func (m *metricSnapshots) Run(ctx context.Context) {
	m.active = true
	defer func() { m.active = false }()
loop:
	for {
		select {
		case snap := <-m.metricsChan:
			m.AddSnapshot(snap)
		case <-ctx.Done():
			break loop
		}
	}
}

func (m *metricSnapshots) IsActive() bool {
	return m.active
}

func (m *metricSnapshots) GetMetricsChannel() chan *Snapshot {
	return m.metricsChan
}

func (m *metricSnapshots) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range m.descriptions {
		ch <- d
	}
}

func (m *metricSnapshots) Collect(metrics chan<- prometheus.Metric) {
	for k, s := range m.snapshots {
		if s.config.WithTimestamp {
			metrics <- prometheus.NewMetricWithTimestamp(s.timestamp, prometheus.MustNewConstMetric(m.descriptions[k], s.getValuetype(), s.value))
		} else {
			metrics <- prometheus.MustNewConstMetric(m.descriptions[k], s.getValuetype(), s.value)
		}
	}
}

func (s *Snapshot) getKey() SnapshotKey {
	return SnapshotKey{
		source: s.source,
		target: s.destination,
	}
}

func (s *Snapshot) getValuetype() prometheus.ValueType {
	if strings.ToLower(s.config.MetricType) == "counter" {
		return prometheus.CounterValue
	} else if strings.ToLower(s.config.MetricType) == "gauge" {
		return prometheus.GaugeValue
	}
	return prometheus.UntypedValue
}

func createMetric(s *Snapshot) *prometheus.Desc {
	return prometheus.NewDesc(s.name, s.config.Comment, []string{}, getSnapshotLabels(s))
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
