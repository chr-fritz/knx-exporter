package knx

import (
	"math"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/chr-fritz/knx-exporter/pkg/metrics/fake"
)

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
