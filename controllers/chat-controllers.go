package controllers

import (
	"context"
	"first/database"
	"first/elasticsearch"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PublicChatHandler(c *gin.Context) {
	c.File("views/public-chat.html")
}

func PvChatHandler(c *gin.Context) {
	c.File("views/pv-chat.html")
}

func SearchInPubChat(c *gin.Context) {

	query := c.Query("q")
	IDs, err := elasticsearch.Client.SearchPubMessages(query, "pubmessages")
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Server error",
		})
		return
	}

	var Array []*database.PublicMessage
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := database.PubMessages.Find(ctx, bson.M{"_id": bson.M{"$in": IDs}}, database.FindPubMessagesBasedOnCreatedAtIndexOption)
	if err != nil {
		fmt.Println("error in getting searched public messages is: ", err)
		c.JSON(500, gin.H{})
		return
	}

	for results.Next(ctx) {
		var document = &database.PublicMessage{}
		if err := results.Decode(document); err != nil {
			fmt.Println("error in reading  results of searched public messages: ", err)
			c.JSON(500, gin.H{})
			return
		}
		Array = append(Array, document)
	}

	if len(Array) != 0 {
		c.JSON(201, Array)
		return
	} else {
		c.JSON(201, []gin.H{})
	}
}

func SearchInPvChat(c *gin.Context) {
	query := c.Query("q")
	user := c.Query("user")
	host := c.Query("host")

	IDs, err := elasticsearch.Client.SearchPvMessages(query, user, host, "pv-messages")
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Server error",
		})
		return
	}

	var Array []*database.PvMessage
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := database.PvMessages.Find(ctx, bson.M{"_id": bson.M{"$in": IDs}}, database.FindPvMessagesOptionWithoutHint)
	if err != nil {
		fmt.Println("error in getting searched pv messages is: ", err)
		c.JSON(500, gin.H{})
		return
	}

	for results.Next(ctx) {
		var document = &database.PvMessage{}
		if err := results.Decode(document); err != nil {
			fmt.Println("error in reading  results of searched pv messages: ", err)
			c.JSON(500, gin.H{})
			return
		}
		Array = append(Array, document)
	}

	if len(Array) != 0 {
		c.JSON(201, Array)
		return
	} else {
		c.JSON(201, []gin.H{})
	}
}
func SearchAllMessages(c *gin.Context) {

	query := c.Query("q")
	res, err := elasticsearch.Client.SearchAllMessages(query, "pv-messages", "pubmessages")
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{
			"message": "Server error",
		})
		return
	}

	var Array []interface{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	var PubIDs []primitive.ObjectID
	var PvIDs []primitive.ObjectID

	for _, data := range res {
		switch data.Index {
		case "pv-messages":
			PvIDs = append(PvIDs, data.Id)
		case "pubmessages":
			PubIDs = append(PubIDs, data.Id)
		default:
			continue
		}
	}

	if len(PubIDs) != 0 {
		PubResults, err := database.PubMessages.Find(ctx, bson.M{"_id": bson.M{"$in": PubIDs}}, database.FindPubMessagesBasedOnCreatedAtIndexOption)
		if err != nil {
			fmt.Println("error in getting searched public messages is: ", err)
			c.JSON(500, gin.H{})
			return
		} else {
			for PubResults.Next(ctx) {
				var document = &database.PublicMessage{}
				if err := PubResults.Decode(document); err != nil {
					fmt.Println("error in reading  results of searched public messages: ", err)
					c.JSON(500, gin.H{})
					return
				}
				Array = append(Array, document)
			}
		}
	}

	if len(PvIDs) != 0 {
		PvResults, err := database.PvMessages.Find(ctx, bson.M{"_id": bson.M{"$in": PvIDs}}, database.FindPubMessagesBasedOnCreatedAtIndexOption)
		if err != nil {
			fmt.Println("error in getting searched pv messages is: ", err)
			c.JSON(500, gin.H{})
			return
		} else {
			for PvResults.Next(ctx) {
				var document = &database.PvMessage{}
				if err := PvResults.Decode(document); err != nil {
					fmt.Println("error in reading  results of searched pv messages: ", err)
					c.JSON(500, gin.H{})
					return
				}
				Array = append(Array, document)
			}
		}
	}

	if len(Array) != 0 {
		c.JSON(201, Array)
		return
	} else {
		c.JSON(201, []gin.H{})
	}

}
// func SearchInMongoForPvMes(c *gin.Context) {

// 	query := c.Query("q")
// 	user := c.Query("user")
// 	host := c.Query("host")
// 	var Array []*database.PvMessage

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	patternOfUser := fmt.Sprintf(`^%s$`, user)
// 	patternOfHost := fmt.Sprintf(`^%s$`, host)
// 	patternOfMessage := query

// 	// Create a regex query
// 	regexOfMessage := bson.M{"$regex": primitive.Regex{Pattern: patternOfMessage, Options: "i"}}
// 	regexOfUser := bson.M{"$regex": primitive.Regex{Pattern: patternOfUser, Options: "i"}}
// 	regexOfHost := bson.M{"$regex": primitive.Regex{Pattern: patternOfHost, Options: "i"}}

// 	filter1 := bson.M{"message": regexOfMessage, "sender": regexOfUser, "receiver": regexOfHost}
// 	filter2 := bson.M{"message": regexOfMessage, "sender": regexOfHost, "receiver": regexOfUser}

// 	combinedFilter := bson.M{"$or": []bson.M{filter1, filter2}}
// 	// Perform the find operation
// 	cur, err := database.PvMessages.Find(ctx, combinedFilter)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer cur.Close(ctx)

// 	// fmt.Println("cur: ", cur)
// 	// Iterate through the results
// 	// var results []*database.PvMessage
// 	for cur.Next(ctx) {

// 		var document = &database.PvMessage{}
// 		if err := cur.Decode(document); err != nil {
// 			fmt.Println("error in reading  results of searched pv messages: ", err)
// 			c.JSON(500, gin.H{})
// 			return
// 		}
// 		// fmt.Println("doc: ", document)
// 		Array = append(Array, document)
// 	}
// 	// Print the results
// 	// for _, result := range results {
// 	fmt.Println(len(Array))
// 	// }

// 	c.JSON(201, Array)

// }
