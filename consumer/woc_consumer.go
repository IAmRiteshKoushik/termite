package consumer

import (
	"fmt"
	"math/rand"
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

// processMessage simulates a process that can fail and returns a boolean indicating success.
// In a real application, this function would contain the business logic to process the message.
func (c *WocConsumer) processMessage(d amqp.Delivery) (bool, error) {
	pkg.Log.Info(fmt.Sprintf("Processing message: %s", string(d.Body)))

	// Simulate a process that might fail.
	// For demonstration, we'll use a random number to simulate success or failure.
	// In a real-world scenario, this would be your actual business logic.
	if rand.Intn(10) < 7 { // 70% chance of failure
		pkg.Log.Warn(fmt.Sprintf("Failed to process message: %s. Retrying...", string(d.Body)))
		return false, nil
	}

	pkg.Log.Info(fmt.Sprintf("Successfully processed message: %s", string(d.Body)))
	return true, nil
}

// Listen starts consuming messages from the woc-registrations queue.
func (c *WocConsumer) Listen() error {
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

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			pkg.Log.Info(fmt.Sprintf("Received a message from %s", q.Name))

			for {
				success, err := c.processMessage(d)
				if err != nil {
					pkg.Log.Error("Error processing message, will not retry", err)
					// For critical errors, you might want to move the message to a dead-letter queue
					// or just acknowledge it to remove it from the queue.
					// For this example, we'll just acknowledge it.
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

