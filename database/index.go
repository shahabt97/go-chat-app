package database

import (
	"context"
	"go-chat-app/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientOptions *options.ClientOptions

var Client *mongo.Client
var ErrOfDB error

var chat *mongo.Database
var Users *mongo.Collection
var PvMessages *mongo.Collection
var PubMessages *mongo.Collection
var FindPubMessagesOption *options.FindOptions
var FindPvMessagesOption *options.FindOptions

func UtilsInitializations() error {

	// configurations

	clientOptions = options.Client().ApplyURI(config.ConfigData.MongoURI)

	Client, ErrOfDB = mongo.Connect(context.Background(), clientOptions)

	if ErrOfDB != nil {
		return ErrOfDB
	}

	chat = Client.Database("chat")
	Users = chat.Collection("users")
	PvMessages = chat.Collection("pv-messages")
	PubMessages = chat.Collection("public-messages")

	// public messages indexes
	PubMessages.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "message", Value: 1}},
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

	// options for quering in database
	FindPubMessagesOption = options.Find().SetSort(bson.D{{Key: "message", Value: 1}})
	FindPvMessagesOption = options.Find().SetHint("senderReceiverIndex")

	return nil
}

type PublicMessage struct {
	Message   string    `bson:"message" json:"message"`
	Sender    string    `bson:"sender" json:"sender"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

type PvMessage struct {
	Message   string    `bson:"message" json:"message"`
	Sender    string    `bson:"sender" json:"sender"`
	Receiver  string    `bson:"receiver" json:"receiver"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
