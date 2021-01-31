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
