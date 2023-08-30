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
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vapourismo/knx-go/knx"
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
				RouterConfig: RouterConfig{
					RetainCount:              32,
					MulticastLoopbackEnabled: false,
					PostSendPauseDuration:    20 * time.Millisecond,
				},
				TunnelConfig: TunnelConfig{
					ResendInterval:    500 * time.Millisecond,
					HeartbeatInterval: 10 * time.Second,
					ResponseTimeout:   10 * time.Second,
					SendLocalAddress:  false,
					UseTCP:            false,
				},
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

func TestTunnelConfig_toKnxTunnelConfig(t *testing.T) {
	tests := []struct {
		name              string
		ResendInterval    time.Duration
		HeartbeatInterval time.Duration
		ResponseTimeout   time.Duration
		SendLocalAddress  bool
		UseTCP            bool
		want              knx.TunnelConfig
	}{
		{"default config",
			500 * time.Millisecond,
			10 * time.Second,
			10 * time.Second,
			false,
			false,
			knx.DefaultTunnelConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := TunnelConfig{
				ResendInterval:    tt.ResendInterval,
				HeartbeatInterval: tt.HeartbeatInterval,
				ResponseTimeout:   tt.ResponseTimeout,
				SendLocalAddress:  tt.SendLocalAddress,
				UseTCP:            tt.UseTCP,
			}
			assert.Equalf(t, tt.want, tc.toKnxTunnelConfig(), "toKnxTunnelConfig()")
		})
	}
}

func TestRouterConfig_toKnxRouterConfig(t *testing.T) {
	iface, err := net.InterfaceByIndex(1)
	assert.NoError(t, err)
	tests := []struct {
		name                     string
		RetainCount              uint
		Interface                string
		MulticastLoopbackEnabled bool
		PostSendPauseDuration    time.Duration
		want                     knx.RouterConfig
		wantErr                  assert.ErrorAssertionFunc
	}{
		{
			"default config",
			32,
			"",
			false,
			20 * time.Millisecond,
			knx.DefaultRouterConfig,
			assert.NoError,
		},
		{
			"wrong interface",
			32,
			"non existing interface",
			false,
			20 * time.Millisecond,
			knx.RouterConfig{},
			func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NotNil(t, err, i)
			},
		},
		{
			"local interface",
			32,
			iface.Name,
			false,
			20 * time.Millisecond,
			knx.RouterConfig{
				RetainCount:           32,
				Interface:             iface,
				PostSendPauseDuration: 20 * time.Millisecond,
			},
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := RouterConfig{
				RetainCount:              tt.RetainCount,
				Interface:                tt.Interface,
				MulticastLoopbackEnabled: tt.MulticastLoopbackEnabled,
				PostSendPauseDuration:    tt.PostSendPauseDuration,
			}
			got, err := rc.toKnxRouterConfig()
			if !tt.wantErr(t, err, "toKnxRouterConfig()") {
				return
			}
			assert.Equalf(t, tt.want, got, "toKnxRouterConfig()")
		})
	}
}
