// Copyright © 2020 Christian Fritz <mail@chr-fritz.de>
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
	"github.com/heptiolabs/healthcheck"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chr-fritz/knx-exporter/pkg/knx"
	"github.com/chr-fritz/knx-exporter/pkg/metrics"
)

const RunPortParm = "exporter.port"
const RunConfigFileParm = "exporter.configFile"

type RunOptions struct {
	port       uint16
	configFile string
}

func NewRunOptions() *RunOptions {
	return &RunOptions{
		port: 8080,
	}
}

func NewRunCommand() *cobra.Command {
	runOptions := NewRunOptions()

	cmd := cobra.Command{
		Use:   "run",
		Short: "Run the exporter",
		Long:  `Run the exporter which exports the received values from all configured Group Addresses to prometheus.`,
		Args:  cobra.NoArgs,
		RunE:  runOptions.run,
	}

	cmd.Flags().Uint16VarP(&runOptions.port, "port", "p", 8080, "The port where all metrics should be exported.")
	cmd.Flags().StringVarP(&runOptions.configFile, "configFile", "f", "config.yaml", "The knx configuration file.")
	_ = viper.BindPFlag(RunPortParm, cmd.Flags().Lookup("port"))
	_ = viper.BindPFlag(RunConfigFileParm, cmd.Flags().Lookup("configFile"))
	_ = cmd.RegisterFlagCompletionFunc("configFile", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "yml"}, cobra.ShellCompDirectiveFilterFileExt
	})
	_ = cmd.RegisterFlagCompletionFunc("port", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	return &cmd
}

func (i *RunOptions) run(_ *cobra.Command, _ []string) error {
	exporter := metrics.NewExporter(i.port)
	metricsExporter, err := knx.NewMetricsExporter(i.configFile, exporter)
	if err != nil {
		return err
	}
	defer metricsExporter.Close()

	exporter.AddLivenessCheck("knxConnection", metricsExporter.IsAlive)
	exporter.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))

	if e := metricsExporter.Run(); e != nil {
		return e
	}

	return exporter.Run()
}

func init() {
	rootCmd.AddCommand(NewRunCommand())
}
