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

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/chr-fritz/knx-exporter/pkg/knx"
)

type ConvertGaOptions struct{}

func NewConvertGaOptions() *ConvertGaOptions {
	return &ConvertGaOptions{}
}

func NewConvertGaCommand() *cobra.Command {
	convertGaOptions := NewConvertGaOptions()

	return &cobra.Command{
		Use:   "convertGA [sourceFile] [targetFile]",
		Short: "Converts the ETS 5 XML group address export into the configuration format.",
		Long: `Converts the ETS 5 XML group address export into the configuration format.

It takes the XML group address export from the ETS 5 tool and converts it into the yaml format
used by the exporter.`,
		Args:              cobra.ExactArgs(2),
		RunE:              convertGaOptions.run,
		ValidArgsFunction: convertGaOptions.ValidArgs,
	}
}

// ValidArgs returns a list of possible arguments.
func (i *ConvertGaOptions) ValidArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return []string{"xml"}, cobra.ShellCompDirectiveFilterFileExt
	} else {
		return nil, cobra.ShellCompDirectiveFilterDirs
	}
}

func (i *ConvertGaOptions) run(_ *cobra.Command, args []string) error {
	return knx.ConvertGroupAddresses(args[0], args[1])
}

func init() {
	rootCmd.AddCommand(NewConvertGaCommand())
}
