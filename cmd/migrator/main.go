package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/platonso/hrmate-api/internal/config"
	"github.com/pressly/goose/v3"
)

var command = flag.String("command", "up", "goose command (up, down, status)")

func main() {
	flag.Parse()

	if *command == "" {
		log.Fatalf("Error: -command flag is required")
	}

	cfg, err := config.NewDB()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	if _, err := os.Stat(cfg.MigrationDir); os.IsNotExist(err) {
		log.Fatalf("Migrations directory does not exist")
	}

	db := connectDB(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	switch *command {
	case "up":
		if err := goose.Up(db, cfg.MigrationDir); err != nil {
			log.Fatalf("failed to run up: %v", err)
		}
	case "down":
		if err := goose.Down(db, cfg.MigrationDir); err != nil {
			log.Fatalf("failed to run down: %v", err)
		}
	case "status":
		if err := goose.Status(db, cfg.MigrationDir); err != nil {
			log.Fatalf("failed to run status: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", *command)
	}
}

func connectDB(cfg *config.PostgresConfig) *sql.DB {
	db, err := sql.Open("pgx", cfg.GetDSN())
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	return db
}
