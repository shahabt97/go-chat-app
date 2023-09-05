package database

import (
	"context"
	"first/hosts"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientOptions = options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", hosts.MongoHost))

var client, _ = mongo.Connect(context.Background(), clientOptions)

var chat = client.Database("chat")
var Users = chat.Collection("users")
var PvMessages = chat.Collection("pv-messages")
var PubMessages = chat.Collection("public-messages")
var FindPubMessagesOption *options.FindOptions
var FindPvMessagesOption *options.FindOptions

func UtilsInitializations() {

	// public messages indexes
	PubMessages.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "CreatedAt", Value: 1}},
		Options: options.Index(),
	})

	// pv messages indexes
	PvMessages.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "CreatedAt", Value: 1}},
		Options: options.Index().SetName("createdDateIndex"),
	})
	PvMessages.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "sender", Value: 1}, {Key: "receiver", Value: 1}},
		Options: options.Index().SetName("senderReceiverIndex"),
	})

	// users indexes
	Users.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	Users.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	FindPubMessagesOption = options.Find().SetSort(bson.D{{Key: "CreatedAt", Value: 1}})
	FindPvMessagesOption = options.Find().SetHint("senderReceiverIndex")
}

type PublicMessage struct {
	Id        string      `bson:"_id" json:"id"`
	Message   string      `bson:"message" json:"message"`
	Sender    string `bson:"sender" json:"sender"`
	CreatedAt time.Time   `bson:"createdAt" json:"createdAt"`
}

type PvMessage struct {
	Message   string    `bson:"message" json:"message"`
	Sender    string    `bson:"sender" json:"sender"`
	Receiver  string    `bson:"receiver" json:"receiver"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
