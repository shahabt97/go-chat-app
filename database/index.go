package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientOptions = options.Client().ApplyURI("mongodb://localhost:27017")

var client, _ = mongo.Connect(context.Background(), clientOptions)

var chat = client.Database("chat")
var Users = chat.Collection("users")
var PvMessages = chat.Collection("pv-messages")
var PubMessages = chat.Collection("public-messages")
var FindPubMessagesBasedOnCreatedAtIndexOption *options.FindOptions

func UtilsInitializations() {
	PubMessages.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "CreatedAt", Value: 1}},
		Options: options.Index(),
	})
	FindPubMessagesBasedOnCreatedAtIndexOption = options.Find().SetSort(bson.D{{Key: "CreatedAt", Value: 1}})
}

type PublicMessage struct {
	Id        string      `bson:"_id" json:"id"`
	Message   string      `bson:"message" json:"message"`
	Sender    UsersSchema `bson:"sender" json:"sender"`
	CreatedAt time.Time   `bson:"CreatedAt" json:"createdAt"`
}
type UsersSchema struct {
	Id       primitive.ObjectID `bson:"id" json:"id"`
	Username string             `bson:"username" json:"username"`
}
