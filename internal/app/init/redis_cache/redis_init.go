package rediscache

import "github.com/redis/go-redis/v9"

func InitRed() *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "my-password", // password set
		DB:       0,             // use default DB
	})

	return client
}
