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

package logging

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_loggerConfig_Initialize(t *testing.T) {
	tests := []struct {
		name              string
		level             string
		format            string
		expectedLevel     slog.Level
		expectedFormatter slog.Handler
	}{
		{
			"info text to stderr",
			"info",
			"text",
			slog.LevelInfo,
			&slog.TextHandler{},
		},
		{
			"info text as json",
			"info",
			"json",
			slog.LevelInfo,
			&slog.JSONHandler{},
		},
		{
			"unknown log formatter",
			"info",
			"unknown",
			slog.LevelInfo,
			&slog.TextHandler{},
		},
		{
			"invalid debug level",
			"not valid",
			"text",
			slog.LevelInfo,
			&slog.TextHandler{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := &loggerConfig{
				level:         tt.level,
				formatterName: tt.format,
			}
			lc.Initialize()
			logger := slog.With("dummy")
			assert.True(t, logger.Enabled(nil, tt.expectedLevel))
			assert.False(t, logger.Enabled(nil, tt.expectedLevel-1))
			assert.IsType(t, tt.expectedFormatter, logger.Handler())
		})
	}
}
