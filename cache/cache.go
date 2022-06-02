package cache

import (
	"log"
	"time"

	"github.com/go-redis/redis"
)

var client *redis.Client

func SetWithNoExpire(key string, value string) (err error) {
	result := client.Set(key, value, time.Duration(0))
	return result.Err()
}

func SetWithExpire(key string, value string, expireSecond uint32) (err error) {
	result := client.Set(key, value, time.Duration(expireSecond)*time.Second)
	return result.Err()
}

func Get(key string) (value string) {
	result := client.Get(key)
	return result.Val()
}

func Bell(channel string, value string) {
	client.Publish(channel, value)
}

func Listen(channel string) (bareChannel <-chan string) {
	sub := client.Subscribe(channel)
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
				log.Println("channel push timeout! closed")
				return
			}
		}
	}()
	return bare
}

// init connect
func initClient() (err error) {
	client = redis.NewClient(&redis.Options{
		Addr: "note.redis.local:6379",
		DB:   1,
	})

	_, err = client.Ping().Result()
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
