package main

import (
	"go-chat-app/database"
	redisServer "go-chat-app/redis"
	routesHanlder "go-chat-app/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	routes := gin.Default()
	database.UtilsInitializations()

	routesHanlder.RouteHandlers(routes)

	// closing Redis client
	defer redisServer.Client.Client.Close()

	//port
	routes.Run(":8080")

}
