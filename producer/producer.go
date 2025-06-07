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

func PublishFanout() error {
	client, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
		return nil
	}

	publisher, err := client.NewPublisher(rabbitmq.PublisherOpts{
		Exchange:    "fanout-exchange",
		Mandatory:   false,
		Immediate:   false,
		ContentType: "application/json",
	})
	if err != nil {
		log.Err(err).Msg("failed to create fanout publisher")
		return err
	}
	defer publisher.Close()

	body := []byte(`{"event": "broadcast message"}`)

	// routingKey kosong untuk Fanout
	err = publisher.Publish(context.Background(), "", body)
	if err != nil {
		log.Err(err).Msg("failed to publish fanout message")
	}

	return nil
}

func PublishTopic() error {
	client, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
		return nil
	}

	publisher, err := client.NewPublisher(rabbitmq.PublisherOpts{
		Exchange:    "topic-exchange",
		Mandatory:   false,
		Immediate:   false,
		ContentType: "application/json",
	})
	if err != nil {
		log.Err(err).Msg("failed to create topic publisher")
		return err
	}
	defer publisher.Close()

	body := []byte(`{"orderId": "12345", "status": "shipped"}`)

	// contoh routing key: order.created atau order.shipped
	err = publisher.Publish(context.Background(), "order.shipped", body)
	if err != nil {
		log.Err(err).Msg("failed to publish topic message")
	}

	return nil
}

func PublishHeaders() error {
	client, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
		return nil
	}

	publisher, err := client.NewPublisher(rabbitmq.PublisherOpts{
		Exchange:    "headers-exchange",
		Mandatory:   false,
		Immediate:   false,
		ContentType: "application/json",
		Headers: amqp.Table{
			"type":   "report",
			"format": "pdf",
		},
	})
	if err != nil {
		log.Err(err).Msg("failed to create headers publisher")
		return err
	}
	defer publisher.Close()

	body := []byte(`{"reportId": "98765", "content": "PDF report content"}`)

	// routingKey kosong untuk Headers exchange
	err = publisher.Publish(context.Background(), "", body)
	if err != nil {
		log.Err(err).Msg("failed to publish headers message")
	}

	return nil
}
