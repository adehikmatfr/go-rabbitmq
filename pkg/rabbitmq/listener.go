package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// LISTENER
type RabbitMQListener interface {
	Listen(ctx context.Context) error
	Close() error
}

type Listener struct {
	channel   *amqp.Channel
	queue     string
	consumer  string
	autoAck   bool
	exclusive bool
	noLocal   bool
	noWait    bool
	args      amqp.Table
}

type QueueBindOpts struct {
	Key      string
	Exchange string
	Args     amqp.Table
}

type QueueDeclareOpts struct {
	Durable    bool
	AutoDelete bool
	Args       amqp.Table
}

type ExchangeDeclareOpts struct {
	Name       string
	Kind       Kind
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

type QueueOpts struct {
	Declare   *QueueDeclareOpts
	Bind      *QueueBindOpts
	Name      string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

type ListenerOpts struct {
	Queue    QueueOpts
	Exchange *ExchangeDeclareOpts
	Consumer string
	Prefetch int
}

func (c *Client) NewListener(opts *ListenerOpts) (*Listener, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Err(err).Msgf("Open Channel for consumer %s", opts.Consumer)
		return nil, err
	}

	// safe mode queue
	if opts.Prefetch > 0 {
		ch.Qos(opts.Prefetch, 0, false) // await queue before ack
	}

	// declare queue
	if opts.Queue.Declare != nil {
		_, err = ch.QueueDeclare(
			opts.Queue.Name,
			opts.Queue.Declare.Durable,    // durable
			opts.Queue.Declare.AutoDelete, // config auto delete if server is
			opts.Queue.Exclusive,          // exclusive
			opts.Queue.NoWait,             // no-wait
			opts.Queue.Declare.Args,       // arguments (e.g. DLX)
		)
		if err != nil {
			log.Err(err).Msgf("Declare Queue %s for consumer %s", opts.Queue.Name, opts.Consumer)
			return nil, err
		}
	}

	// declare exchange
	if opts.Exchange != nil {
		if err := ch.ExchangeDeclare(
			opts.Exchange.Name,
			string(opts.Exchange.Kind),
			opts.Exchange.Durable,
			opts.Exchange.AutoDelete,
			opts.Exchange.Internal,
			opts.Exchange.NoWait,
			opts.Exchange.Args,
		); err != nil {
			log.Err(err).Msgf("Declare Exchange %s for consumer %s", opts.Queue.Name, opts.Consumer)
			return nil, err
		}
	}

	// Optional: Bind to exchange
	if opts.Queue.Bind != nil {
		if opts.Queue.Bind.Exchange == "" {
			err = fmt.Errorf("plase set exchange if use bind")
			log.Err(err).Msgf("Bind Queue exchange empty for consumer %s", opts.Consumer)
			return nil, err
		}
		err = ch.QueueBind(
			opts.Queue.Name,
			opts.Queue.Bind.Key,      // routing key
			opts.Queue.Bind.Exchange, // exchange name
			opts.Queue.NoWait,
			opts.Queue.Bind.Args,
		)
		if err != nil {
			log.Err(err).Msgf("Bind Queue consumer %s", opts.Consumer)
			return nil, err
		}
	}

	return &Listener{
		channel:   ch,
		queue:     opts.Queue.Name,
		consumer:  opts.Consumer,
		autoAck:   opts.Queue.AutoAck,
		exclusive: opts.Queue.Exclusive,
		noLocal:   opts.Queue.NoLocal,
		noWait:    opts.Queue.NoWait,
		args:      opts.Queue.Args,
	}, nil
}

func (l *Listener) Listen(ctx context.Context, handler func(context.Context, *amqp.Delivery) error) error {
	msgs, err := l.channel.ConsumeWithContext(
		ctx,
		l.queue,     // queue
		l.consumer,  // consumer
		l.autoAck,   // autoAck
		l.exclusive, // exclusive
		l.noLocal,   // noLocal
		l.noWait,    // noWait
		l.args,      // args
	)
	if err != nil {
		return err
	}

	log.Info().Msgf("Start listening on queue %s with consumer tag %s", l.queue, l.consumer)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				log.Info().Msg("Message channel closed, exiting listener.")
				return nil
			}
			if err := handler(ctx, &msg); err != nil {
				log.Err(err).Msgf("Handle message for consumer %s", l.consumer)
				msg.Nack(false, true) // requeue
			} else {
				msg.Ack(false) // ack the message
				log.Info().Msgf("Message success release for consumer %s", l.consumer)
			}
		}
	}
}

func (l *Listener) Close() error {
	if err := l.channel.Close(); err != nil {
		log.Err(err).Msgf("Close channel consumer %s", l.consumer)
		return err
	}
	log.Info().Msgf("Channel consumer %s closed.", l.consumer)
	return nil
}
