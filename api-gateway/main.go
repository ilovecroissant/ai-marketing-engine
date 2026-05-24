package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/ilovecroissant/ai-marketing-engine/cache"
    "github.com/ilovecroissant/ai-marketing-engine/db"
    "github.com/ilovecroissant/ai-marketing-engine/handlers"
    appkafka "github.com/ilovecroissant/ai-marketing-engine/kafka"
    "github.com/ilovecroissant/ai-marketing-engine/middleware"
    "github.com/joho/godotenv"
)

func main() {
    if err := godotenv.Load("../.env"); err != nil {
        if err := godotenv.Load(".env"); err != nil {
            log.Println("No .env file found, relying on shell environment")
        }
    }

    db.Connect()
    defer db.Pool.Close()

    cache.Connect()
    defer cache.Client.Close()

    appkafka.Init()
    defer appkafka.Close()

    r := gin.Default()
    r.Use(middleware.RateLimit())

    // Public routes
    auth := r.Group("/auth")
    {
        auth.POST("/register", handlers.Register)
        auth.POST("/login", handlers.Login)
    }

    // Protected routes
    api := r.Group("/api")
    api.Use(middleware.Auth())
    {
        api.POST("/campaigns", handlers.CreateCampaign)
        api.GET("/campaigns/:id", handlers.GetCampaign)
    }

    log.Println("api-gateway ready on :8081")
    r.Run(":8081") // 8080 is taken by kafka-ui
}