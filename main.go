package main

import (
	"go-chat-app/database"
	"go-chat-app/rabbitmq"
	redisServer "go-chat-app/redis"
	routesHanlder "go-chat-app/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	if err := rabbitmq.RabbitMQInitialization(rabbitmq.PubMessagePublishMaster, rabbitmq.PubMessageConsumeMaster); err != nil {
		log.Fatalf("error in connecting to RabbitMQ: %v\n", err)
	}

	routes := gin.Default()
	database.UtilsInitializations()

	routesHanlder.RouteHandlers(routes)

	// closing Redis client
	defer redisServer.Client.Client.Close()

	//port
	routes.Run(":8080")

}
