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
