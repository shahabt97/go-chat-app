package redisServer

import (
	"context"
	"encoding/json"
	"first/database"
	"first/hosts"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
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

func (c *ClientOfRedis) SetPubMes(Array *[]*database.PublicMessage) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jsonData, errOfMarshaling := json.Marshal(Array)
	if errOfMarshaling != nil {
		fmt.Println("error in marshaling pub messages of redis: ", errOfMarshaling)
		go c.SetPubMes(Array)
		return
	}

	if err := c.Client.Set(ctx, "pubmessages", jsonData, 10*time.Hour).Err(); err != nil {
		fmt.Println("error in setting pub messages: ", err)

		// try again to set data in Redis
		go c.SetPubMes(Array)
		return
	}
}
