package websocketServer

import (
	"bytes"
	"context"
	"encoding/json"
	"first/database"
	"first/elasticsearch"
	"fmt"
	"log"
	"sync"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PvClientsMutex struct {
	Clients map[*websocket.Conn]*PvConnection
	sync.Mutex
}

var PvClients = &PvClientsMutex{Clients: make(map[*websocket.Conn]*PvConnection)}

var PvBroadcast = make(chan *PvMsg, 1024)

// var Clients = make(map[*websocket.Conn]string)
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

	PvClients.Lock()
	PvClients.Clients[conn] = &PvConnection{Username: username, Host: host}
	PvClients.Unlock()

	conn.SetCloseHandler(func(code int, text string) error {
		PvClients.Lock()
		delete(PvClients.Clients, conn)
		OnlineUsersChan <- true
		PvClients.Unlock()
		return nil
	})
	for {

		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("err in reading from websocket: ", err)
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
			errPvElas := elasticsearch.Client.CreateDoc("pubmessages", pvReader)
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
			PvClients.Lock()
			for client := range PvClients.Clients {
				if (PvClients.Clients[client].Username == pvMsg.Username) ||
					(PvClients.Clients[client].Username == pvMsg.Host && PvClients.Clients[client].Host == pvMsg.Username) {
					go func(cl *websocket.Conn) {
						if err := cl.WriteMessage(pvMsg.MessageType, pvMsg.Message); err != nil {
							fmt.Println("error in writing in a client: ", err)
							cl.Close()
						}
					}(client)
				}
			}
			PvClients.Unlock()
		}()
	}
}
