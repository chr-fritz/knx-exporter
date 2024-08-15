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

package cmd

import (
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
			tmpFile, err := os.CreateTemp("", "")
			assert.NoError(t, err)
			defer func() {
				_ = os.Remove(tmpFile.Name())
			}()
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
