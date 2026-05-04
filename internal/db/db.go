package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewConnection creates and returns a database connection pool
func NewConnection() *pgxpool.Pool {
	// Note: In production, we will change this to .env file
	// For now, it matches your docker-compose.yml
	dsn := "postgres://admin:secretpassword@localhost:5433/nexaaudit??sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Ping the database to ensure the connection is actually alive
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database is not responding: %v\n", err)
	}

	fmt.Println("✅ Successfully connected to PostgreSQL!")
	return pool
}
