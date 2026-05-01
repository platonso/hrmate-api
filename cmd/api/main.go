package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/platonso/hrmate-api/internal/app"
	"github.com/platonso/hrmate-api/internal/config"
)

// @title           HRMate API
// @version         1.0
// @description     RESTful API для приложения HRMate. Управление пользователями, авторизацией и формами.

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Printf("Config error: %v", err)
		os.Exit(1)
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(sigCtx, cfg)
	if err != nil {
		log.Printf("Failed to init api: %v", err)
		os.Exit(1)
	}

	errChan := make(chan error, 1)
	go a.Start(errChan)

	select {
	case err := <-errChan:
		log.Printf("Server stopped with error: %v", err)
		os.Exit(1)

	case <-sigCtx.Done():
		log.Println("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := a.Stop(shutdownCtx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			os.Exit(1)
		}
	}
}
