package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"tsuskills-dbmanager/config"
	router "tsuskills-dbmanager/internal/delivery/http"
	"tsuskills-dbmanager/internal/delivery/http/handler"
	"tsuskills-dbmanager/internal/infra/migrations"
	"tsuskills-dbmanager/internal/infra/opensearch"
	"tsuskills-dbmanager/internal/logger"
	"tsuskills-dbmanager/internal/search"
	"tsuskills-dbmanager/internal/service"

	_ "github.com/davecgh/go-spew/spew"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FAILED: error when load config: %v", err)
	}

	appLogger, err := logger.New(&cfg.Logger.Logger)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	appLogger.Info(ctx, "Starting application...")

	osclient, err := opensearch.NewClient(cfg.Search.Client)
	if err != nil {
		log.Fatalf("FAILED: error when connect to opensearch: %v", err)
	}

	err = opensearch.Ping(osclient, cfg.Search.Connect)
	if err != nil {
		log.Fatalf("FAILED: error when ping opensearch: %v", err)
	}

	migratorOS := opensearch.NewMigrate(cfg.Migrate)
	migratorOS.WithClient(osclient)

	logFunc := func(level string, msg string) {
		switch level {
		case "INFO":
			log.Println(msg)
		default:
			log.Println(msg)
		}
	}
	err = migrations.RunMigrations(ctx, logFunc, cfg.Migrate, &migratorOS)
	if err != nil {
		log.Fatal("Error when migrate: %w", err)
	}

	log.Println("OpenSearch client connected and migrations applied successfully.")

	searchVC := search.NewVacancySearch(cfg, osclient, appLogger)
	service := service.NewVacancyService(cfg, searchVC, appLogger)
	handler := handler.NewHandler(service, appLogger)
	r := router.NewRouter(handler, appLogger)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Запускаем сервер
	appLogger.Info(ctx, fmt.Sprintf("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appLogger.Fatal(ctx, fmt.Sprintf("Server failed to start: %v", err))
	}
}
