package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunConvertGaCommand(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{"full", "../pkg/knx/fixtures/ga-export.xml", false},
		{"source do not exists", "fixtures/invalid.xml", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := ioutil.TempFile("", "")
			assert.NoError(t, err)
			defer os.Remove(tmpFile.Name())
			cmd := NewConvertGaCommand()

			if err := cmd.RunE(nil, []string{tt.src, tmpFile.Name()}); (err != nil) != tt.wantErr {
				t.Errorf("ConvertGroupAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assert.FileExists(t, tmpFile.Name())
		})
	}
}
