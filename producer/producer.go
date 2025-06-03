package producer

import (
	"adehikmatfr/go-rabbitmq/pkg/rabbitmq"
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

func PublishDirect() error {
	client, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
		return nil
	}
	publisher, err := client.NewPublisher(rabbitmq.PublisherOpts{
		Exchange:    "test-ade-exchange",
		Mandatory:   false,
		Immediate:   false,
		ContentType: "application/json",
		Expiration:  "10000",
		Headers:     amqp.Table{}, // jika ada custom headers
	})
	if err != nil {
		log.Err(err).Msg("failed to create publisher")
		return err
	}
	defer publisher.Close()

	body := []byte(`{"hello": "world"}`)

	err = publisher.Publish(context.Background(), "test-ade-bind", body)
	if err != nil {
		log.Err(err).Msg("failed to publish message")
	}

	return nil
}
