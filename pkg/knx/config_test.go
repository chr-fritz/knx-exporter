package knx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vapourismo/knx-go/knx/cemi"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		config     *Config
		wantErr    bool
	}{
		{"wrong filename", "fixtures/invalid.yaml", nil, true},
		{"full config", "fixtures/full-config.yaml", &Config{
			Connection: Connection{
				Type:            Tunnel,
				Endpoint:        "192.168.1.15:3671",
				PhysicalAddress: PhysicalAddress(cemi.NewIndividualAddr3(2, 0, 1)),
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
			config, err := ReadConfig(tt.configFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.config, config)
		})
	}
}

func TestConfig_NameForGa(t *testing.T) {
	tests := []struct {
		name           string
		address        GroupAddress
		MetricsPrefix  string
		AddressConfigs GroupAddressConfigSet
		want           string
	}{
		{"not found", GroupAddress(1), "knx_", GroupAddressConfigSet{}, ""},
		{"ok", GroupAddress(1), "knx_", GroupAddressConfigSet{1: {Name: "dummy"}}, "knx_dummy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				MetricsPrefix:  tt.MetricsPrefix,
				AddressConfigs: tt.AddressConfigs,
			}
			got := c.NameForGa(tt.address)
			assert.Equal(t, tt.want, got)
		})
	}
}
