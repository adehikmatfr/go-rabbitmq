package consumer

import (
	"adehikmatfr/go-rabbitmq/pkg/rabbitmq"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	client    *rabbitmq.Client
	errChan   chan error
	listeners []*rabbitmq.Listener
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

	// EXAMPLE DIRECT WITH DLX
	listener, err := r.client.NewListener(&rabbitmq.ListenerOpts{
		Consumer: "test-ade",
		Prefetch: 1,
		Queue: rabbitmq.QueueOpts{
			AutoAck: false,
			Declare: &rabbitmq.QueueDeclareOpts{
				AutoDelete: true,
				Args: amqp091.Table{
					"x-dead-letter-exchange":    "test-ade-dlx-exchange", // nama DLX
					"x-dead-letter-routing-key": "test-ade-dlx-key",      // opsional
					"x-message-ttl":             int32(10000),
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
		time.Sleep(10 * time.Second)
		return fmt.Errorf("ini biar masuk ke dlx")
	})

	r.listeners = append(r.listeners, listener)

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

	r.listeners = append(r.listeners, dlxListener)

	// EXAMPLE FANOUT
	fanoutListener, err := r.client.NewListener(&rabbitmq.ListenerOpts{
		Consumer: "fanout-consumer",
		Queue: rabbitmq.QueueOpts{
			Declare: &rabbitmq.QueueDeclareOpts{
				AutoDelete: true,
			},
			Bind: &rabbitmq.QueueBindOpts{
				Exchange: "fanout-exchange",
				// Fanout tidak butuh Key
			},
			Name: "fanout-queue",
		},
		Exchange: &rabbitmq.ExchangeDeclareOpts{
			Name:       "fanout-exchange",
			Kind:       rabbitmq.KindFanout,
			AutoDelete: true,
		},
	})

	if err != nil {
		log.Err(err).Msg("failed to create Fanout listener")
	}

	go fanoutListener.Listen(context.Background(), func(ctx context.Context, d *amqp091.Delivery) error {
		fmt.Printf("[Fanout] Received message: %s\n", string(d.Body))
		return nil
	})

	r.listeners = append(r.listeners, fanoutListener)

	// EXAMPLE TOPIC
	topicListener, err := r.client.NewListener(&rabbitmq.ListenerOpts{
		Consumer: "topic-consumer",
		Queue: rabbitmq.QueueOpts{
			Declare: &rabbitmq.QueueDeclareOpts{
				AutoDelete: true,
			},
			Bind: &rabbitmq.QueueBindOpts{
				Exchange: "topic-exchange",
				Key:      "order.*", // contoh routing key pattern
			},
			Name: "topic-queue",
		},
		Exchange: &rabbitmq.ExchangeDeclareOpts{
			Name:       "topic-exchange",
			Kind:       rabbitmq.KindTopic,
			AutoDelete: true,
		},
	})

	if err != nil {
		log.Err(err).Msg("failed to create Topic listener")
	}

	go topicListener.Listen(context.Background(), func(ctx context.Context, d *amqp091.Delivery) error {
		fmt.Printf("[Topic] Received message with routing key [%s]: %s\n", d.RoutingKey, string(d.Body))
		return nil
	})

	r.listeners = append(r.listeners, topicListener)

	// EXAMPLE HEADER
	headersListener, err := r.client.NewListener(&rabbitmq.ListenerOpts{
		Consumer: "headers-consumer",
		Queue: rabbitmq.QueueOpts{
			Declare: &rabbitmq.QueueDeclareOpts{
				AutoDelete: true,
			},
			Bind: &rabbitmq.QueueBindOpts{
				Exchange: "headers-exchange",
				Args: amqp091.Table{
					"x-match": "all", // "all" = AND, "any" = OR
					"type":    "report",
					"format":  "pdf",
				},
			},
			Name: "headers-queue",
		},
		Exchange: &rabbitmq.ExchangeDeclareOpts{
			Name:       "headers-exchange",
			Kind:       rabbitmq.KindHeaders,
			AutoDelete: true,
		},
	})

	if err != nil {
		log.Err(err).Msg("failed to create Headers listener")
	}

	go headersListener.Listen(context.Background(), func(ctx context.Context, d *amqp091.Delivery) error {
		fmt.Printf("[Headers] Received message: %s\n", string(d.Body))
		return nil
	})

	r.listeners = append(r.listeners, headersListener)
}

func (r *RabbitMQListener) Shutdown(ctx context.Context) error {
	if len(r.listeners) > 0 {
		for _, l := range r.listeners {
			err := l.Close()
			if err != nil {
				log.Err(err).Msg("failed to close listener")
			}
		}
	}
	return r.client.Close()
}
