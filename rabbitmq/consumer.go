package rabbitmq

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Channel *amqp.Channel
	Queue   string
}

func NewConsumer(channel *amqp.Channel, queue string) *Consumer {
	consumer := &Consumer{Channel: channel}
	consumer.DeclareQueue(queue)
	return consumer
}

func (c *Consumer) DeclareQueue(q string) *Consumer {
	c.Queue = q
	return c
}

func (c *Consumer) NewDeliveryChannel() (<-chan amqp.Delivery, error) {

	if c.Queue == "" {
		return nil, errors.New("no queue name has been specified")
	}

	deliveryChannel, err := c.Channel.Consume(c.Queue, "", false, false, false, false, nil)

	if err != nil {
		return nil, err
	}

	return deliveryChannel, nil

}
