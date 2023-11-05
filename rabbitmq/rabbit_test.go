package rabbitmq

import (
	"go-chat-app/config"
	"go-chat-app/database"
	"go-chat-app/elasticsearch"
	redisServer "go-chat-app/redis"
	"math/rand"
	"testing"
	"time"
)

func TestMain(t *testing.T) {

	if err := config.EnvExtractor(); err != nil {
		t.Errorf("error in extracting env: %v\n", err)
	}
	if err := elasticsearch.Init(); err != nil {
		t.Fatalf("error in initating Elastic: %v\n", err)
	}

	if err := redisServer.Init(); err != nil {
		t.Fatalf("error in connecting to Redis: %v\n", err)
	}

	if err := database.UtilsInitializations(); err != nil {
		t.Fatalf("error in connecting to database: %v\n", err)
	}

	publisher := &PubMessagePublishingMaster{}
	consumer := &PubMessageConsumerMaster{}

	err := RabbitMQInitialization(publisher, consumer)
	time.Sleep(1 * time.Second)
	if err != nil {
		t.Errorf("error in starting rabbit: %v\n", err)
	}

	messages := PubMessageEventCreator(1000000)

	for _, message := range *messages {
		time.Sleep(1 * time.Millisecond)
		go PubMessagesPublisher(message, publisher)
	}

	select {}

}

func PubMessageEventCreator(size int) *[]*Event {

	var events []*Event

	for i := 0; i < size; i++ {

		var event Event

		event.Data.Message = GenerateRandomString(40)
		event.EventName = "chat message"
		event.Data.Username = GenerateRandomString(30)
		event.Data.Timestamp = time.Now()

		events = append(events, &event)

	}

	return &events

}

// generateRandomString returns a random string of length n
func GenerateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
