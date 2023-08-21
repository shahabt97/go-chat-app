// package main

import (
	// "github.com/gin-contrib/sessions"
	// "github.com/gin-contrib/sessions/cookie"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("SESSION_KEY"))

func main() {

	r := gin.Default()

	// r.Use(sessions.Sessions("session-name", store))

	r.GET("/", func(c *gin.Context) {
		session, _ := store.Get(c.Request, "session-nameeeerrrr")
		// Set some session values.
		// session.Values["foo"] = "bar"
		// session.Values[42] = 43
		// Save it before we write to the response/return from the handler.
		// err := session.Save(c.Request, c.Writer)
		// if err != nil {
		// 	fmt.Println("err: ", err)
		// 	return
		// }
		// Retrieve the session value
		// username := session.Get("username")
		// if username == nil {
		// 	username = "guest"
		// }
		username, ok := session.Values[42]

		fmt.Println()
		c.String(200, "Hello, %v", username, "ok: ", ok)
	})

	r.Run(":8080")
}
