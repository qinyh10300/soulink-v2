package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"soulink-backend/internal/config"
	"soulink-backend/internal/match"
	"soulink-backend/internal/matchimport"
	"soulink-backend/internal/migrate"
	platformcache "soulink-backend/internal/platform/cache"
	"soulink-backend/internal/platform/database"
	"soulink-backend/internal/store"
)

func main() {
	configPath := flag.String("config", "", "path to the match result import config JSON")
	csvPath := flag.String("csv", "", "path to the match_results.csv file")
	explanationsPath := flag.String("explanations", "", "optional path to the match_explanations.csv file")
	flag.Parse()

	if *configPath == "" || *csvPath == "" {
		fmt.Fprintln(os.Stderr, "-config and -csv are required")
		flag.Usage()
		os.Exit(2)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	plan, err := matchimport.LoadPlan(*configPath)
	if err != nil {
		logger.Error("load match result import plan", "error", err)
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

	dataStore := store.New(db)

	var redisCache *platformcache.Redis
	var resultCache match.ResultCache
	if cfg.Redis.Enabled {
		redisCache, err = platformcache.Open(ctx, cfg.Redis)
		if err != nil {
			logger.Error("open redis", "error", err)
			os.Exit(1)
		}
		defer redisCache.Close()
		resultCache = match.NewRedisResultCache(redisCache)
	}

	matchService := match.NewService(
		dataStore,
		match.WithResultCache(resultCache, cfg.Match.ResultCacheTTL),
	)
	importService := matchimport.NewService(dataStore, matchService)

	stats, err := importService.Run(ctx, plan, *csvPath, *explanationsPath)
	if err != nil {
		logger.Error("import match results", "error", err)
		os.Exit(1)
	}

	fmt.Printf(
		"questionnaire=%s@%s matchedPairs=%d matchedUsers=%d unmatchedUsers=%d results=%d dryRun=%t cycleKey=%s cycleID=%d\n",
		stats.QuestionnaireKey,
		stats.QuestionnaireVersion,
		stats.MatchedPairCount,
		stats.MatchedUserCount,
		stats.UnmatchedUserCount,
		stats.ResultCount,
		stats.DryRun,
		stats.CycleKey,
		stats.CycleID,
	)
}
