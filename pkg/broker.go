package pkg

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueHackathonRegistrations = "ai-hackathon-registrations"
	QueueWocRegistrations       = "woc-registrations"
)

var allQueues = []string{
	QueueWocRegistrations,
	QueueHackathonRegistrations,
}

var Rabbit *MsgBroker

type MsgBroker struct {
	conn    *amqp.Connection
	connURL string
	channel *amqp.Channel
}

func NewBroker(connStr string) (*MsgBroker, error) {
	client := &MsgBroker{
		connURL: connStr,
	}

	if err := client.connect(); err != nil {
		Log.Error("[BAD]: Message broker failed to initialize", err)
		return nil, err
	}

	return client, nil
}

func (r *MsgBroker) connect() error {
	var err error
	r.conn, err = amqp.Dial(r.connURL)
	if err != nil {
		return err
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		_ = r.conn.Close()
		return err
	}
	return nil
}

func (r *MsgBroker) declareQueues() error {
	for _, queueName := range allQueues {
		_, err := r.channel.QueueDeclare(
			queueName,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // args
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}
		Log.Info(fmt.Sprintf("Successfully delcared queue: %s", queueName))
	}

	Log.Info("All queues declared successfully.")
	return nil
}

func (r *MsgBroker) handleReconnect() {
	errChan := make(chan *amqp.Error)
	r.conn.NotifyClose(errChan)

	for err := range errChan {
		Log.Error("Broker connection lost. Attempting to reconnect...", err)

		for {
			time.Sleep(5 * time.Second)

			if err := r.connect(); err == nil {
				Log.Info("Successfully reconnected to message broker")
				r.conn.NotifyClose(errChan)
			}
			errChan = make(chan *amqp.Error)
			r.conn.NotifyClose(errChan)
			break
		}
		Log.Error("Failed to reconnect, retrying...", err)
	}
	// End of handleReconnect
}

func (r *MsgBroker) Consume() {
}

func (r *MsgBroker) Close() error {
	if r.channel == nil {
		return fmt.Errorf("no channels to rabbit-mq")
	}
	if r.conn == nil {
		return fmt.Errorf("no connection to rabbit-mq")
	}

	if err := r.channel.Close(); err != nil {
		return err
	}
	if err := r.conn.Close(); err != nil {
		return err
	}

	return nil
}

func (r *MsgBroker) Conn() *amqp.Connection {
	return r.conn
}
