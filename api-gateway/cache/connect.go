package cache

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Connect() {
    Client = redis.NewClient(&redis.Options{
        Addr:        os.Getenv("REDIS_ADDR"), // localhost:6379
        Password:    "",
        DB:          0,
        DialTimeout: 5 * time.Second,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := Client.Ping(ctx).Err(); err != nil {
        log.Fatalf("Redis connection failed: %v", err)
    }

    log.Println("Connected to Redis")
}