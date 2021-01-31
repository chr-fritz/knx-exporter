package logging

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type LoggerConfiguration interface {
	Initialize()
}

type loggerConfig struct {
	flagSet *pflag.FlagSet
	level   string
}

func InitFlags(flagset *pflag.FlagSet, cmd *cobra.Command) LoggerConfiguration {
	if flagset == nil {
		flagset = pflag.CommandLine
	}
	config := &loggerConfig{
		flagSet: flagset,
	}

	flagName := "log_level"
	flagset.StringVarP(&config.level, flagName, "v", "info", "The minimum log level to print the messages")

	if cmd != nil {
		if e := cmd.RegisterFlagCompletionFunc(flagName, flagCompletion); e != nil {
			logrus.Warn("can not register flag completion for log_level: ", e)
		}
	}

	return config
}

func flagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"panic", "fatal", "error", "warn", "warning", "info", "debug", "trace"}, cobra.ShellCompDirectiveDefault
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
