package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"soulink-backend/internal/ai"
	"soulink-backend/internal/analytics"
	"soulink-backend/internal/app"
	"soulink-backend/internal/auth"
	"soulink-backend/internal/calendar"
	"soulink-backend/internal/chat"
	"soulink-backend/internal/config"
	"soulink-backend/internal/connection"
	"soulink-backend/internal/demoseed"
	"soulink-backend/internal/engagement"
	"soulink-backend/internal/feedback"
	"soulink-backend/internal/friend"
	"soulink-backend/internal/gate"
	"soulink-backend/internal/mail"
	"soulink-backend/internal/match"
	"soulink-backend/internal/meetup"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/migrate"
	"soulink-backend/internal/operations"
	platformcache "soulink-backend/internal/platform/cache"
	"soulink-backend/internal/platform/database"
	"soulink-backend/internal/playground"
	"soulink-backend/internal/profile"
	"soulink-backend/internal/questionnaire"
	"soulink-backend/internal/safety"
	"soulink-backend/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	operationalMetrics := operations.NewCollector(logger)
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}
	ctx := context.Background()
	databaseStartedAt := time.Now()
	db, err := database.Open(ctx, cfg.Database)
	operationalMetrics.ObserveTask(ctx, "database.open", time.Since(databaseStartedAt), err)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if cfg.Database.AutoMigrate {
		migrationStartedAt := time.Now()
		err := migrate.Up(ctx, db)
		operationalMetrics.ObserveTask(ctx, "database.migrate", time.Since(migrationStartedAt), err)
		if err != nil {
			logger.Error("run migrations", "error", err)
			os.Exit(1)
		}
	}

	dataStore := store.New(db)
	if cfg.Demo.Enabled {
		if strings.EqualFold(cfg.AppEnv, "production") {
			logger.Error("demo seed cannot be enabled in production")
			os.Exit(1)
		}
		seedStartedAt := time.Now()
		seedResult, err := demoseed.Ensure(ctx, dataStore, demoseed.Config{
			Password:      cfg.Demo.Password,
			AdminEmail:    cfg.Demo.AdminEmail,
			AdminPassword: cfg.Demo.AdminPassword,
			BCryptCost:    cfg.Auth.BCryptCost,
		})
		operationalMetrics.ObserveTask(ctx, "demo.seed", time.Since(seedStartedAt), err)
		if err != nil {
			logger.Error("seed local demo workspace", "error", err)
			os.Exit(1)
		}
		logger.Info(
			"local demo workspace ready",
			"email", demoseed.DemoEmail,
			"admin_email", cfg.Demo.AdminEmail,
			"users_created", seedResult.UsersCreated,
			"intents_created", seedResult.IntentsCreated,
			"group_intents_created", seedResult.GroupIntentsCreated,
			"rooms_created", seedResult.RoomsCreated,
			"posts_created", seedResult.PostsCreated,
			"groups_created", seedResult.GroupsCreated,
			"events_created", seedResult.EventsCreated,
		)
	}

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
	var modelGenerator ai.Generator
	if cfg.AI.Provider == "openai-compatible" {
		provider := ai.NewOpenAICompatible(ai.OpenAICompatibleConfig{
			BaseURL:    cfg.AI.BaseURL,
			APIKey:     cfg.AI.APIKey,
			Model:      cfg.AI.Model,
			Timeout:    cfg.AI.Timeout,
			MaxRetries: cfg.AI.MaxRetries,
		})
		if provider.Enabled() {
			modelGenerator = operations.InstrumentGenerator(provider, operationalMetrics)
			logger.Info("AI provider enabled", "provider", cfg.AI.Provider, "model", cfg.AI.Model)
		} else {
			logger.Info("AI provider incomplete; deterministic fallback enabled")
		}
	} else if cfg.AI.Provider != "disabled" && cfg.AI.Provider != "" {
		logger.Warn("unsupported AI provider; deterministic fallback enabled", "provider", cfg.AI.Provider)
	}
	connectionService := connection.NewService(
		dataStore,
		connection.NewInterpreterWithGenerator(modelGenerator),
	)
	adminEmails := append([]string(nil), cfg.Safety.AdminEmails...)
	if cfg.Demo.Enabled && cfg.Demo.AdminEmail != "" {
		adminEmails = append(adminEmails, cfg.Demo.AdminEmail)
	}
	safetyService := safety.NewService(dataStore, adminEmails)
	analyticsService := analytics.NewService(dataStore, adminEmails)
	engagementService := engagement.NewService(dataStore, adminEmails).WithTaskObserver(operationalMetrics)
	operationsHandler := operations.NewHandler(operationalMetrics, adminEmails)
	connectionService.WithEventRecorder(analyticsService)
	playgroundService.WithEventRecorder(analyticsService)
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
	chatService := chat.NewService(dataStore).WithGenerator(modelGenerator)
	calendarService := calendar.NewService(dataStore)
	meetupService := meetup.NewService(dataStore, analyticsService).WithMailer(mailer)

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
		RequestObserver:      operationalMetrics,
		AllowedOrigins:       cfg.AllowedOrigins,
		AuthMiddleware:       auth.Middleware(dataStore, tokenManager),
		MatchWriteGuard:      app.MatchWriteGuard(logger, gateService),
		EventRecorder:        analyticsService,
		AnalyticsHandler:     analytics.NewHandler(analyticsService),
		OperationsHandler:    operationsHandler.Metrics,
		AuthHandler:          auth.NewHandler(authService),
		CalendarHandler:      calendar.NewHandler(calendarService),
		FeedbackHandler:      feedback.NewHandler(feedbackService),
		ProfileHandler:       profile.NewHandler(profileService),
		MetaHandler:          meta.NewHandler(metaService),
		QuestionnaireHandler: questionnaire.NewHandler(questionnaireService),
		MatchHandler:         match.NewHandler(matchService),
		MeetupHandler:        meetup.NewHandler(meetupService),
		FriendHandler:        friend.NewHandler(friendService),
		ChatHandler:          chat.NewHandler(chatService, chatHub),
		PlaygroundHandler:    playground.NewHandler(playgroundService),
		ConnectionHandler:    connection.NewHandler(connectionService),
		EngagementHandler:    engagement.NewHandler(engagementService),
		SafetyHandler:        safety.NewHandler(safetyService),
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
