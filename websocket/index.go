package websocketServer

import (
	"context"
	"encoding/json"
	"fmt"
	"go-chat-app/database"
	"go-chat-app/rabbitmq"
	redisServer "go-chat-app/redis"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	Clients         = make(map[*websocket.Conn]string)
	Broadcast       = make(chan *Msg, 1024)
	OnlineUsersChan = make(chan bool, 1024)
	ClientsMutex    = sync.RWMutex{}
)

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
			ClientsMutex.RLock()
			for client := range Clients {
				go func(cl *websocket.Conn) {
					if err := cl.WriteMessage(msg.MessageType, msg.Message); err != nil {
						fmt.Println("error in writing in a client: ", err)
						cl.Close()
					}
				}(client)
			}
			ClientsMutex.RUnlock()
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
		fmt.Printf("error in upgrading http to websocket: %v\n", err)
		return
	}

	defer conn.Close()

	ClientsMutex.Lock()
	Clients[conn] = username
	ClientsMutex.Unlock()

	OnlineUsersChan <- true

	go GetPubMessages(conn)

	for {

		messageType, p, err := conn.ReadMessage()
		if err != nil {

			ClientsMutex.Lock()
			delete(Clients, conn)
			ClientsMutex.Unlock()

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

		ClientsMutex.RLock()
		for client := range Clients {
			Array = append(Array, Clients[client])
		}
		ClientsMutex.RUnlock()

		onlineUsersEvent := &OnlineUsersEvent{EventName: "online users", Data: struct{ OnlineUsers []string }{OnlineUsers: Array}}
		onlineEventJson, _ := json.Marshal(onlineUsersEvent)

		ClientsMutex.RLock()
		for client := range Clients {
			if err := client.WriteMessage(websocket.TextMessage, onlineEventJson); err != nil {
				fmt.Println("error in sending online users: ", err)
				client.Close()
			}
		}
		ClientsMutex.RUnlock()

	}
}

// get all public messages in the beginning of websocket connection
func GetPubMessages(conn *websocket.Conn) {

	var Array = []*database.PublicMessage{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// first check Redis for messages
	val, err := redisServer.Client.Client.Get(ctx, "pubmessages").Result()

	if err != nil {

		if err == redis.Nil {

			// fetch data from database
			results, err := database.PubMessages.Find(ctx, bson.M{}, database.FindPubMessagesOption)

			if err != nil {
				fmt.Println("error in getting all public messages is: ", err)
				conn.Close()
				return
			}

			for results.Next(ctx) {
				var document = &database.PublicMessage{}
				if err := results.Decode(document); err != nil {
					fmt.Println("error in reading all results of public messages: ", err)
					conn.Close()
					return
				}
				Array = append(Array, document)
			}

			go func() {
				// put data in Redis
				err := redisServer.Client.SetPubMes(&Array)
				if err != nil {
					panic(err)
				}
			}()

		} else {
			fmt.Printf("error in getting pubMessages from Redis: %v\n", err)
			panic(err)
		}

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

	jsonData, err := json.Marshal(allMessages)

	if err != nil {
		fmt.Println("error in Marshaling public messages: ", err)
		return
	}

	// send messages via websocket
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		conn.Close()
		return
	}
}
