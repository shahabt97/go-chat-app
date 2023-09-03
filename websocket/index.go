package websocketServer

import (
	"bytes"
	"context"
	"encoding/json"
	"first/database"
	"first/elasticsearch"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	conn.SetCloseHandler(func(code int, text string) error {

		delete(Clients, conn)
		OnlineUsersChan <- true
		return nil

	})

	id := c.Query("id")
	username := c.Query("username")

	Clients[conn] = username
	OnlineUsersChan <- true

	go GetPubMessages(conn)

	for {
		time.Sleep(15 * time.Second)
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

// go json.Compact()


		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				fmt.Println("error in making string of id to ObjectId of mongo: ", err)
			}

			result, _ := database.PubMessages.InsertOne(ctx, bson.D{
				{Key: "message", Value: jsonData.Data.Message},
				{Key: "sender", Value: bson.D{
					{Key: "id", Value: objectID},
					{Key: "username", Value: username},
				}},
				{Key: "createdAt", Value: jsonData.Data.Timestamp},
			})
			pubMessageJson := &elasticsearch.PubMessageIndex{
				Id:      result.InsertedID.(primitive.ObjectID).Hex(),
				Message: jsonData.Data.Message,
			}
			pubJsonBytes, errorOfMar := json.Marshal(pubMessageJson)
			if errorOfMar != nil {
				fmt.Println("Error in Marshaling user data for elastic: ", err)
				return
			}
			pubReader := bytes.NewReader(pubJsonBytes)
			errPubElas := elasticsearch.Client.CreateDoc("pubmessages", pubReader)
			if errPubElas != nil {
				fmt.Println("Error in creating user in elastic: ", errPubElas)
				return
			}
		}()
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

	allMessages := &AllPubMessages{
		EventName: "all messages",
		Data:      Array,
	}

	jsonData, errOfMarshaling := json.Marshal(allMessages)

	if errOfMarshaling != nil {
		fmt.Println("error in Marshaling public messages: ", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		conn.Close()
		return
	}
}
