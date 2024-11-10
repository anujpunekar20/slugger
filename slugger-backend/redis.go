package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func NewRedisClient() *redis.Client {
	redisURL := os.Getenv("HEROKU_REDIS_CYAN_TLS_URL")
	fmt.Println(redisURL) // Get the Redis URL Heroku provides
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable not set")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	// Optionally enable TLS if needed
	opt.TLSConfig = &tls.Config{
		InsecureSkipVerify: true, // If you want to skip the verification
	}

	rdb := redis.NewClient(opt)

	// Test connection
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
	log.Printf("Redis connected successfully: %s", pong)

	return rdb
}
