package controllers

import (
	"context"
	"first/database"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func PublicChatHandler(c *gin.Context) {
	c.File("views/public-chat.html")
}

func GetMessages(c *gin.Context) {
	var Array []*database.PublicMessage
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	status := c.Query("status")
	fmt.Println(status)
	if status == "public" {

		results, err := database.PubMessages.Find(ctx, bson.M{})
		if err != nil {
			fmt.Println("error in getting all public messages is: ", err)
			c.JSON(500, gin.H{})
			return
		}
		for results.Next(ctx) {

			var document = &database.PublicMessage{}

			if err := results.Decode(document); err != nil {
				fmt.Println("error in reading all results of public messages: ", err)
				c.JSON(500, gin.H{})
				return
			}
			Array = append(Array, document)

		}
		fmt.Println("Array: ", Array)
		c.JSON(201, Array)
	}
}
