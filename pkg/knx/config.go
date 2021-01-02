package knx

import (
	"time"
)

// Config defines the structure of the configuration file which defines which
// KNX Group Addresses were mapped into prometheus metrics.
type Config struct {
	// MetricsPrefix is a short prefix which will be added in front of the actual metric name.
	MetricsPrefix  string
	AddressConfigs map[GroupAddress]GroupAddressConfig
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
	MaxAge time.Duration `json:",omitempty"`
}
