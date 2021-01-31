package knx

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"

	"github.com/chr-fritz/knx-exporter/pkg/knx/fake"
)

func Test_getMetricsToPoll(t *testing.T) {

	tests := []struct {
		name   string
		config *Config
		want   GroupAddressConfigSet
	}{
		{"empty", &Config{AddressConfigs: GroupAddressConfigSet{}}, GroupAddressConfigSet{}},
		{"single-no-active-read", &Config{AddressConfigs: GroupAddressConfigSet{0: GroupAddressConfig{ReadActive: false}}}, GroupAddressConfigSet{}},
		{"single-too-small-interval", &Config{AddressConfigs: GroupAddressConfigSet{0: GroupAddressConfig{ReadActive: true, MaxAge: Duration(10 * time.Millisecond)}}}, GroupAddressConfigSet{}},
		{"single-no-export", &Config{AddressConfigs: GroupAddressConfigSet{0: GroupAddressConfig{ReadActive: true, MaxAge: Duration(10 * time.Second), Export: false}}}, GroupAddressConfigSet{}},
		{"single-small-interval", &Config{
			AddressConfigs: GroupAddressConfigSet{0: GroupAddressConfig{ReadActive: true, MaxAge: Duration(1 * time.Second), Name: "a", Export: true}},
			MetricsPrefix:  "knx_",
		}, GroupAddressConfigSet{0: GroupAddressConfig{Name: "knx_a", ReadActive: true, MaxAge: Duration(5 * time.Second)}}},
		{"single", &Config{
			AddressConfigs: GroupAddressConfigSet{0: GroupAddressConfig{ReadActive: true, MaxAge: Duration(10 * time.Second), Name: "a", Export: true}},
			MetricsPrefix:  "knx_",
		}, GroupAddressConfigSet{0: GroupAddressConfig{Name: "knx_a", ReadActive: true, MaxAge: Duration(10 * time.Second)}}},
		{"multiple", &Config{
			AddressConfigs: GroupAddressConfigSet{
				0: GroupAddressConfig{ReadActive: true, MaxAge: Duration(10 * time.Second), Name: "a", Export: true},
				1: GroupAddressConfig{ReadActive: true, MaxAge: Duration(15 * time.Second), Name: "b", Export: true},
				2: GroupAddressConfig{ReadActive: true, MaxAge: Duration(45 * time.Second), Name: "c", Export: true},
			},
			MetricsPrefix: "knx_",
		}, GroupAddressConfigSet{
			0: GroupAddressConfig{Name: "knx_a", ReadActive: true, MaxAge: Duration(10 * time.Second)},
			1: GroupAddressConfig{Name: "knx_b", ReadActive: true, MaxAge: Duration(15 * time.Second)},
			2: GroupAddressConfig{Name: "knx_c", ReadActive: true, MaxAge: Duration(45 * time.Second)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMetricsToPoll(tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_calcPollingInterval(t *testing.T) {
	tests := []struct {
		name      string
		addresses GroupAddressConfigSet
		want      time.Duration
	}{
		{"empty", GroupAddressConfigSet{}, -1},
		{"single", GroupAddressConfigSet{0: {ReadActive: true, MaxAge: Duration(10 * time.Second)}}, 10 * time.Second},
		{"multiple", GroupAddressConfigSet{
			0: {ReadActive: true, MaxAge: Duration(10 * time.Second)},
			1: {ReadActive: true, MaxAge: Duration(15 * time.Second)},
			2: {ReadActive: true, MaxAge: Duration(30 * time.Second)},
			3: {ReadActive: true, MaxAge: Duration(45 * time.Second)},
			4: {ReadActive: true, MaxAge: Duration(60 * time.Second)},
			5: {ReadActive: true, MaxAge: Duration(90 * time.Second)},
		}, 5 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcPollingInterval(tt.addresses)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPoller(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	exporter, err := NewMetricsExporter("fixtures/readConfig.yaml")
	assert.NoError(t, err)
	groupClient := fake.NewMockGroupClient(ctrl)
	exporter.client = groupClient
	go exporter.storeSnapshots()

	exporter.metricsChan <- metricSnapshot{
		name:      "knx_dummy_metric",
		timestamp: time.Now().Add(-14 * time.Second),
	}
	exporter.metricsChan <- metricSnapshot{
		name:      "knx_dummy_metric1",
		timestamp: time.Now().Add(2 * time.Second),
	}

	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 1),
	}).Times(1)
	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 3),
	}).Times(1)

	p := NewPoller(exporter)
	p.Run()
	time.Sleep(5500 * time.Millisecond)

	p.Stop()
	close(exporter.metricsChan)
}
