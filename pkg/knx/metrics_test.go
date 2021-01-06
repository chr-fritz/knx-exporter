package knx

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
)

func Test_exporter_readConfig(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		config     *Config
		wantErr    bool
	}{
		{"wrong filename", "fixtures/invalid.yaml", nil, true},
		{"full config", "fixtures/full-config.yaml", &Config{
			Connection: Connection{
				Type:     Tunnel,
				Endpoint: "192.168.1.15:3671",
			},
			MetricsPrefix: "knx_",
			AddressConfigs: map[GroupAddress]GroupAddressConfig{
				1: {
					Name:       "dummy_metric",
					DPT:        "1.*",
					MetricType: "counter",
					Export:     true,
					ReadActive: true,
					MaxAge:     Duration(10 * time.Minute),
				},
			},
		}, false},
		{"converted config", "fixtures/ga-config.yaml", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := NewMetricsExporter(tt.configFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if e != nil {
				assert.Equal(t, tt.config, e.config)
			}
		})
	}
}

func TestMetricsExporter_handleEvent(t *testing.T) {
	tests := []struct {
		name      string
		event     knx.GroupEvent
		want      *metricSnapshot
		wantError bool
	}{
		{
			"bool false",
			knx.GroupEvent{Destination: cemi.GroupAddr(1), Data: []byte{0}},
			&metricSnapshot{name: "knx_a", value: 0},
			false,
		},
		{
			"bool true",
			knx.GroupEvent{Destination: cemi.GroupAddr(1), Data: []byte{1}},
			&metricSnapshot{name: "knx_a", value: 1},
			false,
		},
		{
			"5.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(2), Data: []byte{0, 255}},
			&metricSnapshot{name: "knx_b", value: 100},
			false,
		},
		{
			"9.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(3), Data: []byte{0, 2, 38}},
			&metricSnapshot{name: "knx_c", value: 5.5},
			false,
		},
		{
			"12.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(4), Data: []byte{0, 0, 0, 0, 5}},
			&metricSnapshot{name: "knx_d", value: 5},
			false,
		},
		{
			"13.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(5), Data: []byte{0, 0, 0, 0, 5}},
			&metricSnapshot{name: "knx_e", value: 5},
			false,
		},
		{
			"14.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(6), Data: []byte{0, 63, 192, 0, 0}},
			&metricSnapshot{name: "knx_f", value: 1.5},
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
				metricsChan:    make(chan metricSnapshot, 1),
				snapshotLock:   sync.RWMutex{},
				metrics:        map[string]metricSnapshot{},
				messageCounter: prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"direction", "processed"}),
			}
			e.handleEvent(tt.event)
			select {
			case got := <-e.metricsChan:
				// ignore timestamps
				got.timestamp = time.Unix(0, 0)
				tt.want.timestamp = time.Unix(0, 0)
				assert.Equal(t, *tt.want, got)
			case <-time.After(10 * time.Millisecond):
				assert.True(t, tt.wantError, "got no metrics snapshot but requires one")
			}
		})
	}
}

func TestMetricsExporter_getMetricsValue(t *testing.T) {
	tests := []struct {
		name     string
		snapshot metricSnapshot
		metric   string
		want     float64
	}{
		{"success", metricSnapshot{name: "a", value: 1.5}, "a", 1.5},
		{"missing", metricSnapshot{name: "a", value: 1.5}, "b", math.NaN()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &MetricsExporter{
				metricsChan:  make(chan metricSnapshot),
				snapshotLock: sync.RWMutex{},
				metrics:      map[string]metricSnapshot{},
			}
			go e.storeSnapshots()
			e.metricsChan <- tt.snapshot
			value := e.getMetricsValue(tt.metric)()
			if !math.IsNaN(tt.want) {
				assert.Equal(t, tt.want, value)
			} else {
				assert.True(t, math.IsNaN(value))
			}
			close(e.metricsChan)
		})
	}
}
