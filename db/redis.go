package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

const (
	redisConnectionErrorMessage   = "Error while connecting to Redis: %v"
	redisConnectionSuccessMessage = "Connection with Redis established successfully"
)

func ConnectRedis(ctx context.Context) {
	db := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := db.Ping(ctx).Result()
	if err != nil {
		log.Println(fmt.Errorf(redisConnectionErrorMessage, err))
	} else {
		log.Println(redisConnectionSuccessMessage)
	}

	Redis = db
}
