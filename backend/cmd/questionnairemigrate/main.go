package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"soulink-backend/internal/config"
	"soulink-backend/internal/migrate"
	"soulink-backend/internal/platform/database"
	"soulink-backend/internal/questionnaire"
	"soulink-backend/internal/store"
)

func main() {
	configPath := flag.String("config", "", "path to the questionnaire answer migration config JSON")
	flag.Parse()

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "-config is required")
		flag.Usage()
		os.Exit(2)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	plan, err := questionnaire.LoadAnswerMigrationPlan(*configPath)
	if err != nil {
		logger.Error("load answer migration plan", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := database.Open(ctx, cfg.Database)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if cfg.Database.AutoMigrate {
		if err := migrate.Up(ctx, db); err != nil {
			logger.Error("run migrations", "error", err)
			os.Exit(1)
		}
	}

	migrationService := questionnaire.NewAnswerMigrationService(store.New(db))
	stats, err := migrationService.Migrate(ctx, plan)
	if err != nil {
		logger.Error("migrate questionnaire answers", "error", err)
		os.Exit(1)
	}

	fmt.Printf("source=%s@%s target=%s@%s users=%d migratedAnswers=%d skippedAnswers=%d completedUsers=%d incompleteUsers=%d\n",
		plan.SourceCatalog.Key,
		plan.SourceCatalog.Version,
		plan.TargetCatalog.Key,
		plan.TargetCatalog.Version,
		stats.UserCount,
		stats.MigratedAnswerCount,
		stats.SkippedAnswerCount,
		stats.CompletedUserCount,
		stats.IncompleteUserCount,
	)
}
