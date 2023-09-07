package routesHanlder

import (
	"go-chat-app/controllers"
	"go-chat-app/middlewares"
	websocketServer "go-chat-app/websocket"

	"github.com/gin-gonic/gin"
)

func RouteHandlers(routes *gin.Engine) {

	// session
	routes.Use(middlewares.SessiosnMiddleware)

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
	UserRoutes.GET("logout", middlewares.AuthHandler, controllers.LogoutHandler)
	UserRoutes.GET("/search", controllers.SearchInUsers)

	// chat routes
	chatRoutes := routes.Group("/chat")
	chatRoutes.Use(middlewares.AuthHandler)
	chatRoutes.GET("/public", controllers.PublicChatHandler)
	chatRoutes.GET("/public/search", controllers.SearchInPubChat)
	chatRoutes.GET("/pv/:username", controllers.PvChatHandler)
	chatRoutes.GET("/pv/search", controllers.SearchInPvChat)
	chatRoutes.GET("/search", controllers.SearchAllMessages)
	chatRoutes.GET("/file", controllers.FileHandler)

	// static files
	routes.Static("/public", "public")
}
