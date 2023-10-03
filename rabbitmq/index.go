package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

var Channel *amqp.Channel

func RabbiInitialization() error {

	var ErrOfConnectingRabbit error

	Channel, ErrOfConnectingRabbit = Init()
	if ErrOfConnectingRabbit != nil {
		return ErrOfConnectingRabbit
	}
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
