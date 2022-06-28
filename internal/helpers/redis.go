package helpers

import (
	"github.com/go-redis/redis/v8"
	"github.com/initialed85/uneventful/internal/constants"
)

func GetRedisClient() (redisClient *redis.Client, err error) {
	redisURL, err := GetEnvironmentVariable("REDIS_URL", false, constants.DefaultRedisURL)
	if err != nil {
		return nil, err
	}

	redisClient = redis.NewClient(&redis.Options{Addr: redisURL})

	return redisClient, nil
}
