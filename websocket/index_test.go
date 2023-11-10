package websocketServer

import (
	"fmt"
	"go-chat-app/database"
	redisServer "go-chat-app/redis"
	"net/http/httptest"
	"strings"

	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestWebsocket(t *testing.T) {

	database.UtilsInitializations()

	WSHandler := func(c *gin.Context) {

		// Upgrade the HTTP connection to a WebSocket connection
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()

		// GetPubMessages(conn)

		defer redisServer.Client.Client.Close()


		for {

			mt, msg, err := conn.ReadMessage()
			if err != nil {
				// t.Error(err)
				return
			}

			// fmt.Printf("message from server: %v \n", string(msg))

			err = conn.WriteMessage(mt, msg)
			if err != nil {
				t.Error(err)
			}

		}

	}

	r := gin.Default()
	r.GET("/ws", WSHandler)

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

	// fmt.Printf("message in front is: %v \n", string(msg))

	c.WriteMessage(websocket.TextMessage, msg)

	_, msg1, err := c.ReadMessage()

	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("message2 in front is: %v \n", string(msg1))

	c.Close()

}
