package websocketServer

import (
	"fmt"
	"go-chat-app/database"

	// redisServer "go-chat-app/redis"
	"net/http/httptest"
	"strings"

	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

func TestWS(t *testing.T) {


	database.UtilsInitializations()

	go HandleAllConnections()
	go HandleOlineUsers()

	r := gin.Default()

	r.Use(func(c *gin.Context) {

		session := &sessions.Session{}

		session.Values = make(map[interface{}]interface{})

		session.Values["username"] = "shahab"

		c.Set("session", session)

	})
	r.GET("/ws", HandleConn)

	// Create a test server with the router
	s := httptest.NewServer(r)
	defer s.Close()

	// Create a test client with the test server URL
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// c.WriteMessage(websocket.TextMessage, []byte("hello everyone"))

	// c.SetReadDeadline(time.Now().Add(time.Second * 50))
	_, msg, _ := c.ReadMessage()
	fmt.Printf("messageeee in front is: %v \n", string(msg))

	_, msg, _ = c.ReadMessage()
	fmt.Printf("message2 in front is: %v \n", string(msg))

	_, msg, _ = c.ReadMessage()
	fmt.Printf("message2 in front is: %v \n", string(msg))

	_, msg, _ = c.ReadMessage()
	fmt.Printf("message2 in front is: %v \n", string(msg))

	// fmt.Printf("message in front is: %v \n", string(msg))

	c.WriteMessage(websocket.TextMessage, msg)

	_, msg1, err := c.ReadMessage()

	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("message2 in front is: %v \n", string(msg1))

	c.Close()

}
