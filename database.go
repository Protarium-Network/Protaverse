package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	ctx := context.Background()
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?pool_max_conns=10",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		getEnvDefault("DB_PORT", "5432"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = pgxpool.New(ctx, connStr)
	if err != nil {
		LogError("PostgreSQL Connection Error: " + err.Error())
		os.Exit(1)
	}

	if err := DB.Ping(ctx); err != nil {
		LogError("PostgreSQL Ping Error: " + err.Error())
		os.Exit(1)
	}

	LogSuccess("PostgreSQL Database connected!")
}

func getEnvDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
