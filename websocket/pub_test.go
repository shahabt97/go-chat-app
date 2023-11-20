package websocketServer

import (
	"context"
	"encoding/json"
	"fmt"
	"go-chat-app/config"
	"go-chat-app/database"
	"go-chat-app/elasticsearch"
	"go-chat-app/rabbitmq"
	redisServer "go-chat-app/redis"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"

	"net/http/httptest"

	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

func TestWS(t *testing.T) {

	runtime.GOMAXPROCS(10)

	if err := config.EnvExtractor(); err != nil {
		log.Fatalf("error in extracting env: %v\n", err)
	}

	if err := elasticsearch.Init(); err != nil {
		log.Fatalf("error in initating Elastic: %v\n", err)
	}

	if err := redisServer.Init(); err != nil {
		log.Fatalf("error in connecting to Redis: %v\n", err)
	}

	if err := database.UtilsInitializations(); err != nil {
		log.Fatalf("error in connecting to database: %v\n", err)
	}

	// if err := rabbitmq.RabbitMQInitialization(rabbitmq.PubMessagePublishMaster, rabbitmq.PubMessageConsumeMaster); err != nil {
	// 	log.Fatalf("error in connecting to RabbitMQ: %v\n", err)
	// }

	go HandleAllConnections()
	go HandleOlineUsers()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(func(c *gin.Context) {

		session := &sessions.Session{}

		session.Values = make(map[interface{}]interface{})

		session.Values["username"] = generateRandomString(4)

		c.Set("session", session)

	})
	r.GET("/ws", HandleConn)

	// Create a test server with the router
	s := httptest.NewServer(r)
	defer s.Close()

	var mainMutex sync.Mutex

	for {

		for i := 0; i < 300; i++ {
			go func() {
				// Create a test client with the test server URL
				// "://localhost:8080"
				wsURL := "ws://localhost" + strings.TrimPrefix(s.URL, "http://127.0.0.1") + "/ws"
				// fmt.Println("url is: ", wsURL)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				mainMutex.Lock()
				c, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
				time.Sleep(3 * time.Millisecond)
				mainMutex.Unlock()
				// go func() {
				// 	time.Sleep(60 * time.Second)
				// 	c.Close()
				// }()

				if err != nil {
					fmt.Printf("error in connecting to websocket server: %v\n", err)
					return
				}
				defer c.Close()

				time.Sleep(1 * time.Second)

				go func() {
					defer c.Close()
					data := rabbitmq.Event{
						EventName: "chat message",
						Data: rabbitmq.MessageContent{
							Message:   "hello everyone",
							Username:  "ShahabTayebi",
							Timestamp: time.Now()}}

					_, err := json.Marshal(data)
					if err != nil {
						fmt.Printf("error in Marshaling test message to write in client test: %v\n", err)
						return
					}

					for {
						time.Sleep(7 * time.Second)
						// err = c.WriteMessage(websocket.TextMessage, []byte("hello everyone"))
						// if err != nil {
						// 	fmt.Printf("error in writing message in client test: %v\n", err)
						// 	return
						// }
					}
				}()

				for {
					_, msg, err := c.ReadMessage()
					if err != nil {
						fmt.Printf("error in reading message in client test: %v\n", err)
						return
					}
					_ = msg
					// fmt.Printf("\n\nmessage in front is: %v \n\n", string(msg))
				}
			}()

		}

		time.Sleep(9 * time.Second)
	}

}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
