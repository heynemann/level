// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>
// This file is part of {{ .appName }}.

package cmd

import (
	"github.com/heynemann/level/channel"
	"github.com/spf13/cobra"
	"github.com/uber-go/zap"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := zap.NewJSON(zap.DebugLevel)
		opts := channel.NewOptions(
			"0.0.0.0",
			3000,
			true,
			"./config/default.yaml",
		)
		channel, err := channel.New(opts, logger)
		if err != nil {
			logger.Error("Could not start channel...", zap.Error(err))
			return
		}
		channel.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
