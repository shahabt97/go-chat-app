package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientOptions = options.Client().ApplyURI("mongodb://localhost:27017")

var client, _ = mongo.Connect(context.Background(), clientOptions)

var chat = client.Database("chat")
var Users = chat.Collection("users")
var PvMessages = chat.Collection("pv-messages")
var PubMessages = chat.Collection("public-messages")

type PublicMessage struct {
	Id        string      `bson:"_id" json:"id"`
	Message   string      `bson:"message" json:"message"`
	Sender    UsersSchema `bson:"sender" json:"sender"`
	CreatedAt time.Time   `bson:"CreatedAt" json:"createdAt"`
}
type UsersSchema struct {
	Id       string `bson:"id" json:"id"`
	Username string `bson:"username" json:"username"`
}
