// Copyright © 2020-2024 Christian Fritz <mail@chr-fritz.de>
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
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/vapourismo/knx-go/knx"
)

// Config defines the structure of the configuration file which defines which
// KNX Group Addresses were mapped into prometheus metrics.
type Config struct {
	Connection Connection `json:",omitempty"`
	// MetricsPrefix is a short prefix which will be added in front of the actual metric name.
	MetricsPrefix  string
	AddressConfigs GroupAddressConfigSet
	// ReadStartupInterval is the intervall to wait between read of group addresses after startup.
	ReadStartupInterval Duration `json:",omitempty"`
}

// ReadConfig reads the given configuration file and returns the parsed Config object.
func ReadConfig(configFile string) (*Config, error) {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("can not read group address configuration: %s", err)
	}
	config := Config{
		Connection: Connection{
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
	}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("can not read config file %s: %s", configFile, err)
	}
	return &config, nil
}

// NameForGa returns the full metric name for the given GroupAddress.
func (c *Config) NameForGa(address GroupAddress) string {
	gaConfig, ok := c.AddressConfigs[address]
	if !ok {
		return ""
	}
	return c.NameFor(gaConfig)
}

// NameFor return s the full metric name for the given GroupAddressConfig.
func (c *Config) NameFor(gaConfig *GroupAddressConfig) string {
	return c.MetricsPrefix + gaConfig.Name
}

// Connection contains the information about how to connect to the KNX system and how to identify itself.
type Connection struct {
	// Type of the actual connection. Can be either Tunnel or Router
	Type ConnectionType
	// Endpoint defines the IP address or hostname and port to where it should connect.
	Endpoint string
	// PhysicalAddress defines how the knx-exporter should identify itself within the KNX system.
	PhysicalAddress PhysicalAddress
	// RouterConfig contains some the specific configurations if connection Type is Router
	RouterConfig RouterConfig
	// TunnelConfig contains some the specific configurations if connection Type is Tunnel
	TunnelConfig TunnelConfig
}

type RouterConfig struct {
	// RetainCount specifies how many sent messages to retain. This is important for when a router indicates that it has
	// lost some messages. If you do not expect to saturate the router, keep this low.
	RetainCount uint
	// Interface specifies the network interface used to send and receive KNXnet/IP packets. If the interface is nil, the
	// system-assigned multicast interface is used.
	Interface string
	// MulticastLoopbackEnabled specifies if Multicast Loopback should be enabled.
	MulticastLoopbackEnabled bool
	// PostSendPauseDuration specifies the pause duration after sending. 0 means disabled. According to the specification,
	// we may choose to always pause for 20 ms after transmitting, but we should always pause for at least 5 ms on a
	// multicast address.
	PostSendPauseDuration time.Duration
}

func (rc RouterConfig) toKnxRouterConfig() (knx.RouterConfig, error) {
	var iface *net.Interface = nil
	if strings.Trim(rc.Interface, "\n\r\t ") != "" {
		var err error
		iface, err = net.InterfaceByName(rc.Interface)
		if err != nil {
			return knx.RouterConfig{}, err
		}
	}
	return knx.RouterConfig{
		RetainCount:              rc.RetainCount,
		Interface:                iface,
		MulticastLoopbackEnabled: rc.MulticastLoopbackEnabled,
		PostSendPauseDuration:    rc.PostSendPauseDuration,
	}, nil
}

type TunnelConfig struct {
	// ResendInterval is the interval with which requests will be resent if no response is received.
	ResendInterval time.Duration

	// HeartbeatInterval specifies the time interval which triggers a heartbeat check.
	HeartbeatInterval time.Duration

	// ResponseTimeout specifies how long to wait for a response.
	ResponseTimeout time.Duration

	// SendLocalAddress specifies if local address should be sent on connection request.
	SendLocalAddress bool

	// UseTCP configures whether to connect to the gateway using TCP.
	UseTCP bool
}

func (tc TunnelConfig) toKnxTunnelConfig() knx.TunnelConfig {
	return knx.TunnelConfig{
		ResendInterval:    tc.ResendInterval,
		HeartbeatInterval: tc.HeartbeatInterval,
		ResponseTimeout:   tc.ResponseTimeout,
		SendLocalAddress:  tc.SendLocalAddress,
		UseTCP:            tc.UseTCP,
	}
}

type ConnectionType string

const Tunnel = ConnectionType("Tunnel")
const Router = ConnectionType("Router")

func (t ConnectionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t *ConnectionType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch strings.ToLower(str) {
	case "tunnel":
		*t = Tunnel
	case "router":
		*t = Router
	default:
		return fmt.Errorf("invalid connection type given: \"%s\"", str)
	}
	return nil
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

type ReadType string

const GroupRead = ReadType("GroupRead")
const WriteOther = ReadType("WriteOther")

func (t ReadType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t *ReadType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch strings.ToLower(str) {
	case "groupread":
		*t = GroupRead
	case "writeother":
		*t = WriteOther
	default:
		*t = GroupRead
	}
	return nil
}

// GroupAddressConfig defines all information to map a KNX group address to a prometheus metric.
type GroupAddressConfig struct {
	// Name defines the prometheus metric name without the MetricsPrefix.
	Name string
	// Comment to identify the group address.
	Comment string `json:",omitempty"`
	// DPT defines the DPT at the knx bus. This is required to parse the values correctly.
	DPT string
	// MetricType is the type that prometheus uses when exporting it. i.e. gauge or counter
	MetricType string
	// Export the metric to prometheus
	Export bool
	// ReadStartup allows the exporter to actively send `GroupValueRead` telegrams to actively read the value at startup instead waiting for it.
	ReadStartup bool `json:",omitempty"`
	// ReadActive allows the exporter to actively send `GroupValueRead` telegrams to actively poll the value instead waiting for it.
	ReadActive bool `json:",omitempty"`
	// ReadType defines the type how to trigger the read request. Possible Values are GroupRead and WriteOther.
	ReadType ReadType `json:",omitempty"`
	// ReadAddress defines the group address to which address a GroupWrite request should be sent to initiate sending the data if ReadType is set to WriteOther.
	ReadAddress GroupAddress `json:",omitempty"`
	// ReadBody is a byte array with the content to sent to ReadAddress if ReadType is set to WriteOther.
	ReadBody []byte `json:",omitempty"`
	// MaxAge of a value until it will actively send a `GroupValueRead` telegram to read the value if ReadActive is set to true.
	MaxAge Duration `json:",omitempty"`
	// Labels defines static labels that should be set when exporting the metric using prometheus.
	Labels map[string]string `json:",omitempty"`
}

// GroupAddressConfigSet is a shortcut type for the group address config map.
type GroupAddressConfigSet map[GroupAddress]*GroupAddressConfig
