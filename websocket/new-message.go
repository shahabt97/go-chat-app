package websocketServer

import (
	"bytes"
	"context"
	"encoding/json"
	"first/database"
	"first/elasticsearch"
	redisServer "first/redis"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleNewPubMes(jsonData *Event, username string) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, errOfMongo := database.PubMessages.InsertOne(ctx, bson.D{
		{Key: "message", Value: jsonData.Data.Message},
		{Key: "sender", Value: username},
		{Key: "createdAt", Value: jsonData.Data.Timestamp},
	})

	if errOfMongo != nil {
		fmt.Println("error storing new public message in Mongo: ", errOfMongo)
		time.Sleep(10 * time.Second)
		go HandleNewPubMes(jsonData, username)
		return
	}

	// Save new public message in elastic
	go SavePubMesInES(result, jsonData)

	// Save new public message in Redis
	go SavePubMesInRedis(result, jsonData, username)

}

func SavePubMesInES(result *mongo.InsertOneResult, jsonData *Event) {

	pubMessageJson := &elasticsearch.PubMessageIndex{
		Id:      result.InsertedID.(primitive.ObjectID).Hex(),
		Message: jsonData.Data.Message,
	}

	pubJsonBytes, errorOfMar := json.Marshal(pubMessageJson)
	if errorOfMar != nil {
		fmt.Println("Error in Marshaling user data for elastic: ", errorOfMar)
		time.Sleep(10 * time.Second)
		go SavePubMesInES(result, jsonData)
		return
	}

	pubReader := bytes.NewReader(pubJsonBytes)
	errPubElas := elasticsearch.Client.CreateDoc("pubmessages", pubReader)
	if errPubElas != nil {
		fmt.Println("Error in creating user in elastic: ", errPubElas)
		time.Sleep(10 * time.Second)
		go SavePubMesInES(result, jsonData)
		return
	}

}

func SavePubMesInRedis(result *mongo.InsertOneResult, jsonData *Event, username string) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newDoc := &database.PublicMessage{
		Message:   jsonData.Data.Message,
		Sender:    username,
		CreatedAt: jsonData.Data.Timestamp,
	}

	var Array = []*database.PublicMessage{}

	val, errOfRedis := redisServer.Client.Client.Get(ctx, "pubmessages").Result()

	if errOfRedis != nil {
		go SavePubMesInRedis(result, jsonData, username)
		return
	}

	errOfUnMar := json.Unmarshal([]byte(val), &Array)
	if errOfUnMar != nil {
		fmt.Println("error in unmarshling: ", errOfUnMar)
		return
	}

	Array = append(Array, newDoc)

	go redisServer.Client.SetPubMes(&Array)

}

func HandleNewPvMes(messageData *PvEvent, username, host string) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := database.PvMessages.InsertOne(ctx, bson.D{
		{Key: "message", Value: messageData.Data.Message},
		{Key: "sender", Value: username},
		{Key: "receiver", Value: messageData.Data.Receiver},
		{Key: "createdAt", Value: messageData.Data.Timestamp},
	})

	if err != nil {
		fmt.Println("error in writing pv message in Mongo: ", err)

		// after 10 seconds again try to insert data in Mongo
		time.Sleep(10 * time.Second)
		go HandleNewPvMes(messageData, username, host)
		return
	}

	// save pv message in elastic
	go SavePvMesInES(result, messageData,username)

	// save pv message in Redis
	go SavePvMesInRedis(messageData, username, host)

}

func SavePvMesInES(result *mongo.InsertOneResult, messageData *PvEvent, username string) {

	pvMessageJson := &elasticsearch.PvMessageIndex{
		Id:       result.InsertedID.(primitive.ObjectID).Hex(),
		Message:  messageData.Data.Message,
		Sender:   username,
		Receiver: messageData.Data.Receiver,
	}

	pvJsonBytes, errorOfMar := json.Marshal(pvMessageJson)
	if errorOfMar != nil {
		fmt.Println("Error in Marshaling user data for elastic: ", errorOfMar)

		// after 10 seconds again try to insert data in Elastic
		time.Sleep(10 * time.Second)
		go SavePvMesInES(result, messageData, username)
		return
	}

	pvReader := bytes.NewReader(pvJsonBytes)
	errPvElas := elasticsearch.Client.CreateDoc("pv-messages", pvReader)
	if errPvElas != nil {
		fmt.Println("Error in creating user in elastic: ", errPvElas)
		time.Sleep(10 * time.Second)
		go SavePvMesInES(result, messageData, username)
		return
	}

}

func SavePvMesInRedis(messageData *PvEvent, username, host string) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newDoc := &database.PvMessage{
		Message:   messageData.Data.Message,
		Sender:    username,
		Receiver:  messageData.Data.Receiver,
		CreatedAt: messageData.Data.Timestamp,
	}

	var Array = []*database.PvMessage{}

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
