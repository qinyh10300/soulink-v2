package questionnaire

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"soulink-backend/internal/store"
)

type AnswerMigrationMapping struct {
	SourceQuestionID string `json:"sourceQuestionId"`
	TargetQuestionID string `json:"targetQuestionId"`
}

type AnswerMigrationConfig struct {
	SourceCatalogFile          string                   `json:"sourceCatalogFile"`
	SourceQuestionnaireKey     string                   `json:"sourceQuestionnaireKey,omitempty"`
	SourceQuestionnaireVersion string                   `json:"sourceQuestionnaireVersion,omitempty"`
	TargetCatalogFile          string                   `json:"targetCatalogFile"`
	TargetQuestionnaireKey     string                   `json:"targetQuestionnaireKey,omitempty"`
	TargetQuestionnaireVersion string                   `json:"targetQuestionnaireVersion,omitempty"`
	Mappings                   []AnswerMigrationMapping `json:"mappings"`
}

type AnswerMigrationPlan struct {
	ConfigPath    string
	SourceCatalog *Catalog
	TargetCatalog *Catalog
	Mappings      []AnswerMigrationMapping
}

type AnswerMigrationStats struct {
	UserCount           int
	MigratedAnswerCount int
	SkippedAnswerCount  int
	CompletedUserCount  int
	IncompleteUserCount int
}

type preparedMigratedAnswer struct {
	QuestionID string
	ValueJSON  json.RawMessage
	Importance bool
	AnsweredAt time.Time
}

type userMigrationStats struct {
	MigratedAnswerCount int
	SkippedAnswerCount  int
	Completed           bool
}

type AnswerMigrationService struct {
	store   *store.Store
	nowFunc func() time.Time
}

func NewAnswerMigrationService(dataStore *store.Store) *AnswerMigrationService {
	return &AnswerMigrationService{store: dataStore, nowFunc: time.Now}
}

func LoadAnswerMigrationPlan(path string) (*AnswerMigrationPlan, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("migration config path is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read migration config %q: %w", path, err)
	}

	var cfg AnswerMigrationConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("decode migration config %q: %w", path, err)
	}

	plan, err := resolveAnswerMigrationPlan(cfg, LoadCatalog)
	if err != nil {
		return nil, err
	}
	plan.ConfigPath = path
	return plan, nil
}

func resolveAnswerMigrationPlan(cfg AnswerMigrationConfig, catalogLoader func(string) (*Catalog, error)) (*AnswerMigrationPlan, error) {
	cfg.SourceCatalogFile = strings.TrimSpace(cfg.SourceCatalogFile)
	cfg.TargetCatalogFile = strings.TrimSpace(cfg.TargetCatalogFile)
	if cfg.SourceCatalogFile == "" {
		return nil, fmt.Errorf("sourceCatalogFile is required")
	}
	if cfg.TargetCatalogFile == "" {
		return nil, fmt.Errorf("targetCatalogFile is required")
	}
	if len(cfg.Mappings) == 0 {
		return nil, fmt.Errorf("mappings is required")
	}

	sourceCatalog, err := catalogLoader(cfg.SourceCatalogFile)
	if err != nil {
		return nil, fmt.Errorf("load source catalog: %w", err)
	}
	targetCatalog, err := catalogLoader(cfg.TargetCatalogFile)
	if err != nil {
		return nil, fmt.Errorf("load target catalog: %w", err)
	}

	if sourceCatalog.Key == targetCatalog.Key && sourceCatalog.Version == targetCatalog.Version {
		return nil, fmt.Errorf("source and target questionnaire versions must be different")
	}
	if expected := strings.TrimSpace(cfg.SourceQuestionnaireKey); expected != "" && expected != sourceCatalog.Key {
		return nil, fmt.Errorf("sourceQuestionnaireKey mismatch: expected %s, got %s", expected, sourceCatalog.Key)
	}
	if expected := strings.TrimSpace(cfg.SourceQuestionnaireVersion); expected != "" && expected != sourceCatalog.Version {
		return nil, fmt.Errorf("sourceQuestionnaireVersion mismatch: expected %s, got %s", expected, sourceCatalog.Version)
	}
	if expected := strings.TrimSpace(cfg.TargetQuestionnaireKey); expected != "" && expected != targetCatalog.Key {
		return nil, fmt.Errorf("targetQuestionnaireKey mismatch: expected %s, got %s", expected, targetCatalog.Key)
	}
	if expected := strings.TrimSpace(cfg.TargetQuestionnaireVersion); expected != "" && expected != targetCatalog.Version {
		return nil, fmt.Errorf("targetQuestionnaireVersion mismatch: expected %s, got %s", expected, targetCatalog.Version)
	}

	seenSource := make(map[string]struct{}, len(cfg.Mappings))
	seenTarget := make(map[string]struct{}, len(cfg.Mappings))
	mappings := make([]AnswerMigrationMapping, 0, len(cfg.Mappings))
	for _, mapping := range cfg.Mappings {
		sourceQuestionID := strings.TrimSpace(mapping.SourceQuestionID)
		targetQuestionID := strings.TrimSpace(mapping.TargetQuestionID)
		if sourceQuestionID == "" || targetQuestionID == "" {
			return nil, fmt.Errorf("question mappings must include sourceQuestionId and targetQuestionId")
		}
		if _, ok := sourceCatalog.QuestionByID(sourceQuestionID); !ok {
			return nil, fmt.Errorf("source question %s not found in %s", sourceQuestionID, cfg.SourceCatalogFile)
		}
		if _, ok := targetCatalog.QuestionByID(targetQuestionID); !ok {
			return nil, fmt.Errorf("target question %s not found in %s", targetQuestionID, cfg.TargetCatalogFile)
		}
		if _, exists := seenSource[sourceQuestionID]; exists {
			return nil, fmt.Errorf("duplicate source question mapping for %s", sourceQuestionID)
		}
		if _, exists := seenTarget[targetQuestionID]; exists {
			return nil, fmt.Errorf("duplicate target question mapping for %s", targetQuestionID)
		}
		seenSource[sourceQuestionID] = struct{}{}
		seenTarget[targetQuestionID] = struct{}{}
		mappings = append(mappings, AnswerMigrationMapping{SourceQuestionID: sourceQuestionID, TargetQuestionID: targetQuestionID})
	}

	return &AnswerMigrationPlan{
		SourceCatalog: sourceCatalog,
		TargetCatalog: targetCatalog,
		Mappings:      mappings,
	}, nil
}

func (s *AnswerMigrationService) Migrate(ctx context.Context, plan *AnswerMigrationPlan) (AnswerMigrationStats, error) {
	if s == nil || s.store == nil {
		return AnswerMigrationStats{}, fmt.Errorf("answer migration service is not ready")
	}
	if plan == nil || plan.SourceCatalog == nil || plan.TargetCatalog == nil {
		return AnswerMigrationStats{}, fmt.Errorf("answer migration plan is not ready")
	}

	sourceStates, err := s.store.ListQuestionnaireStatesByVersion(ctx, nil, plan.SourceCatalog.Key, plan.SourceCatalog.Version)
	if err != nil {
		return AnswerMigrationStats{}, fmt.Errorf("list source questionnaire states: %w", err)
	}

	stats := AnswerMigrationStats{UserCount: len(sourceStates)}
	now := s.now()
	for _, sourceState := range sourceStates {
		userStats, err := s.migrateUser(ctx, plan, sourceState, now)
		if err != nil {
			return AnswerMigrationStats{}, fmt.Errorf("migrate questionnaire answers for user %d: %w", sourceState.UserID, err)
		}
		stats.MigratedAnswerCount += userStats.MigratedAnswerCount
		stats.SkippedAnswerCount += userStats.SkippedAnswerCount
		if userStats.Completed {
			stats.CompletedUserCount++
		} else {
			stats.IncompleteUserCount++
		}
	}

	return stats, nil
}

func (s *AnswerMigrationService) migrateUser(ctx context.Context, plan *AnswerMigrationPlan, sourceState store.QuestionnaireState, now time.Time) (userMigrationStats, error) {
	stats := userMigrationStats{}
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		sourceAnswers, err := s.store.ListQuestionnaireAnswers(ctx, tx, sourceState.UserID, plan.SourceCatalog.Key, plan.SourceCatalog.Version)
		if err != nil {
			return fmt.Errorf("list source questionnaire answers: %w", err)
		}
		preparedAnswers, skippedCount, err := prepareMigratedAnswers(sourceAnswers, plan.TargetCatalog, plan.Mappings)
		if err != nil {
			return err
		}
		stats.SkippedAnswerCount = skippedCount

		if err := s.store.DeleteQuestionnaireVersionDataByUser(ctx, tx, sourceState.UserID, plan.TargetCatalog.Key, plan.TargetCatalog.Version); err != nil {
			return err
		}

		for _, answer := range preparedAnswers {
			if _, err := s.store.UpsertQuestionnaireAnswer(ctx, tx, store.UpsertQuestionnaireAnswerParams{
				UserID:               sourceState.UserID,
				QuestionnaireKey:     plan.TargetCatalog.Key,
				QuestionnaireVersion: plan.TargetCatalog.Version,
				QuestionID:           answer.QuestionID,
				ValueJSON:            answer.ValueJSON,
				IsImportant:          answer.Importance,
				AnsweredAt:           answer.AnsweredAt,
				CreatedAt:            now,
				UpdatedAt:            now,
			}); err != nil {
				return fmt.Errorf("upsert migrated questionnaire answer: %w", err)
			}
		}
		stats.MigratedAnswerCount = len(preparedAnswers)

		targetAnswers, err := s.store.ListQuestionnaireAnswers(ctx, tx, sourceState.UserID, plan.TargetCatalog.Key, plan.TargetCatalog.Version)
		if err != nil {
			return fmt.Errorf("list target questionnaire answers: %w", err)
		}
		answeredCount, requiredAnsweredCount, _, err := buildAnswerSnapshotForCatalog(plan.TargetCatalog, targetAnswers)
		if err != nil {
			return err
		}

		hasCompleted := requiredAnsweredCount == plan.TargetCatalog.RequiredCount
		participatingInMatch := sourceState.ParticipatingInMatch && hasCompleted
		stats.Completed = hasCompleted

		if _, err := s.store.UpsertQuestionnaireState(ctx, tx, store.UpsertQuestionnaireStateParams{
			UserID:                    sourceState.UserID,
			QuestionnaireKey:          plan.TargetCatalog.Key,
			QuestionnaireTitle:        plan.TargetCatalog.Title,
			QuestionnaireVersion:      plan.TargetCatalog.Version,
			TotalCount:                plan.TargetCatalog.TotalCount,
			RequiredCount:             plan.TargetCatalog.RequiredCount,
			AnsweredCount:             answeredCount,
			RequiredAnsweredCount:     requiredAnsweredCount,
			ProgressPercent:           progressPercent(answeredCount, plan.TargetCatalog.TotalCount),
			HasCompletedQuestionnaire: hasCompleted,
			ParticipatingInMatch:      participatingInMatch,
			LastAction:                "migrate",
			LastSavedAt:               now,
			ResultRevealAt:            nil,
			CreatedAt:                 now,
			UpdatedAt:                 now,
		}); err != nil {
			return fmt.Errorf("upsert target questionnaire state: %w", err)
		}

		if err := s.store.SetUserQuestionnaireCompleted(ctx, tx, sourceState.UserID, hasCompleted, now); err != nil {
			return fmt.Errorf("update user questionnaire completion: %w", err)
		}
		return nil
	})
	if err != nil {
		return userMigrationStats{}, err
	}
	return stats, nil
}

func prepareMigratedAnswers(sourceAnswers []store.QuestionnaireAnswer, targetCatalog *Catalog, mappings []AnswerMigrationMapping) ([]preparedMigratedAnswer, int, error) {
	sourceByQuestionID := make(map[string]store.QuestionnaireAnswer, len(sourceAnswers))
	for _, answer := range sourceAnswers {
		sourceByQuestionID[answer.QuestionID] = answer
	}

	prepared := make([]preparedMigratedAnswer, 0, len(mappings))
	skippedCount := 0
	for _, mapping := range mappings {
		sourceAnswer, ok := sourceByQuestionID[mapping.SourceQuestionID]
		if !ok {
			continue
		}
		targetQuestion, ok := targetCatalog.QuestionByID(mapping.TargetQuestionID)
		if !ok {
			return nil, 0, fmt.Errorf("target question %s not found", mapping.TargetQuestionID)
		}

		var importancePtr *bool
		if targetQuestion.SupportsImportance {
			importance := sourceAnswer.IsImportant
			importancePtr = &importance
		}
		normalizedJSON, _, normalizedImportance, err := validateAndNormalizeAnswer(targetQuestion, sourceAnswer.ValueJSON, importancePtr)
		if err != nil {
			skippedCount++
			continue
		}

		prepared = append(prepared, preparedMigratedAnswer{
			QuestionID: mapping.TargetQuestionID,
			ValueJSON:  normalizedJSON,
			Importance: normalizedImportance,
			AnsweredAt: sourceAnswer.AnsweredAt,
		})
	}
	return prepared, skippedCount, nil
}

func (s *AnswerMigrationService) now() time.Time {
	if s != nil && s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}
