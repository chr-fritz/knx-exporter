package knx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getMetricsToRead(t *testing.T) {

	tests := []struct {
		name   string
		config *Config
		want   GroupAddressConfigSet
	}{
		{
			"empty",
			&Config{AddressConfigs: GroupAddressConfigSet{}},
			GroupAddressConfigSet{},
		},
		{
			"single-no-export-no-startup-read",
			&Config{AddressConfigs: GroupAddressConfigSet{0: GroupAddressConfig{ReadStartup: false, Export: false}}},
			GroupAddressConfigSet{},
		},
		{
			"single-no-export-startup-read",
			&Config{AddressConfigs: GroupAddressConfigSet{0: GroupAddressConfig{Export: false, ReadStartup: true}}},
			GroupAddressConfigSet{},
		},
		{
			"single-export-no-startup-read",
			&Config{AddressConfigs: GroupAddressConfigSet{0: GroupAddressConfig{Export: true, ReadStartup: false}}},
			GroupAddressConfigSet{},
		},
		{
			"single-export-startup-read",
			&Config{AddressConfigs: GroupAddressConfigSet{0: GroupAddressConfig{Export: true, ReadStartup: true}}},
			GroupAddressConfigSet{0: GroupAddressConfig{ReadStartup: true}},
		},
		{
			"multiple-export-startup-read",
			&Config{AddressConfigs: GroupAddressConfigSet{
				0: GroupAddressConfig{Export: false, ReadStartup: false},
				1: GroupAddressConfig{Export: true, ReadStartup: false},
				2: GroupAddressConfig{Export: false, ReadStartup: true},
				3: GroupAddressConfig{Export: true, ReadStartup: true},
				4: GroupAddressConfig{Export: true, ReadStartup: true},
			}},
			GroupAddressConfigSet{
				3: GroupAddressConfig{ReadStartup: true},
				4: GroupAddressConfig{ReadStartup: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMetricsToRead(tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}
