package rabbitmq

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	rabbit "github.com/shahabt97/rabbit-pool"
)

type PubMessagePublishingMaster struct {
	Mongo   *rabbit.Publisher
	Elastic *rabbit.Publisher
	Redis   *rabbit.Publisher
}

type PubMessageConsumerMaster struct {
	Mongo   <-chan amqp.Delivery
	Elastic <-chan amqp.Delivery
	Redis   <-chan amqp.Delivery
}

var PubMessagePublishMaster = &PubMessagePublishingMaster{}
var PubMessageConsumeMaster = &PubMessageConsumerMaster{}

type MessageContent struct {
	Message   string    `json:"message"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
}

type Event struct {
	EventName string                 `json:"eventName"`
	Data      MessageContent         `json:"data"`
	// MongoId   *mongo.InsertOneResult `json:"-"`
}
