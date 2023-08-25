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
	// z, _ := database.PubMessages.Indexes().List(context.Background())
	// for z.Next(context.Background()) {
	// 	fmt.Println("g: ", z.Current.Index(1))
	// }
	c.File("views/public-chat.html")
}

func GetMessages(c *gin.Context) {
	var Array []*database.PublicMessage
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	status := c.Query("status")
	if status == "public" {

		results, err := database.PubMessages.Find(ctx, bson.M{}, database.FindPubMessagesBasedOnCreatedAtIndexOption)

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
				c.Abort()
				return
			}
			Array = append(Array, document)
		}
		if len(Array) != 0 {
			c.JSON(201, Array)
			c.Abort()
		} else {
			c.JSON(201, []gin.H{})
		}
	}
}
