package websocketServer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gin-gonic/gin"

	"go-chat-app/database"
	"go-chat-app/rabbitmq"
	redisServer "go-chat-app/redis"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	Clients         = make(map[*websocket.Conn]*ClientDataInPub)
	Broadcast       = make(chan *Msg, 1024)
	OnlineUsersChan = make(chan bool, 1024)
	ClientsMutex    = sync.RWMutex{}
	UpgraderMutex   = sync.Mutex{}
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
			for client, cliData := range Clients {
				go func(cl *websocket.Conn, mu *sync.Mutex) {
					mu.Lock()
					if err := cl.WriteMessage(msg.MessageType, msg.Message); err != nil {
						fmt.Println("error in writing in a client: ", err)
						cl.Close()
					}
					mu.Unlock()
				}(client, &cliData.Mu)
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
	// username := GenerateRandomString(40)

	// UpgraderMutex.Lock()
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	// time.Sleep(3 * time.Millisecond)
	// UpgraderMutex.Unlock()
	if err != nil {
		fmt.Printf("error in upgrading http to websocket: %v\n", err)
		return
	}

	defer conn.Close()

	cliData := &ClientDataInPub{Username: username}

	ClientsMutex.Lock()
	Clients[conn] = cliData
	ClientsMutex.Unlock()

	OnlineUsersChan <- true
	fmt.Println("number of users are: ", len(Clients))
	go GetPubMessages(conn, &cliData.Mu)

	for {

		messageType, p, err := conn.ReadMessage()
		if err != nil {

			fmt.Println("the error in reading from socket client in server: ", err)

			ClientsMutex.Lock()
			delete(Clients, conn)
			ClientsMutex.Unlock()

			OnlineUsersChan <- true
			return
		}

		var jsonData = &rabbitmq.Event{}
		err2 := json.Unmarshal(p, jsonData)
		if err2 != nil {
			fmt.Println("error in Unmarshaling socket client message: ", err2)
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

		var array []*ClientDataInPub

		ClientsMutex.RLock()
		for client := range Clients {
			array = append(array, Clients[client])
		}
		ClientsMutex.RUnlock()

		usernames := lo.Map[*ClientDataInPub, string](array, func(item *ClientDataInPub, _ int) string {
			return item.Username
		})

		onlineUsersEvent := &OnlineUsersEvent{EventName: "online users", Data: struct{ OnlineUsers []string }{OnlineUsers: usernames}}
		onlineEventJson, _ := json.Marshal(onlineUsersEvent)

		ClientsMutex.RLock()
		for client, cliData := range Clients {
			go func(cl *websocket.Conn, mu *sync.Mutex) {
				mu.Lock()
				time.Sleep(3 * time.Millisecond)
				if err := cl.WriteMessage(websocket.TextMessage, onlineEventJson); err != nil {
					fmt.Println("error in sending online users: ", err)
					cl.Close()
				}
				mu.Unlock()
			}(client, &cliData.Mu)
		}
		ClientsMutex.RUnlock()

	}
}

// get all public messages in the beginning of websocket connection
func GetPubMessages(conn *websocket.Conn, mu *sync.Mutex) {

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
					panic(fmt.Sprintf("error in decoding public message document into go struct: %v", err))
				}
				Array = append(Array, document)
			}

			go func() {
				// put data in Redis
				err := redisServer.Client.SetPubMes(&Array)
				if err != nil {
					fmt.Println("error in setting pubMessages in Redis: ", err)
					conn.Close()
					return
				}
			}()

		} else {
			fmt.Printf("error in getting pubMessages from Redis: %v\n", err)
			conn.Close()
			return
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
	mu.Lock()
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		conn.Close()
		mu.Unlock()
		return
	}
	mu.Unlock()

}

func GenerateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
