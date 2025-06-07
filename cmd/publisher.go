package cmd

import (
	"adehikmatfr/go-rabbitmq/producer"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var exchangeType string

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a message to RabbitMQ",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		switch exchangeType {
		case "direct":
			err = producer.PublishDirect()
		case "fanout":
			err = producer.PublishFanout()
		case "topic":
			err = producer.PublishTopic()
		case "headers":
			err = producer.PublishHeaders()
		default:
			log.Fatal().Msgf("Unknown exchange type: %s", exchangeType)
			return
		}

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to publish message")
		} else {
			log.Info().Msgf("Published message to %s exchange", exchangeType)
		}
	},
}

func init() {
	publishCmd.Flags().StringVar(&exchangeType, "exchange-type", "direct", "Type of exchange to publish to (direct, fanout, topic, headers)")
	rootCmd.AddCommand(publishCmd)
}
