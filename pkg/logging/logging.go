package logging

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type LoggerConfiguration interface {
	Initialize()
}

type loggerConfig struct {
	flagSet *pflag.FlagSet
	level   string
}

func InitFlags(flagset *pflag.FlagSet) LoggerConfiguration {
	if flagset == nil {
		flagset = pflag.CommandLine
	}
	config := &loggerConfig{
		flagSet: flagset,
	}

	flagset.StringVarP(&config.level, "log_level", "v", "info", "The minimum log level to print the messages")

	return config
}

func (lc *loggerConfig) Initialize() {
	e := lc.setLevel()

	if e != nil {
		logrus.Warnf("Unable to fully initialize logrus: %s", e)
	}
}

func (lc *loggerConfig) setLevel() error {
	level, e := logrus.ParseLevel(lc.level)
	if e != nil {
		return e
	}
	logrus.SetLevel(level)
	return nil
}
