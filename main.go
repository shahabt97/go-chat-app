package main

import (
	"first/database"
	redisServer "first/redis"
	routesHanlder "first/routes"


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
