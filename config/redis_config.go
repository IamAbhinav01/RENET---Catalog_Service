package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(redis_url string) *redis.Client {
	if redis_url == "" {
		fmt.Println("Warning: REDIS_ADDR not configured. Running without Redis cache.")
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         redis_url,
		Password:     "",
		DB:           0,
		Protocol:     2,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("Warning: Redis unavailable at %s (%v). Running without cache invalidation.\n", redis_url, err)
		return nil
	}

	fmt.Printf("Connected to Redis successfully at %s.\n", redis_url)
	return rdb
}
	
