package rabbitmq

import (
	"fmt"
	"go-chat-app/config"

	amqp "github.com/rabbitmq/amqp091-go"
	rabbit "github.com/shahabt97/rabbit-pool"
)

func RabbitMQInitialization(publisher *PubMessagePublishingMaster, consumer *PubMessageConsumerMaster) error {

	pool, err := Init()
	if err != nil {
		return err
	}

	err = PubMessageQueueHandler(publisher, consumer, pool)
	if err != nil {
		return err
	}

	PubMessagesConsumer(consumer, publisher)

	return nil
}

func Init() (*rabbit.ConnectionPool, error) {

	conn, err := amqp.Dial(config.ConfigData.RabbitMQ)

	if err != nil {
		fmt.Println("error in connecting to rabbitMQ: ", err)
		return nil, err
	}

	pool := rabbit.NewPool(conn, 15, 10)

	return pool, nil

}
