package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"soulink-backend/internal/config"
	"soulink-backend/internal/matchimport"
	"soulink-backend/internal/migrate"
	"soulink-backend/internal/platform/database"
	"soulink-backend/internal/store"
)

type pairKey struct {
	A int64
	B int64
}

type stats struct {
	CSVRows                 int
	UniquePairs             int
	SkippedDuplicatePairs   int
	FriendshipsAccepted     int
	AlreadyAccepted         int
	SkippedBlocked          int
	ConversationsEnsured    int
	SkippedInvalidPairUsers int
	DryRun                  bool
}

func main() {
	csvPath := flag.String("csv", "", "path to the match_results.csv file")
	dryRun := flag.Bool("dry-run", false, "validate and count only, without writing data")
	ensureConversations := flag.Bool("ensure-conversations", false, "also ensure conversations for accepted friendships")
	flag.Parse()

	if *csvPath == "" {
		fmt.Fprintln(os.Stderr, "-csv is required")
		flag.Usage()
		os.Exit(2)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	pairs, err := matchimport.LoadPairs(*csvPath)
	if err != nil {
		logger.Error("load match pairs", "error", err)
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
	now := time.Now()

	seen := make(map[pairKey]struct{}, len(pairs))
	runStats := stats{
		CSVRows: len(pairs),
		DryRun:  *dryRun,
	}

	for _, pair := range pairs {
		key := normalizePairKey(pair.UserAID, pair.UserBID)
		if key.A <= 0 || key.B <= 0 {
			runStats.SkippedInvalidPairUsers++
			continue
		}
		if _, exists := seen[key]; exists {
			runStats.SkippedDuplicatePairs++
			continue
		}
		seen[key] = struct{}{}
		runStats.UniquePairs++

		if _, err := dataStore.GetUserByID(ctx, nil, key.A); err != nil {
			if err == store.ErrNotFound {
				logger.Warn("skip pair because user not found", "user_id", key.A, "pair_user_id", key.B)
				runStats.SkippedInvalidPairUsers++
				continue
			}
			logger.Error("get user by id", "user_id", key.A, "error", err)
			os.Exit(1)
		}
		if _, err := dataStore.GetUserByID(ctx, nil, key.B); err != nil {
			if err == store.ErrNotFound {
				logger.Warn("skip pair because user not found", "user_id", key.B, "pair_user_id", key.A)
				runStats.SkippedInvalidPairUsers++
				continue
			}
			logger.Error("get user by id", "user_id", key.B, "error", err)
			os.Exit(1)
		}

		friendship, err := dataStore.GetFriendshipBetween(ctx, nil, key.A, key.B)
		if err != nil && err != store.ErrNotFound {
			logger.Error("get friendship between users", "user_a_id", key.A, "user_b_id", key.B, "error", err)
			os.Exit(1)
		}

		if err == nil && friendship.Status == "blocked" {
			runStats.SkippedBlocked++
			continue
		}

		if err == nil && friendship.Status == "accepted" {
			runStats.AlreadyAccepted++
			if *ensureConversations && !*dryRun {
				if err := dataStore.WithTx(ctx, func(tx *sql.Tx) error {
					_, createErr := dataStore.CreateConversation(ctx, tx, key.A, key.B, now)
					return createErr
				}); err != nil {
					logger.Error("ensure conversation", "user_a_id", key.A, "user_b_id", key.B, "error", err)
					os.Exit(1)
				}
				runStats.ConversationsEnsured++
			}
			continue
		}

		if *dryRun {
			runStats.FriendshipsAccepted++
			if *ensureConversations {
				runStats.ConversationsEnsured++
			}
			continue
		}

		if err := dataStore.WithTx(ctx, func(tx *sql.Tx) error {
			if _, ensureErr := dataStore.EnsureAcceptedFriendshipBetween(ctx, tx, key.A, key.B, now); ensureErr != nil {
				return ensureErr
			}
			if *ensureConversations {
				if _, convErr := dataStore.CreateConversation(ctx, tx, key.A, key.B, now); convErr != nil {
					return convErr
				}
			}
			return nil
		}); err != nil {
			logger.Error("ensure accepted friendship", "user_a_id", key.A, "user_b_id", key.B, "error", err)
			os.Exit(1)
		}
		runStats.FriendshipsAccepted++
		if *ensureConversations {
			runStats.ConversationsEnsured++
		}
	}

	fmt.Printf(
		"csvRows=%d uniquePairs=%d accepted=%d alreadyAccepted=%d blockedSkipped=%d conversations=%d duplicatePairs=%d invalidPairUsers=%d dryRun=%t\n",
		runStats.CSVRows,
		runStats.UniquePairs,
		runStats.FriendshipsAccepted,
		runStats.AlreadyAccepted,
		runStats.SkippedBlocked,
		runStats.ConversationsEnsured,
		runStats.SkippedDuplicatePairs,
		runStats.SkippedInvalidPairUsers,
		runStats.DryRun,
	)
}

func normalizePairKey(userA, userB int64) pairKey {
	if userA <= 0 || userB <= 0 || userA == userB {
		return pairKey{}
	}
	if userA > userB {
		userA, userB = userB, userA
	}
	return pairKey{A: userA, B: userB}
}
