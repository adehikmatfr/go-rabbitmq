package cmd

import (
	"adehikmatfr/go-rabbitmq/producer"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a message to RabbitMQ",
	Run: func(cmd *cobra.Command, args []string) {
		if err := producer.PublishDirect(); err != nil {
			log.Fatal().Err(err).Msg("Failed to publish message")
		}
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
}
