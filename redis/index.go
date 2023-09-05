package redisServer

import (
	"context"
	"encoding/json"
	"first/database"
	"first/hosts"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

type ClientOfRedis struct {
	Client *redis.Client
}

var client = redis.NewClient(&redis.Options{
	Addr: hosts.RedisHost,
	DB:   0,
})

var Client = &ClientOfRedis{
	Client: client,
}

func (c *ClientOfRedis) SetPubMessages() {

	var Array = []*database.PublicMessage{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	results, err := database.PubMessages.Find(ctx, bson.M{}, database.FindPubMessagesBasedOnCreatedAtIndexOption)

	if err != nil {
		fmt.Println("error in getting all public messages is: ", err)
		return
	}

	for results.Next(ctx) {
		var document = &database.PublicMessage{}
		if err := results.Decode(document); err != nil {
			fmt.Println("error in reading all results of public messages: ", err)
			return
		}
		Array = append(Array, document)
	}

	jsonData, errOfMarshaling := json.Marshal(Array)
	if errOfMarshaling != nil {
		fmt.Println("error in marshaling pub messages of redis: ", errOfMarshaling)
		return
	}
	if err := c.Client.Set(ctx, "pubmessages", jsonData, 10*time.Hour).Err(); err != nil {
		fmt.Println("error in setting pub messages: ", err)
		return
	}

}

func (c *ClientOfRedis) SetPvMessages(username, host string) {

	var Array = []*database.PvMessage{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := database.PvMessages.Find(ctx, bson.D{{Key: "$or", Value: []bson.D{{{Key: "sender", Value: username}, {Key: "receiver", Value: host}},
		{{Key: "sender", Value: host}, {Key: "receiver", Value: username}}}}}, database.FindPvMessagesOption)
	if err != nil {
		fmt.Println("error in getting all pv messages is: ", err)
		return
	}

	for results.Next(ctx) {
		var document = &database.PvMessage{}
		if err := results.Decode(document); err != nil {
			fmt.Println("error in reading all results of pv messages: ", err)
			return
		}
		Array = append(Array, document)
	}

	jsonData, errOfMarshaling := json.Marshal(&Array)
	if errOfMarshaling != nil {
		fmt.Println("error in marshaling pub messages of redis: ", errOfMarshaling)
		return
	}

	keys, errOfKeys := c.Keys()

	if errOfKeys != nil {
		return
	}

	key1 := fmt.Sprintf("pvmes:%s,%s", username, host)
	key2 := fmt.Sprintf("pvmes:%s,%s", host, username)
	exist := false

	for _, value := range keys {

		switch value {

		case key1:
			if err := c.Client.Set(ctx, fmt.Sprintf("pvmes:%s,%s", username, host), jsonData, 10*time.Hour).Err(); err != nil {
				fmt.Println("error in setting pub messages: ", err)
			}
			exist = true
			return

		case key2:
			if err := c.Client.Set(ctx, fmt.Sprintf("pvmes:%s,%s", host, username), jsonData, 10*time.Hour).Err(); err != nil {
				fmt.Println("error in setting pub messages: ", err)
			}
			exist = true
			return

		default:
			continue
		}
	}
	if !exist {
		if err := c.Client.Set(ctx, fmt.Sprintf("pvmes:%s,%s", username, host), jsonData, 10*time.Hour).Err(); err != nil {
			fmt.Println("error in setting pub messages: ", err)
		}
		return
	}
}

func (c *ClientOfRedis) Keys() ([]string, error) {

	keys, err := c.Client.Keys(context.Background(), "*").Result()
	if err != nil {
		fmt.Println("error in finding key of PV Messages: ", err)
		return []string{}, err
	}
	return keys, nil

}
