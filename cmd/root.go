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
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd *cobra.Command
var rootCmdOptions *RootOptions

type RootOptions struct {
	configFile string
}

func NewRootOptions() *RootOptions {
	return &RootOptions{}
}

func NewRootCommand() (*cobra.Command, *RootOptions) {
	rootOptions := NewRootOptions()
	cmd := &cobra.Command{
		Use:   "knx-exporter",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}

	cmd.PersistentFlags().StringVar(&rootOptions.configFile, "config", "", "config file (default is $HOME/.knx-exporter.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	return cmd, rootOptions
}

// initConfig reads in config file and ENV variables if set.
func (o *RootOptions) initConfig() {
	if o.configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(o.configFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".knx-exporter" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".knx-exporter")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd, rootCmdOptions = NewRootCommand()
	cobra.OnInitialize(rootCmdOptions.initConfig)
}
