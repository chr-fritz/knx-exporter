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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type LoggerConfiguration interface {
	Initialize()
}

type loggerConfig struct {
	flagSet       *pflag.FlagSet
	level         string
	formatterName string
}

func InitFlags(flagset *pflag.FlagSet, cmd *cobra.Command) LoggerConfiguration {
	if flagset == nil {
		flagset = pflag.CommandLine
	}
	config := &loggerConfig{
		flagSet: flagset,
	}

	logLevelFlagName := "log_level"
	logFormatterFlagName := "log_format"
	flagset.StringVarP(&config.level, logLevelFlagName, "v", "info", "The minimum log level to print the messages.")
	flagset.StringVarP(&config.formatterName, logFormatterFlagName, "", "text", "The format how to print the log messages.")

	if cmd != nil {
		if e := cmd.RegisterFlagCompletionFunc(logLevelFlagName, flagCompletion); e != nil {
			logrus.Warn("can not register flag completion for log_level: ", e)
		}

		e := cmd.RegisterFlagCompletionFunc(logFormatterFlagName, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"text", "json"}, cobra.ShellCompDirectiveDefault
		})
		if e != nil {
			logrus.Warn("can not register flag completion for log formatter: ", e)
		}
	}

	return config
}

func flagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"panic", "fatal", "error", "warn", "warning", "info", "debug", "trace"}, cobra.ShellCompDirectiveDefault
}

func (lc *loggerConfig) Initialize() {
	lc.setFormatter()
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
func (lc *loggerConfig) setFormatter() {
	var formatter logrus.Formatter
	switch strings.ToLower(lc.formatterName) {
	case "json":
		formatter = &logrus.JSONFormatter{}
	case "text":
		formatter = &logrus.TextFormatter{}
	default:
		formatter = &logrus.TextFormatter{}
	}

	logrus.SetFormatter(formatter)
}
