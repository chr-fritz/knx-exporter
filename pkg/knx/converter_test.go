package knx

import (
	"io/ioutil"
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
			tmpFile, err := ioutil.TempFile("", "")
			assert.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			if err := ConvertGroupAddresses(tt.src, tmpFile.Name()); (err != nil) != tt.wantErr {
				t.Errorf("ConvertGroupAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assert.FileExists(t, tmpFile.Name())

			expected, err := ioutil.ReadFile(tt.wantTarget)
			assert.NoError(t, err)
			expected, err = yaml.YAMLToJSON(expected)
			assert.NoError(t, err)
			actual, err := ioutil.ReadFile(tmpFile.Name())
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
