package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"soulink-backend/internal/config"
	"soulink-backend/internal/gate"
	"soulink-backend/internal/migrate"
	"soulink-backend/internal/platform/database"
	"soulink-backend/internal/store"
)

func main() {
	action := flag.String("action", "status", "set | clear | status")
	gateKey := flag.String("gate", gate.GateKeyMatchingWrite, "gate key")
	startRaw := flag.String("start", "", "start time, e.g. 2026-03-26 10:00 or 2026-03-26T10:00:00+08:00")
	endRaw := flag.String("end", "", "end time, e.g. 2026-03-26 12:00 or 2026-03-26T12:00:00+08:00")
	reason := flag.String("reason", "", "close reason returned to clients")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
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

	gateService := gate.NewService(store.New(db))
	switch strings.ToLower(strings.TrimSpace(*action)) {
	case "set":
		startsAt, err := parseCommandTime(*startRaw)
		if err != nil {
			logger.Error("parse start time", "error", err)
			os.Exit(1)
		}
		endsAt, err := parseCommandTime(*endRaw)
		if err != nil {
			logger.Error("parse end time", "error", err)
			os.Exit(1)
		}
		window, err := gateService.SetWindow(ctx, *gateKey, startsAt, endsAt, *reason)
		if err != nil {
			logger.Error("set gate window", "error", err)
			os.Exit(1)
		}
		fmt.Printf("gate=%s set start=%s end=%s reason=%s\n", window.GateKey, window.StartsAt.Format(time.RFC3339), window.EndsAt.Format(time.RFC3339), window.Reason)
	case "clear":
		if err := gateService.ClearWindow(ctx, *gateKey); err != nil {
			logger.Error("clear gate window", "error", err)
			os.Exit(1)
		}
		fmt.Printf("gate=%s cleared\n", strings.TrimSpace(*gateKey))
	case "status":
		status, err := gateService.Check(ctx, *gateKey)
		if err != nil {
			logger.Error("check gate window", "error", err)
			os.Exit(1)
		}
		if status.StartsAt.IsZero() && status.EndsAt.IsZero() {
			fmt.Printf("gate=%s not configured\n", strings.TrimSpace(*gateKey))
			return
		}
		state := "open_now"
		if status.Active {
			state = "closed_now"
		}
		fmt.Printf("gate=%s state=%s start=%s end=%s reason=%s\n", status.GateKey, state, status.StartsAt.Format(time.RFC3339), status.EndsAt.Format(time.RFC3339), status.Reason)
	default:
		fmt.Fprintln(os.Stderr, "unsupported action, use set | clear | status")
		flag.Usage()
		os.Exit(2)
	}
}

func parseCommandTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("time value is required")
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, fmt.Errorf("load timezone: %w", err)
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format %q", value)
}
