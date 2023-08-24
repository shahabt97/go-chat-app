package websocketServer

import (
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Msg struct {
	MessageType int
	Message     []byte
}

type MessageContent struct {
	Message  string `json:"message"`
	Username string `json:"username"`
	// UserId    int       `json:"userId"`
	Timestamp time.Time `json:"timestamp"`
}

type Event struct {
	Id        string         `json:"id"`
	EventName string         `json:"eventName"`
	Data      MessageContent `json:"data"`
}
