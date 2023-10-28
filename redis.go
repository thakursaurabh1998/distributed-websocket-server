package main

import "github.com/go-redis/redis"

type Redis struct {
	Client *redis.Client
}

func NewRedis(address string) *Redis {
	client := redis.NewClient(&redis.Options{Addr: address})

	_, err := client.Ping().Result()

	if err != nil {
		panic(err)
	}

	return &Redis{Client: client}
}
