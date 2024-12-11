package database

import (
	"context"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"log"
	"sim-server/config"
	"sync"
)

var RedisClient *redis.Client
var redisDefaultOnce sync.Once

var redisCache *cache.Cache
var redisCacheOnce sync.Once

func GetRedisDefaultClient(cfg config.Config) *redis.Client {
	redisDefaultOnce.Do(func() {
		redisClientOptions, err := redis.ParseURL(cfg.RedisUri)
		if err != nil {
			log.Println("Error parsing Redis URI")
			panic(err)
		}
		RedisClient = redis.NewClient(redisClientOptions)
	})
	return RedisClient
}

func CheckRedisConnection(cfg config.Config) *redis.Client {
	redisClient := GetRedisDefaultClient(cfg)
	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Println("Error Connecting to Redis!")
		panic(err)
	}
	log.Print("Redis Address: " + redisClient.Options().Addr)
	log.Print(cfg.RedisUri)
	log.Println("Connected to Redis!")
	return redisClient
}
