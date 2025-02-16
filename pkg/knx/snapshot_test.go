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

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestSnapshot_getKey(t *testing.T) {
	tests := []struct {
		name     string
		snapshot *Snapshot
		want     SnapshotKey
	}{
		{"ok", &Snapshot{name: "metricName", source: 1, config: &GroupAddressConfig{Labels: nil}, destination: 1}, SnapshotKey{target: 1, source: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.snapshot.getKey()
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_metricSnapshots_AddSnapshot(t *testing.T) {

	tests := []struct {
		name         string
		s            *Snapshot
		key          SnapshotKey
		wantRegister int
	}{
		{"new counter", &Snapshot{name: "a", source: 1, destination: 1, config: &GroupAddressConfig{MetricType: "counter"}}, SnapshotKey{target: 1, source: 1}, 1},
		{"new gauge", &Snapshot{name: "b", source: 1, destination: 2, config: &GroupAddressConfig{MetricType: "gauge", Labels: map[string]string{"room": "office"}}}, SnapshotKey{target: 2, source: 1}, 1},
		{"update", &Snapshot{name: "c", source: 1, destination: 3, config: &GroupAddressConfig{MetricType: "gauge"}}, SnapshotKey{target: 3, source: 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := NewMetricsSnapshotHandler()
			snapshots := handler.(*metricSnapshots)
			snapshots.descriptions[SnapshotKey{source: 1, target: 3}] = prometheus.NewDesc("", "", []string{}, map[string]string{})
			handler.AddSnapshot(tt.s)
			assert.NotNil(t, snapshots.descriptions[tt.key])
			assert.NotNil(t, snapshots.snapshots[tt.key])
		})
	}
}

func Test_metricSnapshots_FindSnapshot(t *testing.T) {
	tests := []struct {
		name              string
		existingSnapshots map[SnapshotKey]*Snapshot
		key               SnapshotKey
		want              *Snapshot
		wantErr           bool
	}{
		{
			"found",
			map[SnapshotKey]*Snapshot{SnapshotKey{source: 1, target: 2}: {name: "found"}},
			SnapshotKey{source: 1, target: 2},
			&Snapshot{name: "found"},
			false},
		{
			"found two devs",
			map[SnapshotKey]*Snapshot{
				SnapshotKey{source: 1, target: 1}: {name: "found", source: 1},
				SnapshotKey{source: 2, target: 1}: {name: "found", source: 2},
			},
			SnapshotKey{source: 1, target: 1},
			&Snapshot{name: "found", source: 1},
			false},
		{
			"found one dev",
			map[SnapshotKey]*Snapshot{
				SnapshotKey{source: 1, target: 1}: {name: "found1", source: 1},
				SnapshotKey{source: 1, target: 2}: {name: "found2", source: 1},
			},
			SnapshotKey{source: 1, target: 1},
			&Snapshot{name: "found1", source: 1},
			false},
		{
			"no snapshots",
			map[SnapshotKey]*Snapshot{},
			SnapshotKey{source: 2, target: 1},
			nil,
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metricSnapshots{
				lock:      sync.RWMutex{},
				snapshots: tt.existingSnapshots,
			}
			got, err := m.FindSnapshot(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_metricSnapshots_FindYoungestSnapshot(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name              string
		existingSnapshots map[SnapshotKey]*Snapshot
		metricName        string
		want              *Snapshot
	}{
		{
			"no snapshots",
			map[SnapshotKey]*Snapshot{},
			"metric",
			nil,
		},
		{
			"one snapshot",
			map[SnapshotKey]*Snapshot{SnapshotKey{source: 1, target: 1}: {name: "a", source: 1}},
			"a",
			&Snapshot{name: "a", source: 1},
		},
		{
			"two dev",
			map[SnapshotKey]*Snapshot{
				SnapshotKey{source: 1, target: 1}: {name: "a", source: 1, timestamp: testTime},
				SnapshotKey{source: 2, target: 1}: {name: "a", source: 2, timestamp: testTime.Add(-10 * time.Second)},
				SnapshotKey{source: 3, target: 1}: {name: "a", source: 3, timestamp: testTime.Add(-20 * time.Second)},
			},
			"a",
			&Snapshot{name: "a", source: 1, timestamp: testTime},
		},
		{
			"two metrics",
			map[SnapshotKey]*Snapshot{
				SnapshotKey{source: 1, target: 1}: {name: "a", source: 1},
				SnapshotKey{source: 1, target: 2}: {name: "b", source: 1},
			},
			"a",
			&Snapshot{name: "a", source: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metricSnapshots{
				lock:      sync.RWMutex{},
				snapshots: tt.existingSnapshots,
			}
			got := m.FindYoungestSnapshot(tt.metricName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_metricSnapshots_Describe(t *testing.T) {
	tests := []struct {
		name         string
		snapshots    []*Snapshot
		expectedDesc []*prometheus.Desc
	}{
		{"no snapshots", []*Snapshot{}, []*prometheus.Desc{}},
		{"single snapshots",
			[]*Snapshot{{name: "dummy", value: 1, source: 1, config: &GroupAddressConfig{Comment: "abc"}}},
			[]*prometheus.Desc{
				prometheus.NewDesc("dummy", "abc", []string{}, map[string]string{"physicalAddress": "0.0.1"}),
			},
		},
		{
			"two different snapshots",
			[]*Snapshot{
				{name: "dummy", value: 1, source: 1, config: &GroupAddressConfig{}},
				{name: "dummy1", value: 2, source: 2, config: &GroupAddressConfig{Labels: map[string]string{"room": "outside"}}},
			},
			[]*prometheus.Desc{
				prometheus.NewDesc("dummy", "", []string{}, map[string]string{"physicalAddress": "0.0.1"}),
				prometheus.NewDesc("dummy1", "", []string{}, map[string]string{"physicalAddress": "0.0.2", "room": "outside"}),
			},
		},
		{
			"duplicate snapshots",
			[]*Snapshot{
				{name: "dummy", value: 1, source: 1, config: &GroupAddressConfig{}},
				{name: "dummy", value: 2, source: 1, config: &GroupAddressConfig{}},
			},
			[]*prometheus.Desc{
				prometheus.NewDesc("dummy", "", []string{}, map[string]string{"physicalAddress": "0.0.1"}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMetricsSnapshotHandler()
			for _, snapshot := range tt.snapshots {
				handler.AddSnapshot(snapshot)
			}

			ch := make(chan *prometheus.Desc)
			actualDesc := make([]*prometheus.Desc, 0)
			go func() {
				for desc := range ch {
					actualDesc = append(actualDesc, desc)
				}
			}()

			handler.Describe(ch)
			time.Sleep(100 * time.Millisecond)
			close(ch)
			assert.Equal(t, tt.expectedDesc, actualDesc)
		})
	}
}

func Test_metricSnapshots_Collect(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name      string
		snapshots []*Snapshot
		metrics   []prometheus.Metric
	}{
		{"no metrics", []*Snapshot{}, []prometheus.Metric{}},
		{"single counter metric",
			[]*Snapshot{
				{name: "dummy", value: 1, source: 1, timestamp: testTime, config: &GroupAddressConfig{MetricType: "counter"}},
			},
			[]prometheus.Metric{
				prometheus.MustNewConstMetric(prometheus.NewDesc("dummy", "", []string{}, map[string]string{"physicalAddress": "0.0.1"}), prometheus.CounterValue, 1),
			},
		},
		{"single gauge timestamp metric",
			[]*Snapshot{
				{name: "dummy", value: 1, source: 1, timestamp: testTime, config: &GroupAddressConfig{MetricType: "gauge", WithTimestamp: true}},
			},
			[]prometheus.Metric{
				prometheus.NewMetricWithTimestamp(testTime, prometheus.MustNewConstMetric(prometheus.NewDesc("dummy", "", []string{}, map[string]string{"physicalAddress": "0.0.1"}), prometheus.GaugeValue, 1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMetricsSnapshotHandler()
			for _, snapshot := range tt.snapshots {
				handler.AddSnapshot(snapshot)
			}

			ch := make(chan prometheus.Metric)
			actualMetrics := make([]prometheus.Metric, 0)
			go func() {
				for desc := range ch {
					actualMetrics = append(actualMetrics, desc)
				}
			}()

			handler.Collect(ch)
			time.Sleep(100 * time.Millisecond)
			close(ch)
			assert.Equal(t, tt.metrics, actualMetrics)
		})
	}
}
