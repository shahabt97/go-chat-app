package main

import (
	"first/controllers"
	"first/controllers/auth"
	"first/session"
	websocketServer "first/websocket"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("SESSION_KEY"))

func main() {

	routes := gin.Default()

	// handling websocket
	websocketServer.Websocket(routes)

	// session
	routes.Use(sessiosnMiddleware)

	//homepage
	routes.GET("/", func(c *gin.Context) {
		fmt.Println(c.Get("session"))
		c.File("views/register.html")
	})

	// user
	UserRoutes := routes.Group("/user")
	UserRoutes.POST("/register", controllers.RegisterHandler)
	UserRoutes.POST("/login", controllers.LoginHandler)
	UserRoutes.GET("/login", controllers.GetLoginPage)
	UserRoutes.GET("/get-user-id", controllers.GetUserInfoFromSession)

	// chat
	chatRoutes := routes.Group("chat")
	chatRoutes.Use(auth.AuthHandler)
	chatRoutes.GET("/public", controllers.PublicChatHandler)
	chatRoutes.GET("/get-messages", controllers.GetMessages)

	

	// static files
	routes.Static("/public", "public")

	//port
	routes.Run(":8080")

}
func sessiosnMiddleware(c *gin.Context) {
	session, _ := session.Store.Get(c.Request, "log-session")
	err := session.Save(c.Request, c.Writer)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	c.Set("session", session)
	c.Next()
}
