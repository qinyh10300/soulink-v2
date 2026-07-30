package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"soulink-backend/internal/app"
	"soulink-backend/internal/auth"
	"soulink-backend/internal/chat"
	"soulink-backend/internal/config"
	"soulink-backend/internal/feedback"
	"soulink-backend/internal/friend"
	"soulink-backend/internal/gate"
	"soulink-backend/internal/mail"
	"soulink-backend/internal/match"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/migrate"
	platformcache "soulink-backend/internal/platform/cache"
	"soulink-backend/internal/platform/database"
	"soulink-backend/internal/playground"
	"soulink-backend/internal/profile"
	"soulink-backend/internal/questionnaire"
	"soulink-backend/internal/store"
)

func main() {
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

	dataStore := store.New(db)

	var redisCache *platformcache.Redis
	var verificationCache auth.VerificationCodeCache
	var resultCache match.ResultCache
	var registeredUserCountCache meta.RegisteredUserCountCache
	if cfg.Redis.Enabled {
		redisCache, err = platformcache.Open(ctx, cfg.Redis)
		if err != nil {
			logger.Error("open redis", "error", err)
			os.Exit(1)
		}
		defer redisCache.Close()
		verificationCache = auth.NewRedisVerificationCodeCache(redisCache)
		resultCache = match.NewRedisResultCache(redisCache)
		registeredUserCountCache = meta.NewRedisRegisteredUserCountCache(redisCache)
		logger.Info("redis cache connected", "addr", cfg.Redis.Addr, "db", cfg.Redis.DB)
	}

	mailer, err := mail.NewSender(cfg.Mail, logger)
	if err != nil {
		logger.Error("create mail sender", "error", err)
		os.Exit(1)
	}
	questionnaireCatalog, err := questionnaire.LoadCatalog(cfg.Questionnaire.CatalogFile)
	if err != nil {
		logger.Error("load questionnaire catalog", "error", err)
		os.Exit(1)
	}
	gateService := gate.NewService(dataStore)
	tokenManager := auth.NewTokenManager(cfg.Auth.JWTSecret, cfg.Auth.JWTTTL)
	metaService := meta.NewService(
		meta.WithUserCountStore(dataStore),
		meta.WithRegisteredUserCountCache(registeredUserCountCache, 5*time.Minute),
	)

	authService := auth.NewServiceWithVerificationCache(
		dataStore,
		mailer,
		tokenManager,
		logger,
		cfg.Auth,
		verificationCache,
	).WithRegisteredUserCountRefresher(metaService)

	feedbackService := feedback.NewService(dataStore, logger)
	profileService := profile.NewService(dataStore)
	playgroundService := playground.NewService(dataStore)
	questionnaireService := questionnaire.NewService(
		dataStore,
		questionnaireCatalog,
		questionnaire.WithSoulAgentEnabled(cfg.Questionnaire.SoulAgentEnabled),
	)
	matchService := match.NewService(
		dataStore,
		match.WithResultCache(resultCache, cfg.Match.ResultCacheTTL),
	)
	chatHub := chat.NewHub(logger)
	chatService := chat.NewService(dataStore)

	friendNotify := func(userID int64, eventType string, data any) {
		raw, err := json.Marshal(data)
		if err != nil {
			logger.Warn("failed to marshal friend notification", "error", err)
			return
		}
		chatHub.SendToUser(userID, chat.WSEnvelope{Type: eventType, Data: raw})
	}
	friendService := friend.NewService(dataStore, friendNotify)

	router := app.NewRouter(app.Dependencies{
		Logger:               logger,
		AllowedOrigins:       cfg.AllowedOrigins,
		AuthMiddleware:       auth.Middleware(dataStore, tokenManager),
		MatchWriteGuard:      app.MatchWriteGuard(logger, gateService),
		AuthHandler:          auth.NewHandler(authService),
		FeedbackHandler:      feedback.NewHandler(feedbackService),
		ProfileHandler:       profile.NewHandler(profileService),
		MetaHandler:          meta.NewHandler(metaService),
		QuestionnaireHandler: questionnaire.NewHandler(questionnaireService),
		MatchHandler:         match.NewHandler(matchService),
		FriendHandler:        friend.NewHandler(friendService),
		ChatHandler:          chat.NewHandler(chatService, chatHub),
		PlaygroundHandler:    playground.NewHandler(playgroundService),
		WSHandler:            chat.ServeWS(chatHub, chatService, dataStore, tokenManager, logger),
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       0,
		WriteTimeout:      0,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api server listening", "addr", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	case <-stopCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		logger.Info("shutting down server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown server", "error", err)
			os.Exit(1)
		}
	}
}
