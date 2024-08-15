// Copyright © 2024 Christian Fritz <mail@chr-fritz.de>
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

	gomock "github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
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

func TestStartupReader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	groupClient := NewMockGroupClient(ctrl)
	mockSnapshotHandler := NewMockMetricSnapshotHandler(ctrl)
	messageCounter := prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"direction", "processed"})

	config, err := ReadConfig("fixtures/readConfig.yaml")

	assert.NoError(t, err)

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

	s := NewStartupReader(config, groupClient, mockSnapshotHandler, messageCounter)
	s.Run()
	time.Sleep(2000 * time.Millisecond)

	s.Close()
}
