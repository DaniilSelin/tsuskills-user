package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"tsuskills-user/config"
	router "tsuskills-user/internal/delivery/http"
	"tsuskills-user/internal/delivery/http/handler"
	"tsuskills-user/internal/infra/postgres"
	"tsuskills-user/internal/logger"
	"tsuskills-user/internal/repository"
	"tsuskills-user/internal/security"
	"tsuskills-user/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Logger
	appLogger, err := logger.New(&cfg.Logger.Logger)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	appLogger.Info(ctx, "Starting users service...")

	// Postgres
	pool, err := postgres.Connect(ctx, &cfg.Postgres)
	if err != nil {
		appLogger.Fatal(ctx, fmt.Sprintf("Failed to connect to Postgres: %v", err))
	}
	defer pool.Close()
	appLogger.Info(ctx, "Connected to PostgreSQL")

	// Migrations
	connString := cfg.Postgres.Pool.ConnConfig.ConnString()
	if err := postgres.RunMigrations(connString, cfg.Postgres.MigrationsPath); err != nil {
		appLogger.Fatal(ctx, fmt.Sprintf("Failed to run migrations: %v", err))
	}
	appLogger.Info(ctx, "Database migrations applied")

	// Dependencies
	userRepo := repository.NewUserRepository(pool)
	sec := security.NewSecurity(&cfg.JWT)
	userService := service.NewUserService(userRepo, sec, appLogger)
	h := handler.NewHandler(userService, appLogger)
	r := router.NewRouter(h, appLogger)

	// HTTP Server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		appLogger.Info(ctx, fmt.Sprintf("Received signal: %v, shutting down...", sig))

		shutCtx, shutCancel := context.WithTimeout(context.Background(), cfg.Server.ShutDownTimeOut)
		defer shutCancel()

		if err := httpServer.Shutdown(shutCtx); err != nil {
			appLogger.Error(ctx, fmt.Sprintf("Server shutdown error: %v", err))
		}
		cancel()
	}()

	appLogger.Info(ctx, fmt.Sprintf("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appLogger.Fatal(ctx, fmt.Sprintf("Server failed: %v", err))
	}

	appLogger.Info(ctx, "Server stopped")
}
