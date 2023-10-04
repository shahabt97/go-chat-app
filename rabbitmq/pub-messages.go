package rabbitmq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-chat-app/database"
	"go-chat-app/elasticsearch"
	redisServer "go-chat-app/redis"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func PubMessagesPublisher(jsonData *Event, master *PubMessagePublishingMaster) error {

	p, err1 := json.Marshal(jsonData)
	if err1 != nil {
		return err1
	}

	err2 := master.Mongo.NewPublish(p, "application/json")
	if err2 != nil {
		return err2
	}

	err4 := master.Redis.NewPublish(p, "application/json")
	if err4 != nil {
		return err4
	}

	return nil
}

func PubMessagesConsumer(master *PubMessageConsumerMaster, publishMaster *PubMessagePublishingMaster) {

	go PubMessagesInMongoConsumer(publishMaster, master)
	go PubMesInRedisConsumer(master)
	go PubMesInESConsumer(master)

}

func PubMessageQueueHandler(publisher *PubMessagePublishingMaster, consumer *PubMessageConsumerMaster, ch *amqp.Channel) error {

	var err error

	publisher.Mongo, err = NewPublisher(ch, "public-messages-mongo")
	if err != nil {
		return err
	}
	publisher.Elastic, err = NewPublisher(ch, "public-messages-elastic")
	if err != nil {
		return err
	}
	publisher.Redis, err = NewPublisher(ch, "public-messages-redis")
	if err != nil {
		return err
	}

	consumer.Mongo, err = NewConsumer(ch, "public-messages-mongo").NewDeliveryChannel()
	if err != nil {
		return err
	}
	consumer.Elastic, err = NewConsumer(ch, "public-messages-elastic").NewDeliveryChannel()
	if err != nil {
		return err
	}
	consumer.Redis, err = NewConsumer(ch, "public-messages-redis").NewDeliveryChannel()
	if err != nil {
		return err
	}

	return nil

}

func PubMesInESConsumer(consumer *PubMessageConsumerMaster) {

	for message := range consumer.Redis {
		pubReader := bytes.NewReader(message.Body)
		errPubElas := elasticsearch.Client.CreateDoc("pubmessages", pubReader)
		if errPubElas != nil {
			fmt.Println("Error in creating user in elastic: ", errPubElas)
			continue
		}
		message.Ack(false)
	}

}

func PubMesInRedisConsumer(consumer *PubMessageConsumerMaster) {
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for message := range consumer.Redis {

		var jsonData Event

		errOfUnMarshling := json.Unmarshal(message.Body, &jsonData)

		if errOfUnMarshling != nil {
			continue
		}

		newDoc := &database.PublicMessage{
			Message:   jsonData.Data.Message,
			Sender:    jsonData.Data.Username,
			CreatedAt: jsonData.Data.Timestamp,
		}

		var Array = []*database.PublicMessage{}

		val, errOfRedis := redisServer.Client.Client.Get(ctx, "pubmessages").Bytes()

		if errOfRedis != nil {
			fmt.Println("error in getting data from Redis: ", errOfRedis)
			continue
		}

		errOfUnMar := json.Unmarshal(val, &Array)
		if errOfUnMar != nil {
			fmt.Println("error in unmarshling: ", errOfUnMar)
			continue
		}

		Array = append(Array, newDoc)

		err := redisServer.Client.SetPubMes(&Array)
		if err != nil {
			fmt.Printf("error in setting new pub messagein Redis: %v\n", err)
			continue
		}
		message.Ack(false)
	}

}

func PubMessagesInMongoConsumer(publisher *PubMessagePublishingMaster, consumer *PubMessageConsumerMaster) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for message := range consumer.Mongo {

		var jsonData Event

		errOfUnMarshling := json.Unmarshal(message.Body, &jsonData)

		if errOfUnMarshling != nil {
			continue
		}

		session, errInSession := database.Client.StartSession()

		if errInSession != nil {
			fmt.Printf("error in creating session: %v \n", errInSession)
			continue
		}

		_, err := session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {

			result, errOfMongo := database.PubMessages.InsertOne(ctx, bson.D{
				{Key: "message", Value: jsonData.Data.Message},
				{Key: "sender", Value: jsonData.Data.Username},
				{Key: "createdAt", Value: jsonData.Data.Timestamp},
			})

			if errOfMongo != nil {
				fmt.Println("error storing new public message in Mongo: ", errOfMongo)
				return nil, errOfMongo
			}

			pubMessageJson := &elasticsearch.PubMessageIndex{
				Id:      result.InsertedID.(primitive.ObjectID).Hex(),
				Message: jsonData.Data.Message,
			}

			p, errOfMarshaling := json.Marshal(pubMessageJson)

			if errOfMarshaling != nil {
				return nil, errOfMarshaling
			}

			errOfPublishingToElastic := publisher.Elastic.NewPublish(p, "application/json")

			if errOfPublishingToElastic != nil {
				return nil, errOfPublishingToElastic
			}
			return nil, nil

		})

		if err != nil {
			continue
		}

		message.Ack(false)

	}

}
