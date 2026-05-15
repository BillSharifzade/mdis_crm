package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crm_backend/internal/api"
	"crm_backend/internal/config"
	"crm_backend/internal/migrations"
	"crm_backend/pkg/database"
)

// @title MDIS CRM API
// @version 1.0
// @description MDIS CRM Backend API
// @version 1.0
// @BasePath /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var db *database.DB
	if cfg.Database.DSN != "" {
		db, err = connectWithRetry(cfg.Database.DSN, 30, 2*time.Second)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Pool.Close()

		// Run embedded migrations (idempotent, safe on every boot)
		migCtx, cancelMig := context.WithTimeout(context.Background(), 60*time.Second)
		if err := migrations.Apply(migCtx, db); err != nil {
			cancelMig()
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		cancelMig()
	} else {
		log.Println("Skipping DB connection (no DSN provided)")
	}

	router := api.NewRouter(db, cfg.TelegramBotToken)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router.InitRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// connectWithRetry ждёт, пока БД станет доступна (полезно в docker-compose,
// где backend стартует до того, как Postgres готов принимать соединения).
func connectWithRetry(dsn string, attempts int, delay time.Duration) (*database.DB, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		db, err := database.ConnectDB(ctx, dsn)
		cancel()
		if err == nil {
			return db, nil
		}
		lastErr = err
		log.Printf("DB connect attempt %d/%d failed: %v", i+1, attempts, err)
		time.Sleep(delay)
	}
	return nil, lastErr
}
