package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IAmRiteshKoushik/tentacloid/pkg"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AiConsumer struct {
	conn *amqp.Connection
}

func NewAiConsumer(conn *amqp.Connection) *AiConsumer {
	return &AiConsumer{
		conn: conn,
	}
}

// This consumes the payload extracted from RabbitMQ, marshals it as JSON,
// and then dispatches it over HTTP.
func (c *AiConsumer) payloadDispatch(d amqp.Delivery) (bool, error) {
	var payload HackathonPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		// Cannot retry this error. Event has to be skipped. If this causes a
		// dispute, then it has to be handled manually.
		pkg.Log.Error("Failed to unmarshal message body", err)
		return false, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	err := DispatchHackathonPayload(payload)
	if err != nil {
		// Dispatch failures could be attributed to bad network conditions or
		// listener failures on the other end. Infinite retry setup needed.
		pkg.Log.Warn(fmt.Sprintf("Dispatch failed, will retry: %v", err))
		return false, nil
	}

	return true, nil
}

func (c *AiConsumer) Listen(ctx context.Context) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		pkg.QueueHackathonRegistrations, // name
		true,                            // durable
		false,                           // delete when unused
		false,                           // exclusive
		false,                           // no-wait
		nil,                             // arguments
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
			pkg.Log.Info("Shutting down AI consumer...")
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
					d.Nack(false, true) // Not acknowledging the message
					return nil
				default:
					// Continue processing
				}

				success, err := c.payloadDispatch(d)
				if err != nil {
					pkg.Log.Error("Error processing message, will not retry", err)
					d.Ack(false)
					break // Exit retry loop
				}

				if success {
					pkg.Log.Info(fmt.Sprintf("Acknowledging message: %s", d.MessageId))
					d.Ack(false)
					break // Exit retry loop
				}

				pkg.Log.Info("Retrying in 5 seconds...")

				select {
				case <-time.After(5 * time.Second):
				case <-ctx.Done():
					pkg.Log.Info("Shutdown signal received during retry sleep. Nacking message.")
					d.Nack(false, true) // Not acknowledging message.
					return nil
				}

			} // end of for
		}
	}
}
