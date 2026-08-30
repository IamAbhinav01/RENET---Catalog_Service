package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func connectRedis(redis_url string){
	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_url,
		Password: "", 
		DB:       0,  
		Protocol: 2,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err();err!=nil{
		fmt.Printf("Warning: Redis unavailable at %s (%v). Running without cache invalidation.", redis_url, err)
		return
	}
	fmt.Println("Connected to Redis successfully.")
}
	
