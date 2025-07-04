package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

var dialChannel = func(conn *amqp.Connection) (*amqp.Channel, error) {
	return conn.Channel()
}

var publishWithContext = func(ch *amqp.Channel, ctx context.Context, exchange, key string, mandatory, immediate bool, pub amqp.Publishing) error {
	return ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, pub)
}

var closeChannel = func(ch *amqp.Channel) error {
	return ch.Close()
}

type PublisherOpts struct {
	Exchange    string
	Mandatory   bool
	Immediate   bool
	ContentType string
	Priority    uint8
	Expiration  string
	Headers     amqp.Table
}

type Publisher struct {
	channel     *amqp.Channel
	exchange    string
	mandatory   bool
	immediate   bool
	contentType string
	priority    uint8
	expiration  string
	headers     amqp.Table
}

func (c *Client) NewPublisher(opts PublisherOpts) (*Publisher, error) {
	ch, err := dialChannel(c.conn)
	if err != nil {
		log.Err(err).Msgf("Open channel for publisher to exchange %s", opts.Exchange)
		return nil, err
	}

	return &Publisher{
		channel:     ch,
		exchange:    opts.Exchange,
		mandatory:   opts.Mandatory,
		immediate:   opts.Immediate,
		contentType: opts.ContentType,
		priority:    opts.Priority,
		expiration:  opts.Expiration,
		headers:     opts.Headers,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	err := publishWithContext(
		p.channel,
		ctx,
		p.exchange,
		routingKey,
		p.mandatory,
		p.immediate,
		amqp.Publishing{
			ContentType: p.contentType,
			Body:        body,
			Headers:     p.headers,
			Priority:    p.priority,
			Expiration:  p.expiration,
		},
	)

	if err != nil {
		log.Err(err).Msgf("Publish to exchange %s with routingKey %s", p.exchange, routingKey)
		return err
	}

	log.Info().Msgf("Message published to exchange %s with routingKey %s", p.exchange, routingKey)
	return nil
}

func (p *Publisher) Close() error {
	if err := closeChannel(p.channel); err != nil {
		log.Err(err).Msgf("Close publisher channel for exchange %s", p.exchange)
		return err
	}
	return nil
}
