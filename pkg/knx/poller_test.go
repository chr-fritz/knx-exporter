// Copyright © 2020-2025 Christian Fritz <mail@chr-fritz.de>
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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
)

func Test_getMetricsToPoll(t *testing.T) {

	tests := []struct {
		name   string
		config *Config
		want   GroupAddressConfigSet
	}{
		{"empty", &Config{AddressConfigs: GroupAddressConfigSet{}}, GroupAddressConfigSet{}},
		{"single-no-active-read", &Config{AddressConfigs: GroupAddressConfigSet{0: &GroupAddressConfig{ReadActive: false}}}, GroupAddressConfigSet{}},
		{"single-too-small-interval", &Config{AddressConfigs: GroupAddressConfigSet{0: &GroupAddressConfig{ReadActive: true, MaxAge: Duration(10 * time.Millisecond)}}}, GroupAddressConfigSet{}},
		{"single-no-export", &Config{AddressConfigs: GroupAddressConfigSet{0: &GroupAddressConfig{ReadActive: true, MaxAge: Duration(10 * time.Second), Export: false}}}, GroupAddressConfigSet{}},
		{"single-small-interval", &Config{
			AddressConfigs: GroupAddressConfigSet{0: &GroupAddressConfig{ReadActive: true, MaxAge: Duration(1 * time.Second), Name: "a", Export: true}},
			MetricsPrefix:  "knx_",
		}, GroupAddressConfigSet{0: &GroupAddressConfig{Name: "knx_a", ReadActive: true, MaxAge: Duration(5 * time.Second)}}},
		{"single", &Config{
			AddressConfigs: GroupAddressConfigSet{0: &GroupAddressConfig{ReadActive: true, MaxAge: Duration(10 * time.Second), Name: "a", Export: true}},
			MetricsPrefix:  "knx_",
		}, GroupAddressConfigSet{0: &GroupAddressConfig{Name: "knx_a", ReadActive: true, MaxAge: Duration(10 * time.Second)}}},
		{"multiple", &Config{
			AddressConfigs: GroupAddressConfigSet{
				0: &GroupAddressConfig{ReadActive: true, MaxAge: Duration(10 * time.Second), Name: "a", Export: true},
				1: &GroupAddressConfig{ReadActive: true, MaxAge: Duration(15 * time.Second), Name: "b", Export: true},
				2: &GroupAddressConfig{ReadActive: true, MaxAge: Duration(45 * time.Second), Name: "c", Export: true},
			},
			MetricsPrefix: "knx_",
		}, GroupAddressConfigSet{
			0: &GroupAddressConfig{Name: "knx_a", ReadActive: true, MaxAge: Duration(10 * time.Second)},
			1: &GroupAddressConfig{Name: "knx_b", ReadActive: true, MaxAge: Duration(15 * time.Second)},
			2: &GroupAddressConfig{Name: "knx_c", ReadActive: true, MaxAge: Duration(45 * time.Second)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMetricsToPoll(tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

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
			&Config{AddressConfigs: GroupAddressConfigSet{0: &GroupAddressConfig{ReadStartup: false, Export: false}}},
			GroupAddressConfigSet{},
		},
		{
			"single-no-export-startup-read",
			&Config{AddressConfigs: GroupAddressConfigSet{0: &GroupAddressConfig{Export: false, ReadStartup: true}}},
			GroupAddressConfigSet{},
		},
		{
			"single-export-no-startup-read",
			&Config{AddressConfigs: GroupAddressConfigSet{0: &GroupAddressConfig{Export: true, ReadStartup: false}}},
			GroupAddressConfigSet{},
		},
		{
			"single-export-startup-read",
			&Config{AddressConfigs: GroupAddressConfigSet{0: &GroupAddressConfig{Export: true, ReadStartup: true}}},
			GroupAddressConfigSet{0: &GroupAddressConfig{ReadStartup: true}},
		},
		{
			"multiple-export-startup-read",
			&Config{AddressConfigs: GroupAddressConfigSet{
				0: &GroupAddressConfig{Export: false, ReadStartup: false},
				1: &GroupAddressConfig{Export: true, ReadStartup: false},
				2: &GroupAddressConfig{Export: false, ReadStartup: true},
				3: &GroupAddressConfig{Export: true, ReadStartup: true},
				4: &GroupAddressConfig{Export: true, ReadStartup: true},
			}},
			GroupAddressConfigSet{
				3: &GroupAddressConfig{ReadStartup: true},
				4: &GroupAddressConfig{ReadStartup: true},
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

func TestPoller_Polling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	groupClient := NewMockGroupClient(ctrl)
	mockSnapshotHandler := NewMockMetricSnapshotHandler(ctrl)
	messageCounter := prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"direction", "processed"})

	config, err := ReadConfig("fixtures/readConfig.yaml")

	assert.NoError(t, err)

	mockSnapshotHandler.EXPECT().
		FindYoungestSnapshot("knx_dummy_metric").
		Return(&Snapshot{
			name:      "knx_dummy_metric",
			timestamp: time.Now().Add(-14 * time.Second),
			config:    &GroupAddressConfig{},
		})
	mockSnapshotHandler.EXPECT().
		FindYoungestSnapshot("knx_dummy_metric1").
		Return(&Snapshot{
			name:      "knx_dummy_metric1",
			timestamp: time.Now().Add(2 * time.Second),
			config:    &GroupAddressConfig{},
		})
	mockSnapshotHandler.EXPECT().
		FindYoungestSnapshot("knx_dummy_metric2").
		Return(nil)
	mockSnapshotHandler.EXPECT().
		FindYoungestSnapshot("knx_dummy_metric4").
		Return(&Snapshot{
			name:      "knx_dummy_metric4",
			timestamp: time.Now().Add(-16 * time.Second),
			config:    &GroupAddressConfig{},
		})

	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 1),
	}).Times(1)
	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 2),
	}).Times(1)
	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 3),
	}).Times(1)
	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 4),
	}).Times(0)
	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 5),
	}).Times(0)
	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupWrite, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 6), Data: []byte{1},
	}).Times(1)

	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 1),
	}).Times(1)
	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 3),
	}).Times(1)
	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupRead, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 5),
	}).Times(0)
	groupClient.EXPECT().Send(knx.GroupEvent{
		Command: knx.GroupWrite, Source: cemi.NewIndividualAddr3(2, 0, 1), Destination: cemi.NewGroupAddr3(0, 0, 6), Data: []byte{1},
	}).Times(1)

	p := NewPoller(config, groupClient, mockSnapshotHandler, messageCounter)
	p.Run()
	time.Sleep(5500 * time.Millisecond)

	p.Close()
}
