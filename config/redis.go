package config

import (
	"log"

	"github.com/redis/go-redis/v9" // Ensure you're using v9
	"context"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// For go-redis v9: use Ping(ctx)
	if _, err := RedisClient.Ping(Ctx).Result(); err != nil {
		log.Fatalf("Redis init failed: %v", err)
	}
}
