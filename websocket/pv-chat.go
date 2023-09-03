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

var PvBroadcast = make(chan *PvMsg, 1024)

var PvClients = make(map[*websocket.Conn]*PvConnection)
var PvOnlineUsersChan = make(chan bool, 1024)

func HandlePvConnection(c *gin.Context) {

	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	username := c.Query("username")
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
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			result, err := database.PvMessages.InsertOne(ctx, bson.D{
				{Key: "message", Value: messageData.Data.Message},
				{Key: "sender", Value: messageData.Data.Sender},
				{Key: "receiver", Value: messageData.Data.Receiver},
				{Key: "createdAt", Value: messageData.Data.Timestamp},
			})
			if err != nil {
				fmt.Println("error in writing pv message in Mongo: ", err)
				return
			}
			pvMessageJson := &elasticsearch.PvMessageIndex{
				Id:       result.InsertedID.(primitive.ObjectID).Hex(),
				Message:  messageData.Data.Message,
				Sender:   messageData.Data.Sender,
				Receiver: messageData.Data.Receiver,
			}
			pvJsonBytes, errorOfMar := json.Marshal(pvMessageJson)
			if errorOfMar != nil {
				fmt.Println("Error in Marshaling user data for elastic: ", errorOfMar)
				return
			}
			pvReader := bytes.NewReader(pvJsonBytes)
			errPvElas := elasticsearch.Client.CreateDoc("pv-messages", pvReader)
			if errPvElas != nil {
				fmt.Println("Error in creating user in elastic: ", errPvElas)
				return
			}
		}()
		PvBroadcast <- &PvMsg{MessageType: messageType, Message: message, Username: username, Host: host}
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

	allMessages := &AllPvMessages{
		EventName: "all messages",
		Data:      Array,
	}

	jsonData, errOfMarshaling := json.Marshal(allMessages)

	if errOfMarshaling != nil {
		fmt.Println("error in Marshaling pv messages: ", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		conn.Close()
		return
	}

}
