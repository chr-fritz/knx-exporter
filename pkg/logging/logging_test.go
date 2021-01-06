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
