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

	rabbit "github.com/shahabt97/rabbit-pool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PubMessagesPublisher(jsonData *Event, master *PubMessagePublishingMaster) error {

	p, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	err = master.Mongo.NewPublish(p, "application/json")
	if err != nil {
		return err
	}

	err = master.Redis.NewPublish(p, "application/json")
	if err != nil {
		return err
	}

	return nil
}

func PubMessagesConsumer(master *PubMessageConsumerMaster, publishMaster *PubMessagePublishingMaster) {

	go PubMessagesInMongoConsumer(publishMaster, master)
	go PubMesInRedisConsumer(master)
	go PubMesInESConsumer(master)

}

func PubMessageQueueHandler(publisher *PubMessagePublishingMaster, consumer *PubMessageConsumerMaster, pool *rabbit.ConnectionPool) (err error) {

	publisher.Mongo, err = rabbit.NewPublisher(pool, "public-messages-mongo")
	if err != nil {
		return
	}
	publisher.Elastic, err = rabbit.NewPublisher(pool, "public-messages-elastic")
	if err != nil {
		return
	}
	publisher.Redis, err = rabbit.NewPublisher(pool, "public-messages-redis")
	if err != nil {
		return
	}

	consumer.Mongo, err = rabbit.NewConsumer(pool, "public-messages-mongo")
	if err != nil {
		return
	}
	consumer.Elastic, err = rabbit.NewConsumer(pool, "public-messages-elastic")
	if err != nil {
		return
	}
	consumer.Redis, err = rabbit.NewConsumer(pool, "public-messages-redis")
	if err != nil {
		return
	}

	return

}

func PubMesInESConsumer(consumer *PubMessageConsumerMaster) {

	for message := range consumer.Elastic {

		pubReader := bytes.NewReader(message.Body)
		err := elasticsearch.Client.CreateDoc("pubmessages", pubReader)
		if err != nil {
			fmt.Println("Error in creating user in elastic: ", err)
			continue
		}
		message.Ack(false)
	}
	fmt.Printf("elastic consumer channel was closed. trying to reconnect\n")
	go PubMesInESConsumer(consumer)
	return
}

func PubMesInRedisConsumer(consumer *PubMessageConsumerMaster) {

	for message := range consumer.Redis {

		var jsonData Event

		errOfUnMarshling := json.Unmarshal(message.Body, &jsonData)

		if errOfUnMarshling != nil {
			continue
		}

		newMessage := &database.PublicMessage{
			Message:   jsonData.Data.Message,
			Sender:    jsonData.Data.Username,
			CreatedAt: jsonData.Data.Timestamp,
		}

		var Array []*database.PublicMessage

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		val, err := redisServer.Client.Client.Get(ctx, "pubmessages").Result()
		cancel()

		if err != nil {
			// fmt.Println("error in getting data from Redis: ", err)
			continue
		}

		err = json.Unmarshal([]byte(val), &Array)
		if err != nil {
			fmt.Println("error in unmarshling: ", err)
			continue
		}

		Array = append(Array, newMessage)

		err = redisServer.Client.SetPubMes(&Array)
		if err != nil {
			fmt.Printf("error in setting new pub message in Redis: %v\n", err)
			continue
		}
		message.Ack(false)
	}
	fmt.Printf("redis consumer channel was closed. trying to reconnect\n")
	go PubMesInRedisConsumer(consumer)
	return

}

func PubMessagesInMongoConsumer(publisher *PubMessagePublishingMaster, consumer *PubMessageConsumerMaster) {

	for message := range consumer.Mongo {

		var jsonData Event

		err := json.Unmarshal(message.Body, &jsonData)

		if err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		result, err := database.PubMessages.InsertOne(ctx, bson.D{
			{Key: "message", Value: jsonData.Data.Message},
			{Key: "sender", Value: jsonData.Data.Username},
			{Key: "createdAt", Value: jsonData.Data.Timestamp},
		})
		cancel()

		if err != nil {
			fmt.Println("error storing new public message in Mongo: ", err)
			continue
		}
		message.Ack(false)

		pubMessageJson := &elasticsearch.PubMessageIndex{
			Id:      result.InsertedID.(primitive.ObjectID).Hex(),
			Message: jsonData.Data.Message,
		}

		p, err := json.Marshal(pubMessageJson)

		if err != nil {
			fmt.Println("error in marshaling new public message: ", err)
			continue
		}

		err = publisher.Elastic.NewPublish(p, "application/json")

		if err != nil {
			fmt.Println("error in publishing new public message to elastic: ", err)
		}
	}
	fmt.Printf("mongo consumer channel was closed. trying to reconnect\n")
	go PubMessagesInMongoConsumer(publisher,consumer)
	return
}
