package main

import (
	"context"
	"log/slog"
	"os"

	"soulink-backend/internal/config"
	"soulink-backend/internal/migrate"
	"soulink-backend/internal/platform/database"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(context.Background(), cfg.Database)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := migrate.Up(context.Background(), db); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations applied successfully")
}
