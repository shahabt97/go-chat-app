package redisServer

import (
	"context"
	"encoding/json"
	"fmt"
	"go-chat-app/config"
	"go-chat-app/database"
	"time"

	"github.com/redis/go-redis/v9"
)

type ClientOfRedis struct {
	Client *redis.Client
}

var client *redis.Client

var Client *ClientOfRedis

func Init() error {

	client = redis.NewClient(&redis.Options{
		Addr: config.ConfigData.RedisHost,
		Password: "83000000",
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	status := client.Ping(ctx)

	err := status.Err()
	if err != nil {
		return err
	}

	Client = &ClientOfRedis{
		Client: client,
	}
	return nil

}

func (c *ClientOfRedis) Keys() ([]string, error) {

	keys, err := c.Client.Keys(context.Background(), "*").Result()
	if err != nil {
		fmt.Println("error in finding key of PV Messages: ", err)
		return []string{}, err
	}
	return keys, nil

}

func (c *ClientOfRedis) SetPvMes(username, host string, Array *[]*database.PvMessage) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jsonData, errOfMarshaling := json.Marshal(Array)
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
				go c.SetPvMes(username, host, Array)
			}
			exist = true
			return

		case key2:
			if err := c.Client.Set(ctx, fmt.Sprintf("pvmes:%s,%s", host, username), jsonData, 10*time.Hour).Err(); err != nil {
				fmt.Println("error in setting pub messages: ", err)

				// try again to set data in Redis
				go c.SetPvMes(username, host, Array)
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

			// try again to set data in Redis
			go c.SetPvMes(username, host, Array)
		}
		return
	}
}

func (c *ClientOfRedis) SetPubMes(array *[]*database.PublicMessage) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(array)
	if err != nil {
		return err
	}

	if err := c.Client.Set(ctx, "pubmessages", data, 10*time.Hour).Err(); err != nil {
		return err
	}

	return nil

}
