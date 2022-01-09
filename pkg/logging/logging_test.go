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

package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_loggerConfig_Initialize(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel logrus.Level
	}{
		{
			"info text to stderr",
			"info",
			logrus.InfoLevel,
		},
		{
			"invalid debug level",
			"not valid",
			logrus.InfoLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := &loggerConfig{
				level: tt.level,
			}
			lc.Initialize()
			logger := logrus.StandardLogger()
			assert.Equal(t, tt.expectedLevel, logger.Level)
		})
	}
}
