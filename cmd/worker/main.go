// Workflow Engine worker: polls the database and executes pending step_runs.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SheykoWk/workflow-engine/internal/app/executor"
	"github.com/SheykoWk/workflow-engine/internal/infrastructure/db"
	"github.com/joho/godotenv"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("load .env: %v", err)
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	sqlDB, err := db.OpenSQL(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	stepRunRepo := db.NewStepRunRepository(sqlDB)
	executor.Start(ctx, stepRunRepo)
	log.Printf("workflow worker started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("shutting down worker...")
	stop()
	log.Printf("worker stopped")
}
