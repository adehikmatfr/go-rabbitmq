package consumer

import (
	"adehikmatfr/go-rabbitmq/pkg/rabbitmq"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type RabbitMQListenerService interface {
	RunListener()
	TerminateSignal() chan os.Signal
	Shutdown(ctx context.Context) error
	ListenError() <-chan error
}

type RabbitMQListenerOptions struct {
	URL string
}

type RabbitMQListener struct {
	client  *rabbitmq.Client
	errChan chan error
}

func NewRabbitMQListener(opts *RabbitMQListenerOptions) RabbitMQListenerService {
	client, err := rabbitmq.NewClient(opts.URL)
	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
		return nil
	}

	return &RabbitMQListener{
		client:  client,
		errChan: make(chan error, 10),
	}
}

func (r *RabbitMQListener) TerminateSignal() chan os.Signal {
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	return term
}

func (r *RabbitMQListener) ListenError() <-chan error {
	return r.errChan
}

func (r *RabbitMQListener) RunListener() {
	fmt.Println("=======================================")
	fmt.Println("    R A B I T M Q   L I S T E N E R    ")
	fmt.Println("=======================================")

	listener, err := r.client.NewListener(&rabbitmq.ListenerOpts{
		Consumer: "test-ade",
		Queue: rabbitmq.QueueOpts{
			Declare: &rabbitmq.QueueDeclareOpts{
				AutoDelete: true,
				Args: amqp091.Table{
					"x-dead-letter-exchange":    "test-ade-dlx-exchange", // nama DLX
					"x-dead-letter-routing-key": "test-ade-dlx-key",      // opsional
					"x-message-ttl":             int32(50000),            // TTL 10 detik (opsional)
				},
			},
			Bind: &rabbitmq.QueueBindOpts{
				Key:      "test-ade-bind",
				Exchange: "test-ade-exchange",
			},
			Name: "test-ade-queue",
		},
		Exchange: &rabbitmq.ExchangeDeclareOpts{
			Name:       "test-ade-exchange",
			Kind:       rabbitmq.KindDirect,
			AutoDelete: true,
		},
	})

	if err != nil {
		log.Err(err).Msg(err.Error())
	}

	go listener.Listen(context.Background(), func(ctx context.Context, d *amqp091.Delivery) error {
		fmt.Println(string(d.Body))
		return fmt.Errorf("error")
	})

	dlxListener, err := r.client.NewListener(&rabbitmq.ListenerOpts{
		Consumer: "test-ade-dlx-consumer",
		Queue: rabbitmq.QueueOpts{
			Declare: &rabbitmq.QueueDeclareOpts{
				AutoDelete: true,
			},
			Bind: &rabbitmq.QueueBindOpts{
				Key:      "test-ade-dlx-key",
				Exchange: "test-ade-dlx-exchange",
			},
			Name: "test-ade-dlx-queue",
		},
		Exchange: &rabbitmq.ExchangeDeclareOpts{
			Name:       "test-ade-dlx-exchange",
			Kind:       rabbitmq.KindDirect,
			AutoDelete: true,
		},
	})

	if err != nil {
		log.Err(err).Msg("failed to create DLX listener")
	}

	go dlxListener.Listen(context.Background(), func(ctx context.Context, d *amqp091.Delivery) error {
		fmt.Printf("[DLX] Received dead letter message: %s\n", string(d.Body))
		// Acknowledge it or send to another queue
		return nil
	})
}

func (r *RabbitMQListener) Shutdown(ctx context.Context) error {
	return r.client.Close()
}
