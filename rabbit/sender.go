package rabbit

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

// Sender - rabbit sender interface
type Sender interface {
	Send(data interface{}) error
}

type sender struct {
	conn  Connection
	queue Queue
}

// NewSender - return new rabbit web notifications sender
func NewSender(conn Connection, queue Queue) Sender {
	return &sender{
		conn:  conn,
		queue: queue,
	}
}

// Send - send alert message to web notifier
func (s *sender) Send(data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.conn.Channel().Value.Publish(
		"",
		s.queue.Value.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		},
	)
}
