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

func PvChatHandler(c *gin.Context) {
	c.File("views/pv-chat.html")
}

func GetMessages(c *gin.Context) {
	if c.Query("status") == "public" {
		var Array []*database.PublicMessage
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
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

	} else if c.Query("status") == "pv" {
		var Array []*database.PvMessage
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		host := c.Query("hostUser")
		username := c.Query("username")
		results, err := database.PvMessages.Find(ctx, bson.D{{Key: "$or", Value: []bson.D{{{Key: "sender", Value: username}, {Key: "receiver", Value: host}},
			{{Key: "sender", Value: host}, {Key: "receiver", Value: username}}}}}, database.FindPvMessagesOption)
		if err != nil {
			fmt.Println("error in getting all pv messages is: ", err)
			c.JSON(500, gin.H{})
			return
		}

		for results.Next(ctx) {
			var document = &database.PvMessage{}
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
