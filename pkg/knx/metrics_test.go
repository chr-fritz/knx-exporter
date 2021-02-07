package knx

import (
	"math"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"

	"github.com/chr-fritz/knx-exporter/pkg/metrics/fake"
)

func TestMetricsExporter_handleEvent(t *testing.T) {
	tests := []struct {
		name      string
		event     knx.GroupEvent
		want      *Snapshot
		wantError bool
	}{
		{
			"bool false",
			knx.GroupEvent{Destination: cemi.GroupAddr(1), Data: []byte{0}},
			&Snapshot{name: "knx_a", value: 0, destination: GroupAddress(1), config: &GroupAddressConfig{Name: "a", DPT: "1.001", Export: true}},
			false,
		},
		{
			"bool true",
			knx.GroupEvent{Destination: cemi.GroupAddr(1), Data: []byte{1}},
			&Snapshot{name: "knx_a", value: 1, destination: GroupAddress(1), config: &GroupAddressConfig{Name: "a", DPT: "1.001", Export: true}},
			false,
		},
		{
			"5.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(2), Data: []byte{0, 255}},
			&Snapshot{name: "knx_b", value: 100, destination: GroupAddress(2), config: &GroupAddressConfig{Name: "b", DPT: "5.001", Export: true}},
			false,
		},
		{
			"9.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(3), Data: []byte{0, 2, 38}},
			&Snapshot{name: "knx_c", value: 5.5, destination: GroupAddress(3), config: &GroupAddressConfig{Name: "c", DPT: "9.001", Export: true}},
			false,
		},
		{
			"12.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(4), Data: []byte{0, 0, 0, 0, 5}},
			&Snapshot{name: "knx_d", value: 5, destination: GroupAddress(4), config: &GroupAddressConfig{Name: "d", DPT: "12.001", Export: true}},
			false,
		},
		{
			"13.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(5), Data: []byte{0, 0, 0, 0, 5}},
			&Snapshot{name: "knx_e", value: 5, destination: GroupAddress(5), config: &GroupAddressConfig{Name: "e", DPT: "13.001", Export: true}},
			false,
		},
		{
			"14.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(6), Data: []byte{0, 63, 192, 0, 0}},
			&Snapshot{name: "knx_f", value: 1.5, destination: GroupAddress(6), config: &GroupAddressConfig{Name: "f", DPT: "14.001", Export: true}},
			false,
		},
		{
			"5.* can't unpack",
			knx.GroupEvent{Destination: cemi.GroupAddr(2), Data: []byte{0}},
			nil,
			true,
		},
		{
			"bool unexported",
			knx.GroupEvent{Destination: cemi.GroupAddr(7), Data: []byte{1}},
			nil,
			true,
		},
		{
			"unknown address",
			knx.GroupEvent{Destination: cemi.GroupAddr(255), Data: []byte{0}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &MetricsExporter{
				config: &Config{
					MetricsPrefix: "knx_",
					AddressConfigs: map[GroupAddress]GroupAddressConfig{
						GroupAddress(1): {Name: "a", DPT: "1.001", Export: true},
						GroupAddress(2): {Name: "b", DPT: "5.001", Export: true},
						GroupAddress(3): {Name: "c", DPT: "9.001", Export: true},
						GroupAddress(4): {Name: "d", DPT: "12.001", Export: true},
						GroupAddress(5): {Name: "e", DPT: "13.001", Export: true},
						GroupAddress(6): {Name: "f", DPT: "14.001", Export: true},
						GroupAddress(7): {Export: false},
					},
				},
				metricsChan:    make(chan *Snapshot, 1),
				metrics:        NewMetricsSnapshotHandler(nil),
				messageCounter: prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"direction", "processed"}),
			}
			e.handleEvent(tt.event)
			select {
			case got := <-e.metricsChan:
				// ignore timestamps
				got.timestamp = time.Unix(0, 0)
				tt.want.timestamp = time.Unix(0, 0)
				assert.Equal(t, tt.want, got)
			case <-time.After(10 * time.Millisecond):
				assert.True(t, tt.wantError, "got no metrics snapshot but requires one")
			}
		})
	}
}

func TestMetricsExporter_getMetricsValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		snapshot *Snapshot
		metric   string
		want     float64
	}{
		{"success", &Snapshot{name: "a", value: 1.5, config: &GroupAddressConfig{MetricType: "counter"}}, "a", 1.5},
		{"missing", &Snapshot{name: "a", value: 1.5, config: &GroupAddressConfig{MetricType: "gauge"}}, "b", math.NaN()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := fake.NewMockExporter(ctrl)
			e := &MetricsExporter{
				metricsChan: make(chan *Snapshot),
				metrics:     NewMetricsSnapshotHandler(exporter),
			}
			exporter.EXPECT().Register(gomock.Any()).AnyTimes()
			go e.storeSnapshots()
			e.metricsChan <- tt.snapshot
			value := e.metrics.GetValueFunc(SnapshotKey{name: tt.metric})()
			if !math.IsNaN(tt.want) {
				assert.Equal(t, tt.want, value)
			} else {
				assert.True(t, math.IsNaN(value))
			}
			close(e.metricsChan)
		})
	}
}
