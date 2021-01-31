package knx

import (
	"math"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
)

// Poller defines the interface for active polling for metrics values against the knx system.
type Poller interface {
	// Run starts the polling.
	Run()
	// Stop stops the polling.
	Stop()
}

type poller struct {
	exporter        *MetricsExporter
	pollingInterval time.Duration
	metricsToPoll   GroupAddressConfigSet
	ticker          *time.Ticker
}

// NewPoller creates a new Poller instance using the given MetricsExporter for connection handling and metrics observing.
func NewPoller(exporter *MetricsExporter) Poller {
	metricsToPoll := getMetricsToPoll(exporter.config)
	interval := calcPollingInterval(metricsToPoll)
	return &poller{
		exporter:        exporter,
		pollingInterval: interval,
		metricsToPoll:   metricsToPoll,
	}
}

func (p *poller) Run() {
	if p.pollingInterval <= 0 {
		return
	}
	p.ticker = time.NewTicker(p.pollingInterval)
	c := p.ticker.C
	go func() {
		for t := range c {
			p.pollAddresses(t)
		}
	}()
}

func (p *poller) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}

func (p *poller) pollAddresses(t time.Time) {
	for address, config := range p.metricsToPoll {
		snapshot := p.exporter.getMetricSnapshot(config.Name)
		if snapshot == nil || t.Sub(snapshot.timestamp).Truncate(time.Second) >= time.Duration(config.MaxAge) {
			p.sendReadMessage(address)
		}
	}
}

func (p *poller) sendReadMessage(address GroupAddress) {
	event := knx.GroupEvent{
		Command:     knx.GroupRead,
		Destination: cemi.GroupAddr(address),
		Source:      cemi.IndividualAddr(p.exporter.config.Connection.PhysicalAddress),
	}

	if e := p.exporter.client.Send(event); e != nil {
		logrus.Infof("can not send read request for %s: %s", address.String(), e)
	}
}

func getMetricsToPoll(config *Config) GroupAddressConfigSet {
	toPoll := make(GroupAddressConfigSet)
	for address, addressConfig := range config.AddressConfigs {
		interval := time.Duration(addressConfig.MaxAge).Truncate(time.Second)
		if !addressConfig.Export || !addressConfig.ReadActive || interval < time.Second {
			continue
		}

		interval = time.Duration(math.Max(float64(interval), float64(5*time.Second)))
		toPoll[address] = GroupAddressConfig{
			Name:       config.NameFor(addressConfig),
			ReadActive: true,
			MaxAge:     Duration(interval),
		}
	}
	return toPoll
}

func calcPollingInterval(config GroupAddressConfigSet) time.Duration {
	var intervals []time.Duration
	for _, ga := range config {
		intervals = append(intervals, time.Duration(ga.MaxAge))
	}
	if len(intervals) == 0 {
		return -1
	} else if len(intervals) == 1 {
		return intervals[0]
	}

	ggt := int64(intervals[0].Seconds())
	for i := 1; i < len(intervals); i++ {
		ggt = gcd(ggt, int64(intervals[i].Seconds()))
	}
	return time.Duration(ggt) * time.Second
}

// greatest common divisor (GCD) via Euclidean algorithm
func gcd(a, b int64) int64 {
	for b != 0 {
		t := b
		b = a % b
		a = t
	}
	return a
}
