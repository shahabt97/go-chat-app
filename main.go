package main

import (
	"fmt"
	"go-chat-app/database"
	"go-chat-app/rabbitmq"
	redisServer "go-chat-app/redis"
	routesHanlder "go-chat-app/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	if err := rabbitmq.RabbiInitialization(); err != nil {
		fmt.Printf("error in connecting to RabbitMQ: %v\n", err)
	}
	
	routes := gin.Default()
	database.UtilsInitializations()

	routesHanlder.RouteHandlers(routes)

	// closing Redis client
	defer redisServer.Client.Client.Close()

	//port
	routes.Run(":8080")

}
