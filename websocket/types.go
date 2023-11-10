package websocketServer

import (
	"go-chat-app/database"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     CheckOrigin,
}

func CheckOrigin(r *http.Request) bool {
	return true
}

type Msg struct {
	MessageType int
	Message     []byte
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
	Message   string    `json:"message"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
}

type Event struct {
	EventName string         `json:"eventName"`
	Data      MessageContent `json:"data"`
	// MongoId   *mongo.InsertOneResult `json:"-"`
}

type PvEvent struct {
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
	EventName string                     `json:"eventName"`
	Data      *[]*database.PublicMessage `json:"data"`
}

type AllPvMessages struct {
	EventName string                `json:"eventName"`
	Data      []*database.PvMessage `json:"data"`
}

type ClientDataInPub struct {
	Username string
	Mu       sync.Mutex
}
