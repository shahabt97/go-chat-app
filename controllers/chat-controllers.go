package controllers

import (
	"context"
	// "encoding/json"
	"first/database"
	"first/elasticsearch"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	// "github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PublicChatHandler(c *gin.Context) {
	c.File("views/public-chat.html")
}

func PvChatHandler(c *gin.Context) {
	c.File("views/pv-chat.html")
}

// func GetPubMessages(conn *websocket.Conn) {
// 	// if status == "public" {
// 		var Array []*database.PublicMessage
// 		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 		defer cancel()
// 		results, err := database.PubMessages.Find(ctx, bson.M{}, database.FindPubMessagesBasedOnCreatedAtIndexOption)

// 		if err != nil {
// 			fmt.Println("error in getting all public messages is: ", err)
// 			// c.JSON(500, gin.H{})
// 			return
// 		}

// 		for results.Next(ctx) {
// 			var document = &database.PublicMessage{}
// 			if err := results.Decode(document); err != nil {
// 				fmt.Println("error in reading all results of public messages: ", err)
// 				// c.JSON(500, gin.H{})
// 				return
// 			}
// 			Array = append(Array, document)
// 		}

// 		allMessages := &websocketServer.AllPubMessages{
// 			EventName: "all messages",
// 			Data: Array,
// 		}

// 		jsonData, errOfMarshaling := json.Marshal(allMessages)

// 		if errOfMarshaling != nil {
// 			fmt.Println("error in Marshaling public messages: ", err)
// 			// c.JSON(500, gin.H{})
// 			return
// 		}
// 		// if len(Array) != 0 {
// 			if err :=conn.WriteMessage(websocket.TextMessage,jsonData); err != nil{
// 				conn.Close()
// 				return
// 			// }
// 		} 
		// else {
		// 	c.JSON(201, []gin.H{})
		// }

	// } 
	// else if c.Query("status") == "pv" {
	// 	var Array []*database.PvMessage
	// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()
	// 	host := c.Query("hostUser")
	// 	username := c.Query("username")
	// 	results, err := database.PvMessages.Find(ctx, bson.D{{Key: "$or", Value: []bson.D{{{Key: "sender", Value: username}, {Key: "receiver", Value: host}},
	// 		{{Key: "sender", Value: host}, {Key: "receiver", Value: username}}}}}, database.FindPvMessagesOption)
	// 	if err != nil {
	// 		fmt.Println("error in getting all pv messages is: ", err)
	// 		c.JSON(500, gin.H{})
	// 		return
	// 	}

	// 	for results.Next(ctx) {
	// 		var document = &database.PvMessage{}
	// 		if err := results.Decode(document); err != nil {
	// 			fmt.Println("error in reading all results of public messages: ", err)
	// 			c.JSON(500, gin.H{})
	// 			return
	// 		}
	// 		Array = append(Array, document)
	// 	}
	// 	if len(Array) != 0 {
	// 		c.JSON(201, Array)
	// 		return
	// 	} else {
	// 		c.JSON(201, []gin.H{})
	// 	}

	// }

// }

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

func SearchInMongoForPvMes(c *gin.Context) {

	query := c.Query("q")
	user := c.Query("user")
	host := c.Query("host")
	var Array []*database.PvMessage

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	patternOfUser := fmt.Sprintf(`^%s$`, user)
	patternOfHost := fmt.Sprintf(`^%s$`, host)
	patternOfMessage := query

	// Create a regex query
	regexOfMessage := bson.M{"$regex": primitive.Regex{Pattern: patternOfMessage, Options: "i"}}
	regexOfUser := bson.M{"$regex": primitive.Regex{Pattern: patternOfUser, Options: "i"}}
	regexOfHost := bson.M{"$regex": primitive.Regex{Pattern: patternOfHost, Options: "i"}}

	filter1 := bson.M{"message": regexOfMessage, "sender": regexOfUser, "receiver": regexOfHost}
	filter2 := bson.M{"message": regexOfMessage, "sender": regexOfHost, "receiver": regexOfUser}

	combinedFilter := bson.M{"$or": []bson.M{filter1, filter2}}
	// Perform the find operation
	cur, err := database.PvMessages.Find(ctx, combinedFilter)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	// fmt.Println("cur: ", cur)
	// Iterate through the results
	// var results []*database.PvMessage
	for cur.Next(ctx) {

		var document = &database.PvMessage{}
		if err := cur.Decode(document); err != nil {
			fmt.Println("error in reading  results of searched pv messages: ", err)
			c.JSON(500, gin.H{})
			return
		}
		// fmt.Println("doc: ", document)
		Array = append(Array, document)
	}
	// Print the results
	// for _, result := range results {
	fmt.Println(len(Array))
	// }

	c.JSON(201, Array)

}
