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

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/chr-fritz/knx-exporter/pkg/metrics/fake"
)

func TestSnapshot_GetKey(t *testing.T) {
	tests := []struct {
		name     string
		snapshot *Snapshot
		want     SnapshotKey
	}{
		{"ok", &Snapshot{name: "metricName", source: 1, config: &GroupAddressConfig{Labels: nil}}, SnapshotKey{name: "metricName", source: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.snapshot.GetKey()
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
		{"new counter", &Snapshot{name: "a", source: 1, config: &GroupAddressConfig{MetricType: "counter"}}, SnapshotKey{name: "a", source: 1}, 1},
		{"new gauge", &Snapshot{name: "b", source: 1, config: &GroupAddressConfig{MetricType: "gauge", Labels: map[string]string{"room": "office"}}}, SnapshotKey{name: "b", source: 1, labels: snapshotKeyLabels("{\"room\":\"office\"}")}, 1},
		{"update", &Snapshot{name: "c", source: 1, config: &GroupAddressConfig{MetricType: "gauge"}}, SnapshotKey{name: "c", source: 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			exporter := fake.NewMockExporter(ctrl)
			exporter.EXPECT().Register(gomock.Any()).Times(tt.wantRegister)
			handler := NewMetricsSnapshotHandler(exporter)
			snapshots := handler.(*metricSnapshots)
			snapshots.snapshots[SnapshotKey{name: "c", source: 1}] = snapshot{metric: prometheus.NewCounter(prometheus.CounterOpts{})}
			handler.AddSnapshot(tt.s)
			assert.NotNil(t, snapshots.snapshots[tt.key])
			assert.NotNil(t, snapshots.snapshots[tt.key].snapshot)
			assert.NotNil(t, snapshots.snapshots[tt.key].metric)
		})
	}
}

func Test_metricSnapshots_FindSnapshot(t *testing.T) {
	tests := []struct {
		name              string
		existingSnapshots map[SnapshotKey]snapshot
		key               SnapshotKey
		want              *Snapshot
		wantErr           bool
	}{
		{
			"found",
			map[SnapshotKey]snapshot{SnapshotKey{name: "found"}: {snapshot: &Snapshot{name: "found"}}},
			SnapshotKey{name: "found"},
			&Snapshot{name: "found"},
			false},
		{
			"found two devs",
			map[SnapshotKey]snapshot{
				SnapshotKey{name: "found", source: 1}: {snapshot: &Snapshot{name: "found", source: 1}},
				SnapshotKey{name: "found", source: 2}: {snapshot: &Snapshot{name: "found", source: 2}},
			},
			SnapshotKey{name: "found", source: 1},
			&Snapshot{name: "found", source: 1},
			false},
		{
			"found one dev",
			map[SnapshotKey]snapshot{
				SnapshotKey{name: "found1", source: 1}: {snapshot: &Snapshot{name: "found1", source: 1}},
				SnapshotKey{name: "found2", source: 1}: {snapshot: &Snapshot{name: "found2", source: 1}},
			},
			SnapshotKey{name: "found1", source: 1},
			&Snapshot{name: "found1", source: 1},
			false},
		{
			"no snapshots",
			map[SnapshotKey]snapshot{},
			SnapshotKey{name: "wanted"},
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
		existingSnapshots map[SnapshotKey]snapshot
		metricName        string
		want              *Snapshot
	}{
		{
			"no snapshots",
			map[SnapshotKey]snapshot{},
			"metric",
			nil,
		},
		{
			"one snapshot",
			map[SnapshotKey]snapshot{SnapshotKey{name: "a", source: 1}: {snapshot: &Snapshot{name: "a", source: 1}}},
			"a",
			&Snapshot{name: "a", source: 1},
		},
		{
			"two dev",
			map[SnapshotKey]snapshot{
				SnapshotKey{name: "a", source: 1}: {snapshot: &Snapshot{name: "a", source: 1, timestamp: testTime}},
				SnapshotKey{name: "a", source: 2}: {snapshot: &Snapshot{name: "a", source: 2, timestamp: testTime.Add(-10 * time.Second)}},
				SnapshotKey{name: "a", source: 3}: {snapshot: &Snapshot{name: "a", source: 3, timestamp: testTime.Add(-20 * time.Second)}},
			},
			"a",
			&Snapshot{name: "a", source: 1, timestamp: testTime},
		},
		{
			"two metrics",
			map[SnapshotKey]snapshot{
				SnapshotKey{name: "a", source: 1}: {snapshot: &Snapshot{name: "a", source: 1}},
				SnapshotKey{name: "b", source: 1}: {snapshot: &Snapshot{name: "b", source: 1}},
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

func Test_metricSnapshots_GetValueFunc(t *testing.T) {
	tests := []struct {
		name      string
		snapshots map[SnapshotKey]snapshot
		key       SnapshotKey
		want      float64
	}{
		{
			"ok",
			map[SnapshotKey]snapshot{
				SnapshotKey{name: "a", source: 1}: {snapshot: &Snapshot{name: "a", source: 1, value: 1.5}},
			},
			SnapshotKey{name: "a", source: 1},
			1.5,
		},
		{
			"not found",
			map[SnapshotKey]snapshot{
				SnapshotKey{name: "a", source: 1}: {snapshot: &Snapshot{name: "a", source: 1, value: 1.5}},
			},
			SnapshotKey{name: "b", source: 1},
			math.NaN(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metricSnapshots{
				lock:      sync.RWMutex{},
				snapshots: tt.snapshots,
			}
			value := m.GetValueFunc(tt.key)()
			if !math.IsNaN(tt.want) {
				assert.Equal(t, tt.want, value)
			} else {
				assert.True(t, math.IsNaN(value))
			}
		})
	}
}
