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

package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/heptiolabs/healthcheck"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chr-fritz/knx-exporter/pkg/knx"
	"github.com/chr-fritz/knx-exporter/pkg/metrics"
)

const RunPortParm = "exporter.port"
const RunConfigFileParm = "exporter.configFile"
const RunRestartParm = "exporter.restart"
const WithGoMetricsParamName = "exporter.goMetrics"

type RunOptions struct {
	aliveCheckInterval time.Duration
}

func NewRunOptions() *RunOptions {
	return &RunOptions{
		aliveCheckInterval: 10 * time.Second,
	}
}

func NewRunCommand() *cobra.Command {
	runOptions := NewRunOptions()

	cmd := cobra.Command{
		Use:   "run",
		Short: "Run the exporter",
		Long:  `Run the exporter which exports the received values from all configured Group Addresses to prometheus.`,
		Args:  cobra.NoArgs,
		Run:   runOptions.run,
	}

	cmd.Flags().Uint16P("port", "p", 8080, "The port where all metrics should be exported.")
	cmd.Flags().StringP("configFile", "f", "config.yaml", "The knx configuration file.")
	cmd.Flags().StringP("restart", "r", "health", "The restart behaviour. Can be health or exit")
	cmd.Flags().BoolP("withGoMetrics", "g", true, "Should the go metrics also be exported?")

	_ = viper.BindPFlag(RunPortParm, cmd.Flags().Lookup("port"))
	_ = viper.BindPFlag(RunConfigFileParm, cmd.Flags().Lookup("configFile"))
	_ = viper.BindPFlag(RunRestartParm, cmd.Flags().Lookup("restart"))
	_ = viper.BindPFlag(WithGoMetricsParamName, cmd.Flags().Lookup("withGoMetrics"))

	_ = cmd.RegisterFlagCompletionFunc("configFile", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "yml"}, cobra.ShellCompDirectiveFilterFileExt
	})
	_ = cmd.RegisterFlagCompletionFunc("port", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("restart", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"health", "exit"}, cobra.ShellCompDirectiveDefault
	})
	return &cmd
}

func (i *RunOptions) run(_ *cobra.Command, _ []string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	exporter := metrics.NewExporter(uint16(viper.GetUint(RunPortParm)), viper.GetBool(WithGoMetricsParamName))

	exporter.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))
	metricsExporter, err := i.initAndRunMetricsExporter(ctx, exporter)
	if err != nil {
		slog.Error("Unable to init metrics exporter: " + err.Error())
		return
	}

	go i.aliveCheck(ctx, stop, metricsExporter)

	if err = exporter.Run(ctx); err != nil {
		slog.Error("Can not run metrics exporter: " + err.Error())
	}
}

func (i *RunOptions) aliveCheck(ctx context.Context, cancelFunc context.CancelFunc, metricsExporter knx.MetricsExporter) {
	ticker := time.NewTicker(i.aliveCheckInterval)
	for {
		select {
		case <-ticker.C:
			aliveErr := metricsExporter.IsAlive()
			if aliveErr != nil {
				_, _ = daemon.SdNotify(false, "STATUS=Metrics Exporter is not alive anymore: "+aliveErr.Error())
				_, _ = daemon.SdNotify(false, "ERROR=1")
				if viper.GetString(RunRestartParm) == "exit" {
					cancelFunc()
				}
			}
		case <-ctx.Done():
			cancelFunc()
		}
	}
}

func (i *RunOptions) initAndRunMetricsExporter(ctx context.Context, exporter metrics.Exporter) (knx.MetricsExporter, error) {
	metricsExporter, err := knx.NewMetricsExporter(viper.GetString(RunConfigFileParm), exporter)
	if err != nil {
		return nil, err
	}

	exporter.AddLivenessCheck("knxConnection", metricsExporter.IsAlive)
	if e := metricsExporter.Run(ctx); e != nil {
		return nil, e
	}

	return metricsExporter, nil
}

func init() {
	rootCmd.AddCommand(NewRunCommand())
}
