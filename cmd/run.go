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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chr-fritz/knx-exporter/pkg/metrics"
)

const RunPortParm = "exporter.port"

type RunOptions struct {
	port uint16
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
		RunE:  runOptions.run,
	}
	cmd.Flags().Uint16VarP(&runOptions.port, "port", "p", 8080, "The port where all metrics should be exported.")
	_ = viper.BindPFlag(RunPortParm, cmd.Flags().Lookup("port"))
	return &cmd
}

func (i *RunOptions) run(cmd *cobra.Command, args []string) error {
	exporter := metrics.NewExporter(i.port)

	return exporter.Run()
}

func init() {
	rootCmd.AddCommand(NewRunCommand())
}
