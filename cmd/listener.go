package cmd

import (
	"adehikmatfr/go-rabbitmq/consumer"
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen a message From RabbitMQ",
	Run: func(cmd *cobra.Command, args []string) {
		handler := consumer.NewRabbitMQListener(&consumer.RabbitMQListenerOptions{
			URL: "amqp://guest:guest@localhost:5672/",
		})

		go handler.RunListener()

		select {
		case err := <-handler.ListenError():
			fmt.Println("error starting rabbitmq service: %w", err)
		case <-handler.TerminateSignal():
			// Graceful shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := handler.Shutdown(ctx); err != nil {
				fmt.Printf("Failed to shutdown RabbitMQ service gracefully: %v\n", err)
			}
			fmt.Println("Exiting gracefully...")
		}
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)
}
