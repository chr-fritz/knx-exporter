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
	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/chr-fritz/knx-exporter/pkg/logging"
)

func init() {
	glog.V(0)
	loggerConfig := logging.InitFlags(rootCmd.PersistentFlags(), rootCmd)
	cobra.OnInitialize(loggerConfig.Initialize)
}
