package knx

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
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
}

// SnapshotKey identifies all the snapshots that were received from a specific device and exported with the specific name.
type SnapshotKey struct {
	name   string
	source PhysicalAddress
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
	lock       sync.RWMutex
	snapshots  map[SnapshotKey]snapshot
	registerer prometheus.Registerer
}

type snapshot struct {
	snapshot *Snapshot
	metric   prometheus.Collector
}

func NewMetricsSnapshotHandler(registerer prometheus.Registerer) MetricSnapshotHandler {
	return &metricSnapshots{
		lock:       sync.RWMutex{},
		snapshots:  make(map[SnapshotKey]snapshot),
		registerer: registerer,
	}
}

func (m *metricSnapshots) AddSnapshot(s *Snapshot) {
	key := s.GetKey()
	m.lock.Lock()
	defer m.lock.Unlock()
	meta, ok := m.snapshots[key]
	if ok {
		meta.snapshot = s
	} else {
		meta = snapshot{
			snapshot: s,
			metric:   createMetric(s, m.GetValueFunc(key)),
		}
		err := m.registerer.Register(meta.metric)
		if err != nil && !errors.Is(err, prometheus.AlreadyRegisteredError{}) {
			logrus.Warnf("Can not register new metric %s from %s: %s", s.name, s.source.String(), err)
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

func (s *Snapshot) GetKey() SnapshotKey {
	return SnapshotKey{
		name:   s.name,
		source: s.source,
	}
}

func createMetric(s *Snapshot, getter func() float64) prometheus.Collector {
	var metric prometheus.Collector
	if strings.ToLower(s.config.MetricType) == "counter" {
		metric = prometheus.NewCounterFunc(
			prometheus.CounterOpts{
				Name:        s.name,
				Help:        fmt.Sprintf("Value of %s\n%s", s.destination.String(), s.config.Comment),
				ConstLabels: map[string]string{"physicalAddress": s.source.String()},
			},
			getter,
		)
	} else if strings.ToLower(s.config.MetricType) == "gauge" {
		metric = prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name:        s.name,
				Help:        fmt.Sprintf("Value of %s\n%s", s.destination.String(), s.config.Comment),
				ConstLabels: map[string]string{"physicalAddress": s.source.String()},
			},
			getter,
		)
	}
	return metric
}
