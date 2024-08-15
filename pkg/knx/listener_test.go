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
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
)

func Test_listener_Run(t *testing.T) {
	tests := []struct {
		name      string
		event     knx.GroupEvent
		want      *Snapshot
		wantError bool
	}{
		{
			"bool false",
			knx.GroupEvent{Destination: cemi.GroupAddr(1), Data: []byte{0}},
			&Snapshot{name: "knx_a", value: 0, destination: GroupAddress(1), config: &GroupAddressConfig{Name: "a", DPT: "1.001", Export: true}},
			false,
		},
		{
			"bool true",
			knx.GroupEvent{Destination: cemi.GroupAddr(1), Data: []byte{1}},
			&Snapshot{name: "knx_a", value: 1, destination: GroupAddress(1), config: &GroupAddressConfig{Name: "a", DPT: "1.001", Export: true}},
			false,
		},
		{
			"5.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(2), Data: []byte{0, 255}},
			&Snapshot{name: "knx_b", value: 100, destination: GroupAddress(2), config: &GroupAddressConfig{Name: "b", DPT: "5.001", Export: true}},
			false,
		},
		{
			"9.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(3), Data: []byte{0, 2, 38}},
			&Snapshot{name: "knx_c", value: 5.5, destination: GroupAddress(3), config: &GroupAddressConfig{Name: "c", DPT: "9.001", Export: true}},
			false,
		},
		{
			"12.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(4), Data: []byte{0, 0, 0, 0, 5}},
			&Snapshot{name: "knx_d", value: 5, destination: GroupAddress(4), config: &GroupAddressConfig{Name: "d", DPT: "12.001", Export: true}},
			false,
		},
		{
			"13.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(5), Data: []byte{0, 0, 0, 0, 5}},
			&Snapshot{name: "knx_e", value: 5, destination: GroupAddress(5), config: &GroupAddressConfig{Name: "e", DPT: "13.001", Export: true}},
			false,
		},
		{
			"14.*",
			knx.GroupEvent{Destination: cemi.GroupAddr(6), Data: []byte{0, 63, 192, 0, 0}},
			&Snapshot{name: "knx_f", value: 1.5, destination: GroupAddress(6), config: &GroupAddressConfig{Name: "f", DPT: "14.001", Export: true}},
			false,
		},
		{
			"5.* can't unpack",
			knx.GroupEvent{Destination: cemi.GroupAddr(2), Data: []byte{0}},
			nil,
			true,
		},
		{
			"bool unexported",
			knx.GroupEvent{Destination: cemi.GroupAddr(7), Data: []byte{1}},
			nil,
			true,
		},
		{
			"unknown address",
			knx.GroupEvent{Destination: cemi.GroupAddr(255), Data: []byte{0}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			inbound := make(chan knx.GroupEvent)
			metricsChan := make(chan *Snapshot)

			l := NewListener(
				&Config{
					MetricsPrefix: "knx_",
					AddressConfigs: map[GroupAddress]*GroupAddressConfig{
						GroupAddress(1): {Name: "a", DPT: "1.001", Export: true},
						GroupAddress(2): {Name: "b", DPT: "5.001", Export: true},
						GroupAddress(3): {Name: "c", DPT: "9.001", Export: true},
						GroupAddress(4): {Name: "d", DPT: "12.001", Export: true},
						GroupAddress(5): {Name: "e", DPT: "13.001", Export: true},
						GroupAddress(6): {Name: "f", DPT: "14.001", Export: true},
						GroupAddress(7): {Export: false},
					},
				},
				inbound,
				metricsChan,
				prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"direction", "processed"}),
			)

			go l.Run()
			inbound <- tt.event

			select {
			case got := <-metricsChan:
				// ignore timestamps
				got.timestamp = time.Unix(0, 0)
				tt.want.timestamp = time.Unix(0, 0)
				assert.Equal(t, tt.want, got)
			case <-time.After(10 * time.Millisecond):
				assert.True(t, tt.wantError, "got no metrics snapshot but requires one")
			}
			close(inbound)
			close(metricsChan)
		})
	}
}
