package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/openlab-red/kiam-webhook/pkg/engine"
)

var handlerCmd = &cobra.Command{
	Use:   "start",
	Short: "Start webhook server",
	Long:  `Start webhook server`,
	Run: func(cmd *cobra.Command, args []string) {
		engine.Start()
	},
}

func init() {
	RootCmd.AddCommand(handlerCmd)
	viper.SetDefault("log-level", "INFO")
	viper.SetDefault("port", "8080")
}
