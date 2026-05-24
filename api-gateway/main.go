package main

import (
	"log"

	"github.com/ilovecroissant/ai-marketing-engine/cache"
	"github.com/ilovecroissant/ai-marketing-engine/db"
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

	log.Println("api-gateway ready")
}
