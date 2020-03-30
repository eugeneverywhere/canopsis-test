package rabbit

import (
	"sync"
	"time"

	"github.com/lillilli/logger"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

const ReconnectInterval = time.Second

// Connection - represents amqp connection interface
type Connection interface {
	// DeclareQueue - declare amqp queue
	DeclareQueue(queue string) (Queue, error)
	// Do not use this method in sample cases, it returns connection without reconnecting algorithm.
	Channel() Channel
	// Close - close connection
	Close() error
}

type connection struct {
	addr    string
	conn    *amqp.Connection
	channel *amqp.Channel
	closed  bool
	log     logger.Logger
	sync.Mutex
}

// NewConnection - returns new amqp connection
func NewConnection(addr string) (Connection, error) {
	connection := new(connection)

	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect")
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to rabbit channel")
	}

	connection.addr = addr
	connection.conn = conn
	connection.channel = channel
	connection.log = logger.NewLogger("rabbit connection")

	go connection.watchForBroke()
	return connection, nil
}

func (c *connection) watchForBroke() {
	for {
		reason, ok := <-c.channel.NotifyClose(make(chan *amqp.Error))

		c.Lock()
		closed := c.closed
		c.Unlock()

		if (!ok || c.conn.IsClosed()) && closed {
			c.channel.Close() // close again, ensure closed flag set when connection closed
			break
		}

		for {
			c.log.Infof("Channel closed, reason: %q, try to recreate channel", reason)
			time.Sleep(ReconnectInterval)

			if c.conn.IsClosed() {
				conn, err := amqp.Dial(c.addr)
				if err != nil {
					c.log.Errorf("Connection recreating failed: %v", err)
					continue
				}

				c.log.Infof("Connection recreating success complete")
				c.conn = conn
			}

			channel, err := c.conn.Channel()
			if err == nil {
				c.log.Infof("Channel recreating success complete")
				c.channel = channel
				break
			}

			c.log.Errorf("Channel recreating failed: %v", err)
		}
	}
}

func (c *connection) DeclareQueue(queue string) (Queue, error) {
	v, err := c.channel.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	return Queue{Value: v}, err
}

func (c *connection) Channel() Channel {
	return Channel{Value: c.channel}
}

func (c *connection) Close() error {
	c.Lock()
	c.closed = true
	c.Unlock()

	if err := c.conn.Close(); err != nil {
		return err
	}

	return c.channel.Close()
}
