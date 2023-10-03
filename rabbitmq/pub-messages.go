package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)



func PubMessagesQueue() error {

	if err != nil {
		return err
	}

	publisher, err2 := NewPublisher(channel, "public-messages")

}
