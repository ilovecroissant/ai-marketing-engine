package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),     // admin
		os.Getenv("DB_PASSWORD"), // secret
		os.Getenv("DB_HOST"),     // localhost
		os.Getenv("DB_PORT"),     // 5432
		os.Getenv("DB_NAME"),     // marketing_engine
	)

	var err error
	Pool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Verify connection
	if err = Pool.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}

	log.Println("Connected to PostgreSQL")
}
