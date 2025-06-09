package rabbitmq

import (
	"context"
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

// mock amqp dial
func mockAMQPDial(conn *amqp.Connection, retErr error) func(string) (*amqp.Connection, error) {
	return func(url string) (*amqp.Connection, error) {
		if retErr != nil {
			return nil, retErr
		}
		return conn, retErr
	}
}

// mock dial channel
func mockDialChannel(conn *amqp.Connection, retErr error) func(conn *amqp.Connection) (*amqp.Channel, error) {
	return func(conn *amqp.Connection) (*amqp.Channel, error) {
		if retErr != nil {
			return nil, retErr
		}
		return &amqp.Channel{}, nil
	}
}

// mock publish with context
func mockPublishWithContext(ch *amqp.Channel, ctx context.Context, exchange, key string, mandatory, immediate bool, pub amqp.Publishing, retErr error) func(ch *amqp.Channel, ctx context.Context, exchange, key string, mandatory, immediate bool, pub amqp.Publishing) error {
	return func(ch *amqp.Channel, ctx context.Context, exchange, key string, mandatory, immediate bool, pub amqp.Publishing) error {
		if retErr != nil {
			return retErr
		}
		return nil
	}
}

// mock close channel
func mockCloseChannel(ch *amqp.Channel, retErr error) func(ch *amqp.Channel) error {
	return func(ch *amqp.Channel) error {
		if retErr != nil {
			return retErr
		}
		return nil
	}
}

type moduleTest struct {
	amqpMock func(conn *amqp.Connection, retErr error) func(string) (*amqp.Connection, error)
	mockUrl  string
	mockConn *amqp.Connection
	// mock publisher
	mockDialChannel        func(conn *amqp.Connection, retErr error) func(conn *amqp.Connection) (*amqp.Channel, error)
	mockPublishWithContext func(ch *amqp.Channel, ctx context.Context, exchange, key string, mandatory, immediate bool, pub amqp.Publishing, retErr error) func(ch *amqp.Channel, ctx context.Context, exchange, key string, mandatory, immediate bool, pub amqp.Publishing) error
	mockCloseChannel       func(ch *amqp.Channel, retErr error) func(ch *amqp.Channel) error
}

type TestCase struct {
	name string
	test func(obj *moduleTest)
}

func cleanup() {
	amqpDial = amqp.Dial
	dialChannel = func(conn *amqp.Connection) (*amqp.Channel, error) { return conn.Channel() }
	publishWithContext = func(ch *amqp.Channel, ctx context.Context, exchange, key string, mandatory, immediate bool, pub amqp.Publishing) error {
		return ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, pub)
	}
	closeChannel = func(ch *amqp.Channel) error { return ch.Close() }
}

func doTestModule(t *testing.T, fn func(*moduleTest)) {
	// t.Parallel() // NOTE: not use this becouse amqpDial is a global variable
	_ = t
	module := moduleTest{
		amqpMock:               mockAMQPDial,
		mockUrl:                "amqp://localhost:5672/",
		mockConn:               &amqp.Connection{},
		mockDialChannel:        mockDialChannel,
		mockPublishWithContext: mockPublishWithContext,
		mockCloseChannel:       mockCloseChannel,
	}
	// cleanup after runing set to default dial
	defer cleanup()
	fn(&module)
}

func run(t *testing.T, testCases []TestCase) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doTestModule(t, tc.test)
		})
	}
}

func TestNewClient(t *testing.T) {
	doTestModule(t, func(module *moduleTest) {
		instance, err := NewClient(module.mockUrl)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestClientTestCases(t *testing.T) {
	run(t, []TestCase{
		{
			name: "SuccessOnFirstTry",
			test: func(obj *moduleTest) {
				amqpDial = obj.amqpMock(obj.mockConn, nil)
				client, err := NewClient(obj.mockUrl)
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, obj.mockConn, client.conn)
			},
		},
		{
			name: "RetrySuccess",
			test: func(obj *moduleTest) {
				err := errors.New("connection refused")
				// Simulate success on third attempt
				attempts := 0
				amqpDial = func(url string) (*amqp.Connection, error) {
					attempts++
					if attempts < 3 {
						return nil, err
					}
					return obj.mockConn, nil
				}

				client, err := NewClient(obj.mockUrl)
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, 3, attempts)
			},
		},
		{
			name: "AllRetriesFail",
			test: func(obj *moduleTest) {
				err := errors.New("failed connection")
				attempts := 1
				amqpDial = func(url string) (*amqp.Connection, error) {
					attempts++
					if attempts == 4 {
						return nil, err
					}
					return obj.mockConn, err
				}

				client, err := NewClient(obj.mockUrl)
				assert.Error(t, err)
				assert.Nil(t, client)
			},
		},
	})
}

func TestNewPublisher(t *testing.T) {
	doTestModule(t, func(module *moduleTest) {
		client, err := NewClient(module.mockUrl)
		assert.NotNil(t, client)
		assert.Nil(t, err)
		instance, err := client.NewPublisher(PublisherOpts{})
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestPublisherTestCases(t *testing.T) {
	run(t, []TestCase{
		{
			name: "Success",
			test: func(obj *moduleTest) {
				amqpDial = obj.amqpMock(obj.mockConn, nil)
				client, _ := NewClient(obj.mockUrl)

				dialChannel = obj.mockDialChannel(obj.mockConn, nil)
				publisher, err := client.NewPublisher(PublisherOpts{
					Exchange:    "test-exchange",
					ContentType: "application/json",
				})

				assert.NoError(t, err)
				assert.NotNil(t, publisher)
				assert.NotNil(t, publisher.channel)
				assert.Equal(t, "test-exchange", publisher.exchange)
				assert.Equal(t, "application/json", publisher.contentType)
			},
		},
		{
			name: "ChannelError",
			test: func(obj *moduleTest) {
				amqpDial = obj.amqpMock(obj.mockConn, nil)
				client, _ := NewClient(obj.mockUrl)

				dialChannel = obj.mockDialChannel(obj.mockConn, errors.New("channel failed"))
				publisher, err := client.NewPublisher(PublisherOpts{})

				assert.Error(t, err)
				assert.Nil(t, publisher)
			},
		},
		{
			name: "PublishSuccess",
			test: func(obj *moduleTest) {
				amqpDial = obj.amqpMock(obj.mockConn, nil)
				client, _ := NewClient(obj.mockUrl)

				dialChannel = obj.mockDialChannel(obj.mockConn, nil)
				publisher, _ := client.NewPublisher(PublisherOpts{
					Exchange:    "test-exchange",
					ContentType: "text/plain",
				})

				ctx := context.Background()
				routingKey := "key"
				body := []byte("hello")

				publishWithContext = obj.mockPublishWithContext(
					publisher.channel,
					ctx,
					publisher.exchange,
					routingKey,
					publisher.mandatory,
					publisher.immediate,
					amqp.Publishing{
						ContentType: publisher.contentType,
						Body:        body,
						Headers:     publisher.headers,
						Priority:    publisher.priority,
						Expiration:  publisher.expiration,
					},
					nil,
				)

				err := publisher.Publish(ctx, routingKey, body)
				assert.NoError(t, err)
			},
		},
		{
			name: "PublishError",
			test: func(obj *moduleTest) {
				amqpDial = obj.amqpMock(obj.mockConn, nil)
				client, _ := NewClient(obj.mockUrl)

				dialChannel = obj.mockDialChannel(obj.mockConn, nil)
				publisher, _ := client.NewPublisher(PublisherOpts{
					Exchange:    "test-exchange",
					ContentType: "text/plain",
				})

				ctx := context.Background()
				routingKey := "key"
				body := []byte("hello")

				publishWithContext = obj.mockPublishWithContext(
					publisher.channel,
					ctx,
					publisher.exchange,
					routingKey,
					publisher.mandatory,
					publisher.immediate,
					amqp.Publishing{
						ContentType: publisher.contentType,
						Body:        body,
						Headers:     publisher.headers,
						Priority:    publisher.priority,
						Expiration:  publisher.expiration,
					},
					errors.New("publish failed"),
				)

				err := publisher.Publish(ctx, routingKey, body)
				assert.Error(t, err)
			},
		},
		{
			name: "CloseSuccess",
			test: func(obj *moduleTest) {
				amqpDial = obj.amqpMock(obj.mockConn, nil)
				client, _ := NewClient(obj.mockUrl)

				dialChannel = obj.mockDialChannel(obj.mockConn, nil)
				publisher, _ := client.NewPublisher(PublisherOpts{
					Exchange:    "test-exchange",
					ContentType: "text/plain",
				})

				closeChannel = obj.mockCloseChannel(publisher.channel, nil)
				err := publisher.Close()

				assert.NoError(t, err)
			},
		},
		{
			name: "CloseError",
			test: func(obj *moduleTest) {
				amqpDial = obj.amqpMock(obj.mockConn, nil)
				client, _ := NewClient(obj.mockUrl)

				dialChannel = obj.mockDialChannel(obj.mockConn, nil)
				publisher, _ := client.NewPublisher(PublisherOpts{
					Exchange:    "test-exchange",
					ContentType: "text/plain",
				})

				closeChannel = obj.mockCloseChannel(publisher.channel, errors.New("close failed"))
				err := publisher.Close()

				assert.Error(t, err)
			},
		},
	})
}
