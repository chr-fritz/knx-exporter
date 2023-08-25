package knx

import (
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
)

// StartupReader defines the interface for active polling for metrics values against the knx system at startup.
type StartupReader interface {
	// Run starts the startup reading.
	Run()
	// Close stops the startup reading.
	Close()
}

type startupReader struct {
	client          GroupClient
	config          *Config
	messageCounter  *prometheus.CounterVec
	snapshotHandler MetricSnapshotHandler
	metricsToRead   GroupAddressConfigSet
	ticker          *time.Ticker
}

// NewStartupReader creates a new StartupReader instance using the given MetricsExporter for connection handling and metrics observing.
func NewStartupReader(config *Config, client GroupClient, metricsHandler MetricSnapshotHandler, messageCounter *prometheus.CounterVec) StartupReader {
	metricsToRead := getMetricsToRead(config)
	return &startupReader{
		client:          client,
		config:          config,
		messageCounter:  messageCounter,
		snapshotHandler: metricsHandler,
		metricsToRead:   metricsToRead,
	}
}

func (s *startupReader) Run() {
	readInterval := time.Duration(s.config.ReadStartupInterval)
	if readInterval.Milliseconds() <= 0 {
		readInterval = 200 * time.Millisecond
	}
	logrus.Infof("start reading addresses after startup in %dms intervals.", readInterval.Milliseconds())
	s.ticker = time.NewTicker(readInterval)
	c := s.ticker.C
	go func() {
		addressesToRead := reflect.ValueOf(s.metricsToRead).MapKeys()
		for range c {
			if len(addressesToRead) == 0 {
				break
			}
			addressToRead := addressesToRead[0].Interface().(GroupAddress)
			s.sendReadMessage(addressToRead)
			addressesToRead = addressesToRead[1:]
		}
		s.ticker.Stop()
	}()
}

func (s *startupReader) Close() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
}

func (s *startupReader) sendReadMessage(address GroupAddress) {
	event := knx.GroupEvent{
		Command:     knx.GroupRead,
		Destination: cemi.GroupAddr(address),
		Source:      cemi.IndividualAddr(s.config.Connection.PhysicalAddress),
	}

	if e := s.client.Send(event); e != nil {
		logrus.Errorf("can not send read request for %s: %s", address.String(), e)
	}
	s.messageCounter.WithLabelValues("sent", "true").Inc()
}

func getMetricsToRead(config *Config) GroupAddressConfigSet {
	toRead := make(GroupAddressConfigSet)
	for address, addressConfig := range config.AddressConfigs {
		if !addressConfig.Export || !addressConfig.ReadStartup {
			continue
		}

		toRead[address] = GroupAddressConfig{
			Name:        config.NameFor(addressConfig),
			ReadStartup: true,
		}
	}
	return toRead
}
