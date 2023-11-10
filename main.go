package main

import (
	"fmt"
	"go-chat-app/config"
	"go-chat-app/database"
	"go-chat-app/elasticsearch"
	"go-chat-app/rabbitmq"
	redisServer "go-chat-app/redis"
	routesHanlder "go-chat-app/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	if err := config.EnvExtractor(); err != nil {
		panic(fmt.Sprintf("error in extracting env: %v\n", err))
	}

	if err := elasticsearch.Init(); err != nil {
		panic(fmt.Sprintf("error in initating Elastic: %v\n", err))
	}

	if err := redisServer.Init(); err != nil {
		panic(fmt.Sprintf("error in connecting to Redis: %v\n", err))
	}

	if err := database.UtilsInitializations(); err != nil {
		panic(fmt.Sprintf("error in connecting to database: %v\n", err))
	}

	if err := rabbitmq.RabbitMQInitialization(rabbitmq.PubMessagePublishMaster, rabbitmq.PubMessageConsumeMaster); err != nil {
		panic(fmt.Sprintf("error in connecting to RabbitMQ: %v\n", err))
	}

	routes := gin.Default()

	routesHanlder.RouteHandlers(routes)

	//port
	routes.Run(":8080")

}
