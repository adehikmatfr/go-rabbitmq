package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rabbitcli",
	Short: "CLI to test RabbitMQ",
}

func Execute() error {
	return rootCmd.Execute()
}
