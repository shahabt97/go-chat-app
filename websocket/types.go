package websocketServer

import (
	"first/database"
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
	Username    string
}
type PvMsg struct {
	MessageType int
	Message     []byte
	Username    string
	Host        string
}
type PvMessageContent struct {
	Message   string    `json:"message"`
	Sender    string    `json:"sender"`
	Receiver  string    `json:"receiver"`
	Timestamp time.Time `json:"timestamp"`
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

type PvEvent struct {
	Id        string           `json:"id"`
	EventName string           `json:"eventName"`
	Data      PvMessageContent `json:"data"`
}

type OnlineUsersEvent struct {
	EventName string `json:"eventName"`
	Data      struct {
		OnlineUsers []string
	} `json:"data"`
}

type PvConnection struct {
	Username string
	Host     string
}

type AllPubMessages struct {
	EventName string                    `json:"eventName"`
	Data      []*database.PublicMessage `json:"data"`
}

type AllPvMessages struct {
	EventName string                    `json:"eventName"`
	Data      []*database.PvMessage `json:"data"`
}
