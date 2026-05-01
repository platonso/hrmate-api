package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/platonso/hrmate-api/internal/config"
	"github.com/platonso/hrmate-api/internal/handler"
	"github.com/platonso/hrmate-api/internal/repository/postgres"
	"github.com/platonso/hrmate-api/internal/service/auth"
	"github.com/platonso/hrmate-api/internal/service/form"
	"github.com/platonso/hrmate-api/internal/service/user"
	"github.com/platonso/hrmate-api/internal/storage/s3"
)

type Application struct {
	config *config.Config
	repo   *postgres.Repository
	server *http.Server
}

func New(ctx context.Context, cfg *config.Config) (*Application, error) {
	postgresRepo, txMgr, err := postgres.NewRepository(ctx, cfg.Postgres.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	if err := postgresRepo.Users.CheckSchema(ctx); err != nil {
		postgresRepo.Close()
		return nil, fmt.Errorf("database schema not ready: %w\n💡 Run: make migrate-up", err)
	}

	storage, err := s3.New(ctx, &cfg.MinIO)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize s3 storage: %w", err)
	}

	userSvc := user.NewService(postgresRepo.Users)
	authSvc := auth.NewService(txMgr, postgresRepo.Users, cfg.JWTSecret)
	formSvc := form.NewService(txMgr, postgresRepo.Forms, postgresRepo.Users, postgresRepo.Documents, storage)

	if err := authSvc.ImplementAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		postgresRepo.Close()
		return nil, fmt.Errorf("failed to implement admin: %w", err)
	}

	router := handler.NewRouter(cfg, authSvc, userSvc, formSvc)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      router.Routes(),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	app := &Application{
		config: cfg,
		repo:   postgresRepo,
		server: srv,
	}

	return app, nil
}

func (app *Application) Start(errChan chan<- error) {
	log.Printf("Starting server on port %s", app.config.HTTP.Port)

	if err := app.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- err
	}
}

func (app *Application) Stop(ctx context.Context) error {
	log.Println("Initiating graceful shutdown...")
	var shutdownErr error

	if app.server != nil {
		log.Println("Shutting down HTTP server...")
		if err := app.server.Shutdown(ctx); err != nil {
			shutdownErr = fmt.Errorf("server shutdown failed: %w", err)
			log.Printf("Error during server shutdown: %v", err)
		}
	}

	log.Println("Closing database connection...")
	app.repo.Close()

	if shutdownErr == nil {
		log.Println("Graceful shutdown completed successfully")
	}

	return shutdownErr
}
