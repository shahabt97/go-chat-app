package main

import (
	"first/controllers"
	"first/controllers/auth"
	"first/database"
	"first/session"
	websocketServer "first/websocket"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {

	routes := gin.Default()
	database.UtilsInitializations()

	// handling websocket
	websocketServer.Websocket(routes)

	// session
	routes.Use(sessiosnMiddleware)

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
	chatRoutes.GET("/get-messages", controllers.GetMessages)
	chatRoutes.GET("/pv/:username", controllers.PvChatHandler)
	chatRoutes.GET("/pv/search", controllers.SearchInPvChat)
	chatRoutes.GET("/search", controllers.SearchAllMessages)

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
