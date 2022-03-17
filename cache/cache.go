package cache

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis"
)

var client *redis.Client
var Background = context.Background()

func SetWithNoExpire(key string, value string) (err error) {
	result := client.Set(Background, key, value, time.Duration(0))
	return result.Err()
}

func SetWithExpire(key string, value string, expireSecond uint32) (err error) {
	result := client.Set(Background, key, value, time.Duration(expireSecond)*time.Second)
	return result.Err()
}

func Get(key string) (value string) {
	result := client.Get(Background, key)
	return result.Val()
}

func Bell(channel string, value string) {
	client.Publish(Background, channel, value)
}

func Listen(channel string) (bareChannel <-chan string) {
	sub := client.Subscribe(Background, channel)
	bare := make(chan string)
	go func() {
		origin := sub.Channel()
		for {
			value := <-origin
			select {
			case bare <- value.Payload:
				continue
			case <-time.After(time.Duration(1) * time.Second):
				close(bare)
				sub.Close()
				log.Println("channel push timeout")
				return
			}
		}
	}()
	return bare
}

// init connect
func initClient() (err error) {
	client = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	_, err = client.Ping(context.TODO()).Result()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	err := initClient()
	if err != nil {
		log.Fatal("can not connect to redis server")
	}
}
