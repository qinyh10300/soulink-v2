package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"soulink-backend/internal/config"
	"soulink-backend/internal/migrate"
	"soulink-backend/internal/platform/database"
	"soulink-backend/internal/questionnaire"
	"soulink-backend/internal/store"
)

func main() {
	catalogPath := flag.String("catalog", "", "path to the questionnaire catalog markdown file")
	outputPath := flag.String("output", "", "output CSV file path (default: stdout)")
	completedOnly := flag.Bool("completed-only", false, "only export users who completed the questionnaire")
	flag.Parse()

	if *catalogPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: questionnaireexport -catalog <path> [-output <path>] [-completed-only]")
		flag.PrintDefaults()
		os.Exit(2)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	rawCatalog, err := os.ReadFile(*catalogPath)
	if err != nil {
		logger.Error("read catalog file", "error", err, "path", *catalogPath)
		os.Exit(1)
	}
	catalog, err := questionnaire.ParseCatalog(string(rawCatalog))
	if err != nil {
		logger.Error("parse catalog", "error", err)
		os.Exit(1)
	}
	logger.Info("catalog loaded",
		"key", catalog.Key,
		"version", catalog.Version,
		"questions", len(catalog.Questions),
	)

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

	states, err := dataStore.ListQuestionnaireStatesByVersion(ctx, nil, catalog.Key, catalog.Version)
	if err != nil {
		logger.Error("list questionnaire states", "error", err)
		os.Exit(1)
	}

	if *completedOnly {
		filtered := states[:0]
		for _, s := range states {
			if s.HasCompletedQuestionnaire {
				filtered = append(filtered, s)
			}
		}
		states = filtered
	}

	if len(states) == 0 {
		logger.Info("no users found",
			"key", catalog.Key,
			"version", catalog.Version,
			"completedOnly", *completedOnly,
		)
		return
	}
	logger.Info("users found", "count", len(states))

	header := buildHeader(catalog)

	var out *os.File
	actualPath := *outputPath
	if actualPath != "" {
		actualPath = versionedPath(actualPath, catalog.Key, catalog.Version)
		f, err := os.Create(actualPath)
		if err != nil {
			logger.Error("create output file", "error", err, "path", actualPath)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	} else {
		out = os.Stdout
	}

	// UTF-8 BOM for Excel compatibility with Chinese characters
	out.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(out)
	defer w.Flush()

	if err := w.Write(header); err != nil {
		logger.Error("write csv header", "error", err)
		os.Exit(1)
	}

	exported := 0
	for _, state := range states {
		user, err := dataStore.GetUserByID(ctx, nil, state.UserID)
		if err != nil {
			logger.Warn("skip user (deleted or not found)", "user_id", state.UserID, "error", err)
			continue
		}

		answers, err := dataStore.ListQuestionnaireAnswers(ctx, nil, state.UserID, catalog.Key, catalog.Version)
		if err != nil {
			logger.Warn("list answers failed", "user_id", state.UserID, "error", err)
			continue
		}

		row := buildRow(catalog, user, state, answers)
		if err := w.Write(row); err != nil {
			logger.Error("write csv row", "user_id", user.ID, "error", err)
			os.Exit(1)
		}
		exported++
	}

	w.Flush()
	if err := w.Error(); err != nil {
		logger.Error("flush csv", "error", err)
		os.Exit(1)
	}

	logger.Info("export complete",
		"key", catalog.Key,
		"version", catalog.Version,
		"exported", exported,
		"questions", len(catalog.Questions),
	)
	if actualPath != "" {
		logger.Info("output written", "path", actualPath)
	}
}

// versionedPath inserts _{key}_{version} before the file extension.
// e.g. "export.csv" with key="thu-date", version="v0.3" → "export_thu-date_v0.3.csv"
func versionedPath(path, key, version string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	return base + "_" + key + "_" + version + ext
}

func buildHeader(catalog *questionnaire.Catalog) []string {
	header := []string{
		"user_id", "email", "nickname", "department", "grade",
		"completed", "participating", "progress_percent", "answered_count",
	}
	for _, q := range catalog.Questions {
		header = append(header, q.ID+"_value")
		if q.SupportsImportance {
			header = append(header, q.ID+"_importance")
		}
	}
	return header
}

func buildRow(catalog *questionnaire.Catalog, user store.User, state store.QuestionnaireState, answers []store.QuestionnaireAnswer) []string {
	answerMap := make(map[string]store.QuestionnaireAnswer, len(answers))
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}

	row := []string{
		strconv.FormatInt(user.ID, 10),
		user.Email,
		user.Nickname,
		user.Department,
		user.Grade,
		strconv.FormatBool(state.HasCompletedQuestionnaire),
		strconv.FormatBool(state.ParticipatingInMatch),
		strconv.Itoa(state.ProgressPercent),
		strconv.Itoa(state.AnsweredCount),
	}

	for _, q := range catalog.Questions {
		a, ok := answerMap[q.ID]
		if !ok {
			row = append(row, "")
			if q.SupportsImportance {
				row = append(row, "")
			}
			continue
		}
		row = append(row, flattenValue(q.Type, a.ValueJSON))
		if q.SupportsImportance {
			row = append(row, strconv.FormatBool(a.IsImportant))
		}
	}

	return row
}

func flattenValue(qType questionnaire.QuestionType, raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	switch qType {
	case questionnaire.QuestionTypeSingleChoice, questionnaire.QuestionTypeText:
		var s string
		if json.Unmarshal(raw, &s) == nil {
			return s
		}

	case questionnaire.QuestionTypeMultipleChoice:
		var arr []string
		if json.Unmarshal(raw, &arr) == nil {
			return strings.Join(arr, ";")
		}

	case questionnaire.QuestionTypeNumber, questionnaire.QuestionTypeScale, questionnaire.QuestionTypeRange:
		var n float64
		if json.Unmarshal(raw, &n) == nil {
			if n == float64(int64(n)) {
				return strconv.FormatInt(int64(n), 10)
			}
			return strconv.FormatFloat(n, 'f', -1, 64)
		}
	}

	return string(raw)
}
