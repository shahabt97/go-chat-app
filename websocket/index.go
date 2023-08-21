package websocketServer

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)


var Broadcast = make(chan *Msg)
var Clients = make(map[*websocket.Conn]bool)

func Websocket(routes *gin.Engine) {
	routes.GET("/ws", HandleConn)
	go HandleAllConnections()
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
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
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

	}

}
