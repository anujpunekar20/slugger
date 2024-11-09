package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

const keyPrefix = "urlshortener:"

func shortenURL(rdb *redis.Client, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty URL provided")
	}

	hash := sha1.New()
	hash.Write([]byte(url))
	shortURL := base64.URLEncoding.EncodeToString(hash.Sum(nil))[:8]

	key := keyPrefix + shortURL
	log.Printf("Generated short URL: %s for original URL: %s", shortURL, url)
	log.Printf("Redis key: %s", key)

	// Store with expiration (optional, remove 0 and add time if you want URLs to expire)
	err := rdb.Set(ctx, key, url, 0).Err()
	if err != nil {
		log.Printf("Redis SET error: %v", err)
		return "", err
	}

	// Verify the storage immediately
	storedURL, err := rdb.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Redis GET verification error: %v", err)
		return "", err
	}

	if storedURL != url {
		log.Printf("Storage verification failed. Stored: %s, Expected: %s", storedURL, url)
		return "", fmt.Errorf("storage verification failed")
	}

	log.Printf("Successfully stored URL. Key: %s, Value: %s", key, storedURL)
	return shortURL, nil
}

func redirectToURL(rdb *redis.Client, shortURL string) (string, error) {
	key := keyPrefix + shortURL
	log.Printf("Attempting to retrieve from Redis - Key: %s", key)

	url, err := rdb.Get(ctx, keyPrefix+shortURL).Result()
	if err != nil {
		log.Printf("Redis retrieval error: %v", err)
		return "", err
	}

	log.Printf("Found URL in Redis: %s", url)
	return url, nil
}

func addCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}