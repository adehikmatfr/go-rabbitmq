package rabbitmq

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type Kind string

const (
	KindDirect  Kind = "direct"
	KindFanout  Kind = "fanout"
	KindTopic   Kind = "topic"
	KindHeaders Kind = "headers"
)

type Client struct {
	conn *amqp.Connection
}

func NewClient(url string) (*Client, error) {
	var conn *amqp.Connection
	var err error

	for i := range 5 {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		// Retry dengan exponential backoff
		wait := time.Duration(i+1) * time.Second
		log.Info().Msgf("Try to connect RabbitMQ, retry attempt %v...\n", wait)
		time.Sleep(wait)
	}
	log.Info().Msg("RabbitMQ Connected")
	return &Client{
		conn: conn,
	}, err
}

func (c *Client) Close() error {
	return c.conn.Close()
}
