package rabbit

import (
	"time"

	"github.com/lillilli/logger"
	"github.com/streadway/amqp"
)

// QueueSubscriber - represents queue subscriber
type QueueSubscriber interface {
	SubscribeOnQueue() (<-chan amqp.Delivery, error)
}

type subscriber struct {
	queue Queue
	conn  Connection

	externalChannel chan amqp.Delivery

	log logger.Logger
}

// NewQueueSubscriber - returns new queue subscriber instance
func NewQueueSubscriber(conn Connection, queue Queue) (QueueSubscriber, error) {
	return &subscriber{
		queue:           queue,
		conn:            conn,
		externalChannel: make(chan amqp.Delivery),
		log:             logger.NewLogger("amqp subscriber"),
	}, nil
}

func (s *subscriber) Subscribe(queue Queue) (<-chan amqp.Delivery, error) {
	return s.conn.Channel().Value.Consume(
		queue.Value.Name, // queue
		"",               // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
}

func (s *subscriber) SubscribeOnQueue() (<-chan amqp.Delivery, error) {
	ch, err := s.conn.Channel().Value.Consume(
		s.queue.Value.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)

	if err != nil {
		return ch, err
	}

	go s.watchForBroke()
	go s.startProxyEvents(ch)
	return s.externalChannel, nil
}

func (s *subscriber) watchForBroke() {
	for {
		reason := <-s.conn.Channel().Value.NotifyClose(make(chan *amqp.Error))

		for {
			s.log.Infof("Channel closed, reason: %q, try to recreate channel", reason)
			time.Sleep(ReconnectInterval)

			queue, err := s.conn.DeclareQueue(s.queue.Value.Name)
			if err != nil {
				s.log.Errorf("Recreating queue failed: %v", err)
				continue
			}

			s.queue = queue

			newInternalChannel, err := s.conn.Channel().Value.Consume(
				s.queue.Value.Name, // queue
				"",                 // consumer
				true,               // auto-ack
				false,              // exclusive
				false,              // no-local
				false,              // no-wait
				nil,                // args
			)

			if err == nil {
				s.log.Infof("Consumer restarting success complete")
				go s.startProxyEvents(newInternalChannel)
				break
			}

			s.log.Errorf("Consumer restarting failed: %v", err)
		}
	}
}

func (s *subscriber) startProxyEvents(ch <-chan amqp.Delivery) {
	for data := range ch {
		s.externalChannel <- data
	}
}
