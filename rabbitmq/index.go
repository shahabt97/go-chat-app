package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func RabbitMQInitialization(publisher *PubMessagePublishingMaster, consumer *PubMessageConsumerMaster) error {

	channel, errOfConnectingRabbit := Init()
	if errOfConnectingRabbit != nil {
		return errOfConnectingRabbit
	}

	err := PubMessageQueueHandler(publisher, consumer, channel)
	if err != nil {
		return err
	}

	PubMessagesConsumer(consumer, publisher)

	return nil
}

func Init() (*amqp.Channel, error) {

	conn, err := amqp.Dial("amqp://shahab:83000000@localhost:5672/chat-application")

	if err != nil {
		fmt.Println("error in connecting to rabbitMQ: ", err)
		return nil, err
	}

	ch, _ := conn.Channel()

	return ch, nil

}
