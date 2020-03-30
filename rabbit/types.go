package rabbit

import "github.com/streadway/amqp"

// Queue - represents amqp queue broker
type Queue struct {
	Value amqp.Queue
}

// Channel - represents amqp channel
type Channel struct {
	Value *amqp.Channel
}

// Delivery - represents rabbit delivery methods interface
type Delivery interface {
	Ack(multiple bool) error
}
