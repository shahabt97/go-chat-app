package websocketServer

import (
	"context"
	"encoding/json"
	"first/database"
	redisServer "first/redis"
	"fmt"
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

		var jsonData = &Event{}
		err2 := json.Unmarshal(p, jsonData)
		if err2 != nil {
			fmt.Println(err2)
			continue
		}

		// add new message to Mongo ,Elastic and Redis
		go HandleNewPubMes(jsonData, username)

		Broadcast <- &Msg{MessageType: messageType, Message: p, Username: username}

	}

}

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

func GetPubMessages(conn *websocket.Conn) {

	var Array = []*database.PublicMessage{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	val, errOfRedis := redisServer.Client.Client.Get(ctx, "pubmessages").Result()

	if errOfRedis != nil {
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
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		conn.Close()
		return
	}
}
