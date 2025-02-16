// Copyright © 2020-2025 Christian Fritz <mail@chr-fritz.de>
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
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"strings"

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

	_ = viper.BindPFlag("logging.level", flagset.Lookup(logLevelFlagName))
	_ = viper.BindPFlag("logging.format", flagset.Lookup(logFormatterFlagName))

	if cmd != nil {
		if e := cmd.RegisterFlagCompletionFunc(logLevelFlagName, flagCompletion); e != nil {
			slog.Warn("can not register flag completion for log_level", "err", e)
		}

		e := cmd.RegisterFlagCompletionFunc(logFormatterFlagName, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"text", "json"}, cobra.ShellCompDirectiveDefault
		})
		if e != nil {
			slog.Warn("can not register flag completion for log formatter", "err", e)
		}
	}

	return config
}

func flagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"error", "warn", "info", "debug"}, cobra.ShellCompDirectiveDefault
}

func (lc *loggerConfig) Initialize() {
	level := lc.setLevel()

	opts := &slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: nil,
	}
	logger := slog.New(lc.setFormatter(opts))
	slog.SetDefault(logger)
}

func (lc *loggerConfig) setLevel() slog.Level {
	var level slog.Level
	e := level.UnmarshalText([]byte(lc.level))

	if e != nil {
		slog.Warn("Can not parse level", "invalid-level", lc.level)
		return slog.LevelInfo
	}
	return level
}
func (lc *loggerConfig) setFormatter(options *slog.HandlerOptions) slog.Handler {
	switch strings.ToLower(lc.formatterName) {
	case "json":
		return slog.NewJSONHandler(os.Stdout, options)
	case "text":
		fallthrough
	default:
		return slog.NewTextHandler(os.Stdout, options)
	}
}
