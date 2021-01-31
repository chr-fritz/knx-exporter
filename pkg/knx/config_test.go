package knx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
