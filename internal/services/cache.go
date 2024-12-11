package services

import (
	"context"
	"sim-server/database"
	"time"
)

func Set(key string, message []byte) error {
	return database.RedisClient.Set(context.Background(), key, message, 0*time.Minute).Err()
}

// CheckAndGetKey to check if a key exists in Redis
func CheckAndGetKey(key string) (string, bool) {
	// Use EXISTS to check if the key exists
	exists, err := database.RedisClient.Exists(context.Background(), key).Result()
	if err != nil {
		return "", false
	}
	// If key does not exist, return an empty string
	if exists == 0 {
		return "", false
	}

	// Key exists, so fetch its value using GET
	value, err := database.RedisClient.Get(context.Background(), key).Result()
	if err != nil {
		return "", false
	}

	return value, true
}
