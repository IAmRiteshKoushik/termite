package consumer

import (
	"fmt"
	"math/rand"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rk/tentacloid/pkg"
)

// AiConsumer represents a consumer for the ai-hackathon-registrations queue.
type AiConsumer struct {
	conn *amqp.Connection
}

// NewAiConsumer creates a new AiConsumer.
func NewAiConsumer(conn *amqp.Connection) *AiConsumer {
	return &AiConsumer{
		conn: conn,
	}
}

// processMessage simulates a process that can fail and returns a boolean indicating success.
func (c *AiConsumer) processMessage(d amqp.Delivery) (bool, error) {
	pkg.Log.Info(fmt.Sprintf("Processing message: %s", string(d.Body)))

	// Simulate a process that might fail.
	if rand.Intn(10) < 7 { // 70% chance of failure
		pkg.Log.Warn(fmt.Sprintf("Failed to process message: %s. Retrying...", string(d.Body)))
		return false, nil
	}

	pkg.Log.Info(fmt.Sprintf("Successfully processed message: %s", string(d.Body)))
	return true, nil
}

// Listen starts consuming messages from the ai-hackathon-registrations queue.
func (c *AiConsumer) Listen() error {
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

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			pkg.Log.Info(fmt.Sprintf("Received a message from %s", q.Name))

			for {
				success, err := c.processMessage(d)
				if err != nil {
					pkg.Log.Error("Error processing message, will not retry", err)
					d.Ack(false)
					break
				}

				if success {
					pkg.Log.Info(fmt.Sprintf("Acknowledging message: %s", d.MessageId))
					d.Ack(false)
					break // Exit the retry loop
				} else {
					pkg.Log.Info("Retrying in 5 seconds...")
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()

	pkg.Log.Info(fmt.Sprintf("[*] Waiting for messages on %s. To exit press CTRL+C", q.Name))
	<-forever

	return nil
}

