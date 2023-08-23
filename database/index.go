package database

import (
	"context"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientOptions = options.Client().ApplyURI("mongodb://localhost:27017")

var client, _ = mongo.Connect(context.Background(), clientOptions)

var chat = client.Database("chat")
var Users = chat.Collection("users")
var PvMessages = chat.Collection("pv-messages")
var PubMessages = chat.Collection("public-messages")
