package websocketServer

import (
	"context"
	"encoding/json"
	"first/database"
	redisServer "first/redis"
	"fmt"
	"log"
	"sort"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var PvBroadcast = make(chan *PvMsg, 1024)

var PvClients = make(map[*websocket.Conn]*PvConnection)
var PvOnlineUsersChan = make(chan bool, 1024)

func HandlePvConnection(c *gin.Context) {

	sessionRaw, _ := c.Get("session")
	session, _ := sessionRaw.(*sessions.Session)
	if session.Values["username"] == nil {
		c.JSON(403, gin.H{
			"message": "User unAuthorized",
		})
		c.Abort()
		return
	}

	username := session.Values["username"].(string)

	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	host := c.Query("host")

	PvClients[conn] = &PvConnection{Username: username, Host: host}

	go GetPvMessages(username, host, conn)

	for {

		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("err in reading from websocket: ", err)
			delete(PvClients, conn)
			OnlineUsersChan <- true
			return
		}

		messageData := &PvEvent{}

		if err := json.Unmarshal(message, messageData); err != nil {
			fmt.Println("error in unmarshaling data")
			continue
		}

		messageData.Data.Sender = username

		newP, err2 := json.Marshal(messageData)
		if err2 != nil {
			fmt.Println(err2)
			continue
		}

		go HandleNewPvMes(messageData, host)

		PvBroadcast <- &PvMsg{MessageType: messageType, Message: newP, Username: username, Host: host}

	}

}

func HandleAllPvConnections() {
	for {
		pvMsg := <-PvBroadcast
		go func() {
			for client := range PvClients {
				if (PvClients[client].Username == pvMsg.Username) ||
					(PvClients[client].Username == pvMsg.Host && PvClients[client].Host == pvMsg.Username) {
					go func(cl *websocket.Conn) {
						if err := cl.WriteMessage(pvMsg.MessageType, pvMsg.Message); err != nil {
							fmt.Println("error in writing in a client: ", err)
							cl.Close()
						}
					}(client)
				}
			}
		}()
	}
}

func GetPvMessages(username, host string, conn *websocket.Conn) {

	var Array = []*database.PvMessage{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	val, errOfRedis := redisServer.Client.Client.MGet(ctx, fmt.Sprintf("pvmes:%s,%s", username, host), fmt.Sprintf("pvmes:%s,%s", host, username)).Result()

	if errOfRedis != nil || (val[0] == nil && val[1] == nil) {

		results, err := database.PvMessages.Find(ctx, bson.D{{Key: "$or", Value: []bson.D{{{Key: "sender", Value: username}, {Key: "receiver", Value: host}},
			{{Key: "sender", Value: host}, {Key: "receiver", Value: username}}}}}, database.FindPvMessagesOption)
		if err != nil {
			fmt.Println("error in getting all pv messages is: ", err)
			return
		}

		for results.Next(ctx) {
			var document = &database.PvMessage{}
			if err := results.Decode(document); err != nil {
				fmt.Println("error in reading all results of public messages: ", err)
				return
			}
			Array = append(Array, document)
		}

		sort.Slice(Array, func(i, j int) bool {
			return Array[i].CreatedAt.Before(Array[j].CreatedAt)
		})

		go redisServer.Client.SetPvMes(username, host, &Array)

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

	allMessages := &AllPvMessages{
		EventName: "all messages",
		Data:      Array,
	}

	jsonData, errOfMarshaling := json.Marshal(allMessages)

	if errOfMarshaling != nil {
		fmt.Println("error in Marshaling pv messages: ", errOfMarshaling)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		conn.Close()
		return
	}

}
