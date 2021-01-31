package knx

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Config defines the structure of the configuration file which defines which
// KNX Group Addresses were mapped into prometheus metrics.
type Config struct {
	Connection Connection `json:",omitempty"`
	// MetricsPrefix is a short prefix which will be added in front of the actual metric name.
	MetricsPrefix  string
	AddressConfigs GroupAddressConfigSet
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
func (c *Config) NameFor(gaConfig GroupAddressConfig) string {
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
	// ReadActive allows the exporter to actively send `GroupValueRead` telegrams to actively poll the value instead waiting for it.
	ReadActive bool `json:",omitempty"`
	// MaxAge of a value until it will actively send a `GroupValueRead` telegram to read the value if ReadActive is set to true.
	MaxAge Duration `json:",omitempty"`
}

// GroupAddressConfigSet is a shortcut type for the group address config map.
type GroupAddressConfigSet map[GroupAddress]GroupAddressConfig
