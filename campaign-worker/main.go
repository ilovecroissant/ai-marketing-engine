package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/ilovecroissant/ai-marketing-engine/campaign-worker/consumer"
    "github.com/ilovecroissant/ai-marketing-engine/campaign-worker/processor"
    "github.com/joho/godotenv"
)

func main() {
    if err := godotenv.Load("../.env"); err != nil {
        log.Println("No .env file found, relying on shell environment")
    }

    // Graceful shutdown — wait for CTRL+C
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        sig := make(chan os.Signal, 1)
        signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
        <-sig
        log.Println("Shutdown signal received")
        cancel()
    }()

    c := consumer.New()
    defer c.Close()

    // Start blocking consumer loop
    c.Start(ctx, processor.HandleCampaign)
}