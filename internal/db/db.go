package db

import (
	"context"
	"fmt"
	"log"
	"os" // NEW: We need the os package to read environment variables

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnection() *pgxpool.Pool {
	// 1. Grab the connection string from the .env file
	dsn := os.Getenv("DATABASE_URL")

	// 2. Add a safety fallback just in case the .env file is missing
	if dsn == "" {
		fmt.Println("⚠️ DATABASE_URL not found in environment, falling back to default...")
		dsn = "postgres://nexa_admin:supersecretpassword123@localhost:5434/nexaaudit?sslmode=disable"
	}

	// 3. Connect to the database
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// 4. Ping to verify the connection is actually alive
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database is not responding: %v\n", err)
	}

	fmt.Println("✅ Successfully connected to PostgreSQL!")
	return pool
}
