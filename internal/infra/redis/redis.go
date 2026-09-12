package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

const BlacklistKeyPrefix = "blacklist:jti:"

func InitRedis() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Note: .env file not found or failed to load: %v", err)
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("Failed to parse REDIS_URL: %v", err)
			return
		}
		Client = redis.NewClient(opt)
	} else {
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			addr = "localhost:6379"
		}

		password := os.Getenv("REDIS_PASSWORD")
		dbStr := os.Getenv("REDIS_DB")
		dbNum := 0
		if dbStr != "" {
			if parsed, err := strconv.Atoi(dbStr); err == nil {
				dbNum = parsed
			}
		}

		Client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       dbNum,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis at %v: %v", Client.Options().Addr, err)
		return
	}

	log.Printf("Redis connection successful")
}

func GetClient() *redis.Client {
	return Client
}

func BlacklistToken(ctx context.Context, jti string, expiration time.Duration) error {
	if Client == nil || jti == "" {
		return nil
	}

	if expiration <= 0 {
		return nil
	}

	key := fmt.Sprintf("%s%s", BlacklistKeyPrefix, jti)
	return Client.Set(ctx, key, "revoked", expiration).Err()
}

func IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	if Client == nil || jti == "" {
		return false, nil
	}

	key := fmt.Sprintf("%s%s", BlacklistKeyPrefix, jti)
	count, err := Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
