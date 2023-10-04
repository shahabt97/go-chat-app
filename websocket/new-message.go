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

func HandleNewPubMes(jsonData *Event) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, errOfMongo := database.PubMessages.InsertOne(ctx, bson.D{
		{Key: "message", Value: jsonData.Data.Message},
		{Key: "sender", Value: jsonData.Data.Username},
		{Key: "createdAt", Value: jsonData.Data.Timestamp},
	})

	if errOfMongo != nil {
		fmt.Println("error storing new public message in Mongo: ", errOfMongo)
		return
	}

	// Save new public message in elastic
	go SavePubMesInES(result, jsonData)

	// Save new public message in Redis
	go SavePubMesInRedis(result, jsonData)

}

func SavePubMesInES(result *mongo.InsertOneResult, jsonData *Event) {

	pubMessageJson := &elasticsearch.PubMessageIndex{
		Id:      result.InsertedID.(primitive.ObjectID).Hex(),
		Message: jsonData.Data.Message,
	}

	pubJsonBytes, errorOfMar := json.Marshal(pubMessageJson)
	if errorOfMar != nil {
		fmt.Println("Error in Marshaling user data for elastic: ", errorOfMar)
		return
	}

	pubReader := bytes.NewReader(pubJsonBytes)
	errPubElas := elasticsearch.Client.CreateDoc("pubmessages", pubReader)
	if errPubElas != nil {
		fmt.Println("Error in creating user in elastic: ", errPubElas)
		return
	}

}

func SavePubMesInRedis(result *mongo.InsertOneResult, jsonData *Event) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newDoc := &database.PublicMessage{
		Message:   jsonData.Data.Message,
		Sender:    jsonData.Data.Username,
		CreatedAt: jsonData.Data.Timestamp,
	}

	var Array = []*database.PublicMessage{}

	val, errOfRedis := redisServer.Client.Client.Get(ctx, "pubmessages").Result()

	if errOfRedis != nil {
		fmt.Println("error in getting data from Redis: ", errOfRedis)
		return
	}

	errOfUnMar := json.Unmarshal([]byte(val), &Array)
	if errOfUnMar != nil {
		fmt.Println("error in unmarshling: ", errOfUnMar)
		return
	}

	Array = append(Array, newDoc)

	err:= redisServer.Client.SetPubMes(&Array)
	if err != nil {
		
	}

}

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
