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
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func TestConvertGroupAddresses(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		wantTarget string
		wantErr    bool
	}{
		{"full", "fixtures/ga-export.xml", "fixtures/ga-config.yaml", false},
		{"source do not exists", "fixtures/invalid.xml", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "")
			assert.NoError(t, err)
			defer func() {
				_ = os.Remove(tmpFile.Name())
			}()

			if err := ConvertGroupAddresses(tt.src, tmpFile.Name()); (err != nil) != tt.wantErr {
				t.Errorf("ConvertGroupAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assert.FileExists(t, tmpFile.Name())

			expected, err := os.ReadFile(tt.wantTarget)
			assert.NoError(t, err)
			expected, err = yaml.YAMLToJSON(expected)
			assert.NoError(t, err)
			actual, err := os.ReadFile(tmpFile.Name())
			assert.NoError(t, err)
			actual, err = yaml.YAMLToJSON(actual)
			assert.NoError(t, err)

			assert.JSONEq(t, string(expected), string(actual))
		})
	}
}

func Test_normalizeMetricName(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{"is_valid_regex", "is_valid_regex", false},
		{"Is_1_valid_regex", "Is_1_valid_regex", false},
		{"eine gültige ga", "eine_gueltige_ga", false},
		{"ÄÜÖäüöß", "AeUeOeaeueoess", false},
		{"6_asdf", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeMetricName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeMetricName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_normalizeDPTs(t *testing.T) {
	tests := []struct {
		name    string
		dpt     string
		want    string
		wantErr bool
	}{
		{"DPST-1-1", "DPST-1-1", "1.001", false},
		{"DPST-1-5", "DPST-1-5", "1.005", false},
		{"DPST-1-7", "DPST-1-7", "1.007", false},
		{"DPST-1-8", "DPST-1-8", "1.008", false},
		{"DPST-20-102", "DPST-20-102", "20.102", false},
		{"DPST-3-7", "DPST-3-7", "3.007", false},
		{"DPST-5-1", "DPST-5-1", "5.001", false},
		{"DPST-7-7", "DPST-7-7", "7.007", false},
		{"DPST-9-1", "DPST-9-1", "9.001", false},
		{"DPT-1", "DPT-1", "1.*", false},
		{"DPT-13", "DPT-13", "13.*", false},
		{"DPT-5", "DPT-5", "5.*", false},
		{"DPT-7", "DPT-7", "7.*", false},
		{"DPT-9", "DPT-9", "9.*", false},
		{"invalid", "DPT9", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDPTs(tt.dpt)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeDPTs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
