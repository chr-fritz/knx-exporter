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
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/chr-fritz/knx-exporter/version"
)

type VersionOptions struct{}

func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

func NewVersionCommand() *cobra.Command {
	versionOptions := NewVersionOptions()

	return &cobra.Command{
		Use:   "version",
		Short: "Show the version information",
		Args:  cobra.NoArgs,
		Run:   versionOptions.run,
	}
}

func (v *VersionOptions) run(_ *cobra.Command, _ []string) {
	parsedDate, _ := strconv.ParseInt(version.CommitDate, 10, 64)
	commitDate := time.Unix(parsedDate, 0).Format("2006-01-02 15:04:05 Z07:00")
	fmt.Printf(`KNX Prometheus Exporter
Version:     %s
Commit:      %s
Commit Date: %s
Branch:      %s
`,
		version.Version,
		version.Revision,
		commitDate,
		version.Branch,
	)
}

func init() {
	rootCmd.AddCommand(NewVersionCommand())
}
