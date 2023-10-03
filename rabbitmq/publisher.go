package rabbitmq

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	Channel *amqp.Channel
	Queue   string
}

func NewPublisher(channel *amqp.Channel, queue string) (*Publisher, error) {

	publisher := &Publisher{Channel: channel}

	publisher.DeclareQueue(queue)

	_, err := channel.QueueDeclare(
		queue,
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

func (p *Publisher) DeclareQueue(q string) *Publisher {
	p.Queue = q
	return p
}

func (p *Publisher) NewPublish(body []byte, contentType string) error {

	if p.Queue == "" {
		return errors.New("no queue name has been specified")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err := p.Channel.PublishWithContext(ctx, "", p.Queue, false, false, amqp.Publishing{
		ContentType: contentType,
		Body:        body,
	})

	if err != nil {
		return err
	}
	return nil

}
