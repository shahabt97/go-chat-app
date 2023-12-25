package rabbitmq

import (
	"fmt"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	rabbit "github.com/shahabt97/rabbit-pool/v3"
)

func TestPool(t *testing.T) {

	// conn, err := amqp.Dial("amqp://shahab:83000000@localhost:5672/my_chat_app")
	// if err != nil {
	// 	t.Errorf("error in connecting to rabbitMQ: %v\n", err)
	// }

	url := "amqp://shahab:83000000@localhost:5672/my_chat_app"
	pool := rabbit.NewPool(url, 4, 20, 15)

	arr := GetFromPool(pool)
	go PutInPool(pool, arr)
	arr2 := GetFromPool(pool)
	go PutInPool(pool, arr2)

	time.Sleep(20 * time.Second)

	pool.Mu.Lock()
	fmt.Println("length of channels afterrrrr is: ", pool.IdleNum)
	fmt.Println("IdleConn is: ", pool.IdleNum)
	pool.Mu.Unlock()

	select {}
}

func GetFromPool(pool *rabbit.ConnectionPool) *[]*amqp.Channel {

	var array []*amqp.Channel
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 60; i++ {
		wg.Add(1)
		// time.Sleep(10 * time.Millisecond)
		go func(arr *[]*amqp.Channel, mutex *sync.Mutex) {
			mutex.Lock()
			*arr = append(*arr, pool.Get())
			mutex.Unlock()
			wg.Done()
		}(&array, &mu)

	}

	wg.Wait()

	pool.Mu.Lock()
	fmt.Println("length of channels before is: ", pool.IdleNum)
	pool.Mu.Unlock()

	fmt.Println("length of array of channels is: ", len(array))

	return &array

}

func PutInPool(pool *rabbit.ConnectionPool, arr *[]*amqp.Channel) {
	// time.Sleep(18 * time.Second)
	var wg sync.WaitGroup

	st := time.Now()
	for _, ch := range *arr {
		wg.Add(1)
		// time.Sleep(10 * time.Millisecond)
		go func(channel *amqp.Channel) {
			pool.Put(channel)
			wg.Done()
		}(ch)
	}
	wg.Wait()
	fmt.Println("Time is: ", time.Since(st))
	pool.Mu.Lock()
	fmt.Println("length of channels after is: ", pool.IdleNum)
	pool.Mu.Unlock()

}
