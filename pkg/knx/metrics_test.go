package knx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
