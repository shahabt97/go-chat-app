package websocketServer

import (
	"context"
	"encoding/json"
	"fmt"
	"go-chat-app/database"
	"go-chat-app/rabbitmq"
	redisServer "go-chat-app/redis"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var Clients = make(map[*websocket.Conn]string)

var Broadcast = make(chan *Msg, 1024)

var OnlineUsersChan = make(chan bool, 1024)

func Websocket(routes *gin.Engine) {
	routes.GET("/ws", HandleConn)
	routes.GET("/ws/pv", HandlePvConnection)
	go HandleAllConnections()
	go HandleOlineUsers()
	go HandleAllPvConnections()
}

// send a message to all online connections
func HandleAllConnections() {
	for {
		msg := <-Broadcast

		go func() {
			for client := range Clients {
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

	sessionRaw, _ := c.Get("session")
	session, _ := sessionRaw.(*sessions.Session)
	if session.Values["username"] == nil {
		c.JSON(403, gin.H{
			"message": "User unAuthorized",
		})
		c.Abort()
		return
	}

	// get username from session
	username := session.Values["username"].(string)

	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	Clients[conn] = username
	OnlineUsersChan <- true

	go GetPubMessages(conn)

	for {

		messageType, p, err := conn.ReadMessage()
		if err != nil {
			delete(Clients, conn)
			OnlineUsersChan <- true
			return
		}

		var jsonData = &rabbitmq.Event{}
		err2 := json.Unmarshal(p, jsonData)
		if err2 != nil {
			fmt.Println(err2)
			continue
		}

		jsonData.Data.Username = username

		newP, err3 := json.Marshal(jsonData)
		if err3 != nil {
			fmt.Println(err3)
			continue
		}

		// add new message to Mongo ,Elastic and Redis
		err = rabbitmq.PubMessagesPublisher(jsonData, rabbitmq.PubMessagePublishMaster)
		if err != nil {
			fmt.Printf("error in publishing a pub message: %v\n", err)
			continue
		}

		Broadcast <- &Msg{MessageType: messageType, Message: newP}

	}

}

// sending online users
func HandleOlineUsers() {

	for {
		<-OnlineUsersChan
		Array := []string{}
		for client := range Clients {
			Array = append(Array, Clients[client])
		}
		onlineUsersEvent := &OnlineUsersEvent{EventName: "online users", Data: struct{ OnlineUsers []string }{OnlineUsers: Array}}
		onlineEventJson, _ := json.Marshal(onlineUsersEvent)

		for client := range Clients {
			if err := client.WriteMessage(websocket.TextMessage, onlineEventJson); err != nil {
				fmt.Println("error in sending online users: ", err)
				client.Close()
			}
		}
	}

}

// get all public messages in the beginning of websocket connection
func GetPubMessages(conn *websocket.Conn) {

	var Array = []*database.PublicMessage{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// first check Redis for messages
	val, errOfRedis := redisServer.Client.Client.Get(ctx, "pubmessages").Result()

	if errOfRedis != nil {

		// fetch data from database
		results, err := database.PubMessages.Find(ctx, bson.M{}, database.FindPubMessagesOption)

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

		// put data in Redis
		go redisServer.Client.SetPubMes(&Array)

	} else {
		err := json.Unmarshal([]byte(val), &Array)
		if err != nil {
			fmt.Println("error in unmarshling: ", err)
			return
		}
	}

	allMessages := &AllPubMessages{
		EventName: "all messages",
		Data:      &Array,
	}

	jsonData, errOfMarshaling := json.Marshal(allMessages)

	if errOfMarshaling != nil {
		fmt.Println("error in Marshaling public messages: ", errOfMarshaling)
		return
	}

	// send messages via websocket
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		conn.Close()
		return
	}
}
