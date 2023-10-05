package websocketServer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-chat-app/database"
	"go-chat-app/elasticsearch"
	redisServer "go-chat-app/redis"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleNewPvMes(messageData *PvEvent, host string) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := database.PvMessages.InsertOne(ctx, bson.D{
		{Key: "message", Value: messageData.Data.Message},
		{Key: "sender", Value: messageData.Data.Sender},
		{Key: "receiver", Value: messageData.Data.Receiver},
		{Key: "createdAt", Value: messageData.Data.Timestamp},
	})

	if err != nil {
		fmt.Println("error in writing pv message in Mongo: ", err)

		// after 10 seconds again try to insert data in Mongo
		time.Sleep(10 * time.Second)
		go HandleNewPvMes(messageData, host)
		return
	}

	// save pv message in elastic
	go SavePvMesInES(result, messageData)

	// save pv message in Redis
	go SavePvMesInRedis(messageData, host)

}

func SavePvMesInES(result *mongo.InsertOneResult, messageData *PvEvent) {

	pvMessageJson := &elasticsearch.PvMessageIndex{
		Id:       result.InsertedID.(primitive.ObjectID).Hex(),
		Message:  messageData.Data.Message,
		Sender:   messageData.Data.Sender,
		Receiver: messageData.Data.Receiver,
	}

	pvJsonBytes, errorOfMar := json.Marshal(pvMessageJson)
	if errorOfMar != nil {
		fmt.Println("Error in Marshaling user data for elastic: ", errorOfMar)

		// after 10 seconds again try to insert data in Elastic
		time.Sleep(10 * time.Second)
		go SavePvMesInES(result, messageData)
		return
	}

	pvReader := bytes.NewReader(pvJsonBytes)
	errPvElas := elasticsearch.Client.CreateDoc("pv-messages", pvReader)
	if errPvElas != nil {
		fmt.Println("Error in creating user in elastic: ", errPvElas)
		time.Sleep(10 * time.Second)
		go SavePvMesInES(result, messageData)
		return
	}

}

func SavePvMesInRedis(messageData *PvEvent, host string) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newDoc := &database.PvMessage{
		Message:   messageData.Data.Message,
		Sender:    messageData.Data.Sender,
		Receiver:  messageData.Data.Receiver,
		CreatedAt: messageData.Data.Timestamp,
	}

	var Array = []*database.PvMessage{}

	username := messageData.Data.Sender

	val, errOfRedis := redisServer.Client.Client.MGet(ctx, fmt.Sprintf("pvmes:%s,%s", username, host), fmt.Sprintf("pvmes:%s,%s", host, username)).Result()

	if errOfRedis != nil || (val[0] == nil && val[1] == nil) {
		return
	} else {

		if val[0] != nil {

			val1 := val[0].(string)
			err := json.Unmarshal([]byte(val1), &Array)
			if err != nil {
				fmt.Println("error in unmarshling: ", err)
				return
			}

		} else if val[1] != nil {

			val2 := val[1].(string)

			err := json.Unmarshal([]byte(val2), &Array)
			if err != nil {
				fmt.Println("error in unmarshling: ", err)
				return
			}
		}

	}

	Array = append(Array, newDoc)

	go redisServer.Client.SetPvMes(username, host, &Array)

}
