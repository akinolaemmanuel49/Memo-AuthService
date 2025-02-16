package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/akinolaemmanuel49/Memo-Microservices/AuthService/api/server"
	"github.com/akinolaemmanuel49/Memo-Microservices/AuthService/config"
	"github.com/akinolaemmanuel49/Memo-Microservices/AuthService/internal/repository/database"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(ctx, cfg)
	if err != nil {
		cfg.Logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize server
	srv := server.NewServer(cfg, cfg.Logger, db)

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		cfg.Logger.Info("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			cfg.Logger.Error("failed to shutdown server", "error", err)
		}
	}()

	// Start server
	if err := srv.Start(); err != nil {
		cfg.Logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
