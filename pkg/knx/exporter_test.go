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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/chr-fritz/knx-exporter/pkg/metrics/fake"
)

func TestNewMetricsExporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	exporter := fake.NewMockExporter(ctrl)
	exporter.EXPECT().Register(gomock.Any()).AnyTimes()
	exp, err := NewMetricsExporter("fixtures/readConfig.yaml", exporter)
	metricsExporter, ok := exp.(*metricsExporter)
	assert.True(t, ok)
	assert.NoError(t, err)

	err = metricsExporter.Run()
	assert.NoError(t, err)

	assert.NotNil(t, metricsExporter.metrics)
	assert.NotNil(t, metricsExporter.poller)
	assert.NotNil(t, metricsExporter.listener)

	time.Sleep(1 * time.Second)

	metricsExporter.Close()
}

func TestMetricsExporter_createClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{"wrong-type", &Config{Connection: Connection{Type: ConnectionType("wrong")}}, true},
		{"tunnel", &Config{Connection: Connection{Type: Tunnel, Endpoint: "127.0.0.1:3761"}}, true},
		{"router", &Config{Connection: Connection{Type: Router, Endpoint: "224.0.0.120:3672"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &metricsExporter{
				config: tt.config,
			}
			if err := e.createClient(); (err != nil) != tt.wantErr {
				t.Errorf("createClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
