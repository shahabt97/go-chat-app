package main

import (
	"first/controllers"
	"first/controllers/auth"
	"first/database"
	redisServer "first/redis"
	"first/session"
	websocketServer "first/websocket"
	// "fmt"

	"github.com/gin-gonic/gin"
)

func main() {

	routes := gin.Default()
	database.UtilsInitializations()

	// session
	routes.Use(sessiosnMiddleware)

	// handling websocket
	websocketServer.Websocket(routes)
	
	//homepage
	routes.GET("/", controllers.HomePageHandler)

	// user routes
	UserRoutes := routes.Group("/user")
	UserRoutes.POST("/register", controllers.RegisterHandler)
	UserRoutes.POST("/login", controllers.LoginHandler)
	UserRoutes.GET("/register", controllers.RegisterPage)
	UserRoutes.GET("/login", controllers.GetLoginPage)
	UserRoutes.GET("/get-user-id", controllers.GetUserInfoFromSession)
	UserRoutes.GET("logout", auth.AuthHandler, controllers.LogoutHandler)
	UserRoutes.GET("/search", controllers.SearchInUsers)

	// chat routes
	chatRoutes := routes.Group("/chat")
	chatRoutes.Use(auth.AuthHandler)
	chatRoutes.GET("/public", controllers.PublicChatHandler)
	chatRoutes.GET("/public/search", controllers.SearchInPubChat)
	chatRoutes.GET("/pv/:username", controllers.PvChatHandler)
	chatRoutes.GET("/pv/search", controllers.SearchInPvChat)
	chatRoutes.GET("/search", controllers.SearchAllMessages)
	chatRoutes.GET("/file", controllers.FileHandler)

	// static files
	routes.Static("/public", "public")

	// closing Redis client
	defer redisServer.Client.Client.Close()

	//port
	routes.Run(":8080")

}
func sessiosnMiddleware(c *gin.Context) {

	session, _ := session.Store.Get(c.Request, "log-session")
	
	c.Set("session", session)
	c.Next()

}
