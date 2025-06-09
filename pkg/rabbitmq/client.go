package rabbitmq

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

var amqpDial = amqp.Dial

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
		conn, err = amqpDial(url)
		if err == nil {
			break
		} else {
			if i == 4 {
				log.Info().Msg("Failed to connect RabbitMQ, retry attempt 5")
				return nil, err
			}
		}
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
