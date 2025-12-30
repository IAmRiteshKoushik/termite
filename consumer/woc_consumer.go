package consumer

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rk/tentacloid/pkg"
)

// WocConsumer represents a consumer for the woc-registrations queue.
type WocConsumer struct {
	conn *amqp.Connection
}

// NewWocConsumer creates a new WocConsumer.
func NewWocConsumer(conn *amqp.Connection) *WocConsumer {
	return &WocConsumer{
		conn: conn,
	}
}

// Listen starts consuming messages from the woc-registrations queue.
func (c *WocConsumer) Listen(ctx context.Context) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		pkg.QueueWocRegistrations, // name
		true,                      // durable
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	pkg.Log.Info(fmt.Sprintf("[*] Waiting for messages on %s.", q.Name))

	for {
		select {
		case <-ctx.Done():
			pkg.Log.Info("Shutting down WOC consumer...")
			return nil
		case d, ok := <-msgs:
			if !ok {
				pkg.Log.Info("Message channel closed by RabbitMQ.")
				return nil
			}
			pkg.Log.Info(fmt.Sprintf("Received a message from %s", q.Name))

			for {
				// Inner loop for retries
				select {
				case <-ctx.Done():
					pkg.Log.Info("Shutdown signal received during message processing. Nacking message.")
					d.Nack(false, true) // Re-queue the message
					return nil
				default:
					// Continue processing
				}

				success, err := c.processMessage(d)
				if err != nil {
					pkg.Log.Error("Error processing message, will not retry", err)
					d.Ack(false)
					break // Exit retry loop
				}

				if success {
					pkg.Log.Info(fmt.Sprintf("Acknowledging message: %s", d.MessageId))
					d.Ack(false)
					break // Exit retry loop
				} else {
					pkg.Log.Info("Retrying in 5 seconds...")
					// Use a select to avoid blocking the shutdown signal during sleep
					select {
					case <-time.After(5 * time.Second):
						// Continue to next retry
					case <-ctx.Done():
						pkg.Log.Info("Shutdown signal received during retry sleep. Nacking message.")
						d.Nack(false, true) // Re-queue the message
						return nil
					}
				}
			}
		}
	}
}
