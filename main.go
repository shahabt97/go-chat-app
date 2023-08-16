package main

import (
	// "first/controllers"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Msg struct {
	MessageType int
	Message     []byte
}

type MessageContent struct {
	Message   string    `json:"message"`
	Username  string    `json:"username"`
	UserId    int       `json:"userId"`
	Timestamp time.Time `json:"timestamp"`
}

type Event struct {
	Id        int            `json:"id"`
	EventName string         `json:"eventName"`
	Data      MessageContent `json:"data"`
}

var Broadcast = make(chan *Msg)
var Clients = make(map[*websocket.Conn]bool)

func main() {

	routes := gin.Default()



	routes.GET("/ws", HandleConn)
	go HandleAllConnections()
	routes.GET("/", func(c *gin.Context) {
		c.File("views/public-chat.html")
	})

	routes.Static("/public", "public")
	routes.Run(":8080")

}

func HandleAllConnections() {
	for {
		msg := <-Broadcast
		for client := range Clients {
			if err := client.WriteMessage(msg.MessageType, msg.Message); err != nil {
				fmt.Println("error in writing in a client: ", err)
				delete(Clients, client)
				client.Close()
			}
		}
	}
}

func HandleConn(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	fmt.Println("conn", conn.LocalAddr().String())
	Clients[conn] = true
	for {
		messageType, p, err := conn.ReadMessage()
		var jsonData Event
		err2 := json.Unmarshal(p, &jsonData)
		if err2 != nil {
			fmt.Println(err2)
		} else {
			fmt.Println("hiii: ", jsonData)
		}

		if err != nil {
			log.Println(err)
			return
		}
		Broadcast <- &Msg{MessageType: messageType, Message: p}
		fmt.Println(string(p), "and", messageType)

		// if err := conn.WriteMessage(messageType, p); err != nil {
		// 	log.Println(err)
		// 	return
		// }
	}

}
