package pkg

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer represents a RabbitMQ consumer for a specific queue.
type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

// NewConsumer creates a new Consumer.
func NewConsumer(conn *amqp.Connection, queueName string) *Consumer {
	return &Consumer{
		conn:      conn,
		queueName: queueName,
	}
}

// Listen starts consuming messages from the queue and calls the handler function for each message.
func (c *Consumer) Listen(handler func(amqp.Delivery)) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		c.queueName, // queue
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			handler(d)
		}
	}()

	Log.Info(fmt.Sprintf("Waiting for messages on queue: %s", c.queueName))
	<-forever

	return nil
}
