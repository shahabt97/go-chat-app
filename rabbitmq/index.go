package rabbitmq

import (
	// "fmt"
	"go-chat-app/config"

	// amqp "github.com/rabbitmq/amqp091-go"
	rabbit "github.com/shahabt97/rabbit-pool/v3"
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
	
	PubMessagesConsumer(consumer, publisher,pool)

	return nil
}

func Init() (*rabbit.ConnectionPool, error) {

	// conn, err := amqp.Dial(config.ConfigData.RabbitMQ)

	// if err != nil {
	// 	fmt.Println("error in connecting to rabbitMQ: ", err)
	// 	return nil, err
	// }

	pool := rabbit.NewPool(config.ConfigData.RabbitMQ, 1, 5, 4)

	return pool, nil

}
