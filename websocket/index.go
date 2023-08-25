package websocketServer

import (
	"context"
	"encoding/json"
	"first/database"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type OnlineUsersHandligStruct struct {
// 	status   string
// 	Conn     *websocket.Conn
// 	Username string
// }

type ClientsMutex struct {
	Clients map[*websocket.Conn]string
	sync.Mutex
}

var Clients = &ClientsMutex{Clients: make(map[*websocket.Conn]string)}

var Broadcast = make(chan *Msg, 1024)

// var Clients = make(map[*websocket.Conn]string)
var OnlineUsersChan = make(chan bool, 1024)

func Websocket(routes *gin.Engine) {
	routes.GET("/ws", HandleConn)
	go HandleAllConnections()
	go HandleOlineUsers()
}

func HandleAllConnections() {
	for {
		msg := <-Broadcast
		go func() {
			for client := range Clients.Clients {
				go func(cl *websocket.Conn) {
					if err := cl.WriteMessage(msg.MessageType, msg.Message); err != nil {
						fmt.Println("error in writing in a client: ", err)
						cl.Close()
					}
				}(client)
			}
		}()

	}
}

func HandleConn(c *gin.Context) {
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	conn.SetCloseHandler(func(code int, text string) error {

		Clients.Lock()
		delete(Clients.Clients, conn)
		OnlineUsersChan <- true
		Clients.Unlock()

		// fmt.Println("code: ", code, "reason: ", text)
		return nil
	})
	id := c.Query("id")
	username := c.Query("username")

	Clients.Lock()
	Clients.Clients[conn] = username
	OnlineUsersChan <- true
	Clients.Unlock()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("err in reading from websocket: ", err)
			return
		}
		var jsonData = &Event{}
		err2 := json.Unmarshal(p, jsonData)
		if err2 != nil {
			fmt.Println(err2)
			continue
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				fmt.Println("error in making string of id to ObjectId of mongo: ", err)
			}
			database.PubMessages.InsertOne(ctx, bson.M{
				"sender": bson.M{
					"username": username,
					"id":       objectID,
				},
				"message":   jsonData.Data.Message,
				"CreatedAt": jsonData.Data.Timestamp,
			})
		}()

		Broadcast <- &Msg{MessageType: messageType, Message: p, Username: username}
		// fmt.Println(string(p), "and", messageType)
	}

}

func HandleOlineUsers() {
	for {
		<-OnlineUsersChan
		Array := []string{}
		Clients.Lock()
		for client := range Clients.Clients {
			Array = append(Array, Clients.Clients[client])
		}
		onlineUsersEvent := &OnlineUsersEvent{EventName: "online users", Data: struct{ OnlineUsers []string }{OnlineUsers: Array}}
		onlineEventJson, _ := json.Marshal(onlineUsersEvent)

		for client := range Clients.Clients {
			if err := client.WriteMessage(1, onlineEventJson); err != nil {
				fmt.Println("error in sending online users: ", err)
				client.Close()
			}
		}
		Clients.Unlock()
	}

}
