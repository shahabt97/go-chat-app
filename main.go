package main

import (
	"first/controllers"
	websocketServer "first/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("SESSION_KEY"))

func main() {

	routes := gin.Default()

	// handling websocket
	websocketServer.Websocket(routes)

	

	//hoempage
	routes.GET("/", func(c *gin.Context) {
		c.File("views/register.html")
	})
	// user
	UserRoutes := routes.Group("/user")
	UserRoutes.POST("/register", controllers.RegisterHandler)
	UserRoutes.POST("/login", controllers.LoginHandler)
	UserRoutes.GET("/login", controllers.GetLoginPage)

	// chat
	chatRoutes := routes.Group("chat")
	chatRoutes.GET("/public")

	// static files
	routes.Static("/public", "public")

	//port
	routes.Run(":8080")

}
