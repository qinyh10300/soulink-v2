package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

type AIUserPreferences struct {
	UserID                        int64
	RecommendationLearningEnabled bool
	AssistantEnabled              bool
	LearningResetAt               sql.NullTime
	AssistantMemoryResetAt        sql.NullTime
	CreatedAt                     time.Time
	UpdatedAt                     time.Time
}

type RecommendationInteractionSignal struct {
	ID          int64
	UserID      int64
	IntentID    int64
	OtherUserID int64
	Category    string
	SignalType  string
	Weight      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RevokedAt   sql.NullTime
}

type RecommendationAffinity struct {
	Category string
	Weight   int
	Count    int
}

type RecommendationRecentSignal struct {
	SignalType string
	Category   string
	Weight     int
	Source     string
	CreatedAt  time.Time
}

type RecommendationLearningSnapshot struct {
	CategoryWeights map[string]int
	UserWeights     map[int64]int
	Affinities      []RecommendationAffinity
	RecentSignals   []RecommendationRecentSignal
}

type UpsertRecommendationInteractionSignalParams struct {
	UserID      int64
	IntentID    int64
	OtherUserID int64
	Category    string
	SignalType  string
	Weight      int
	Now         time.Time
}

type DailyRecommendation struct {
	ID                    int64
	UserID                int64
	SourceIntentID        int64
	CandidateIntentID     int64
	Rank                  int
	Score                 int
	FitReason             string
	Uncertainty           string
	SharedContext         []string
	AlgorithmVersion      string
	InferenceSource       string
	ModelName             string
	Confidence            float64
	PromptVersion         string
	SourceIntentUpdatedAt time.Time
	CreatedAt             time.Time
	ExpiresAt             time.Time
}

type CreateDailyRecommendationParams struct {
	UserID                int64
	SourceIntentID        int64
	CandidateIntentID     int64
	Rank                  int
	Score                 int
	FitReason             string
	Uncertainty           string
	SharedContext         []string
	AlgorithmVersion      string
	InferenceSource       string
	ModelName             string
	Confidence            float64
	PromptVersion         string
	SourceIntentUpdatedAt time.Time
	Now                   time.Time
	ExpiresAt             time.Time
}

func (s *Store) GetAIUserPreferences(ctx context.Context, tx *sql.Tx, userID int64) (AIUserPreferences, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT user_id, recommendation_learning_enabled, assistant_enabled,
		       learning_reset_at, assistant_memory_reset_at, created_at, updated_at
		FROM ai_user_preferences
		WHERE user_id = $1
	`, userID)
	return scanAIUserPreferences(row)
}

func (s *Store) SetRecommendationLearningEnabled(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	enabled bool,
	now time.Time,
) (AIUserPreferences, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO ai_user_preferences (
			user_id, recommendation_learning_enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET recommendation_learning_enabled = EXCLUDED.recommendation_learning_enabled,
		    updated_at = EXCLUDED.updated_at
		RETURNING user_id, recommendation_learning_enabled, assistant_enabled,
		          learning_reset_at, assistant_memory_reset_at, created_at, updated_at
	`, userID, enabled, now)
	return scanAIUserPreferences(row)
}

func (s *Store) ResetRecommendationLearning(ctx context.Context, tx *sql.Tx, userID int64, now time.Time) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO ai_user_preferences (
			user_id, learning_reset_at, created_at, updated_at
		) VALUES ($1, $2, $2, $2)
		ON CONFLICT (user_id) DO UPDATE
		SET learning_reset_at = EXCLUDED.learning_reset_at,
		    updated_at = EXCLUDED.updated_at
	`, userID, now); err != nil {
		return fmt.Errorf("reset recommendation preference: %w", err)
	}
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE recommendation_interaction_signals
		SET revoked_at = $2, updated_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, now); err != nil {
		return fmt.Errorf("revoke recommendation interaction signals: %w", err)
	}
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE recommendation_learning_signals
		SET revoked_at = $2, updated_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, now); err != nil {
		return fmt.Errorf("revoke meetup recommendation signals: %w", err)
	}
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM daily_recommendations
		WHERE user_id = $1 AND expires_at > $2
	`, userID, now); err != nil {
		return fmt.Errorf("clear current daily recommendations: %w", err)
	}
	return nil
}

func (s *Store) UpsertRecommendationInteractionSignal(
	ctx context.Context,
	tx *sql.Tx,
	params UpsertRecommendationInteractionSignalParams,
) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO recommendation_interaction_signals (
			user_id, intent_id, other_user_id, category, signal_type,
			weight, created_at, updated_at, revoked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $7, NULL)
		ON CONFLICT (user_id, intent_id, signal_type) DO UPDATE
		SET other_user_id = EXCLUDED.other_user_id,
		    category = EXCLUDED.category,
		    weight = EXCLUDED.weight,
		    updated_at = EXCLUDED.updated_at,
		    revoked_at = NULL
	`, params.UserID, params.IntentID, params.OtherUserID, params.Category,
		params.SignalType, params.Weight, params.Now)
	if err != nil {
		return fmt.Errorf("upsert recommendation interaction signal: %w", err)
	}
	return nil
}

func (s *Store) RevokeRecommendationInteractionSignal(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	intentID int64,
	signalType string,
	now time.Time,
) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE recommendation_interaction_signals
		SET revoked_at = $4, updated_at = $4
		WHERE user_id = $1 AND intent_id = $2 AND signal_type = $3 AND revoked_at IS NULL
	`, userID, intentID, signalType, now); err != nil {
		return fmt.Errorf("revoke recommendation interaction signal: %w", err)
	}
	return nil
}

func (s *Store) GetRecommendationLearningSnapshot(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	after time.Time,
	recentLimit int,
) (RecommendationLearningSnapshot, error) {
	snapshot := RecommendationLearningSnapshot{
		CategoryWeights: make(map[string]int),
		UserWeights:     make(map[int64]int),
		Affinities:      []RecommendationAffinity{},
		RecentSignals:   []RecommendationRecentSignal{},
	}

	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT category, COALESCE(SUM(weight), 0)::int, COUNT(*)::int
		FROM recommendation_interaction_signals
		WHERE user_id = $1 AND revoked_at IS NULL AND created_at >= $2
		GROUP BY category
	`, userID, after)
	if err != nil {
		return snapshot, fmt.Errorf("list recommendation category affinity: %w", err)
	}
	for rows.Next() {
		var affinity RecommendationAffinity
		if err := rows.Scan(&affinity.Category, &affinity.Weight, &affinity.Count); err != nil {
			rows.Close()
			return snapshot, fmt.Errorf("scan recommendation category affinity: %w", err)
		}
		snapshot.CategoryWeights[affinity.Category] += affinity.Weight
		snapshot.Affinities = append(snapshot.Affinities, affinity)
	}
	if err := rows.Close(); err != nil {
		return snapshot, fmt.Errorf("close recommendation category affinity: %w", err)
	}

	meetupRows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT meetup.category, COALESCE(SUM(signal.weight), 0)::int, COUNT(*)::int
		FROM recommendation_learning_signals signal
		JOIN meetups meetup ON meetup.id = signal.source_meetup_id
		WHERE signal.user_id = $1 AND signal.revoked_at IS NULL AND signal.created_at >= $2
		GROUP BY meetup.category
	`, userID, after)
	if err != nil {
		return snapshot, fmt.Errorf("list meetup category affinity: %w", err)
	}
	for meetupRows.Next() {
		var affinity RecommendationAffinity
		if err := meetupRows.Scan(&affinity.Category, &affinity.Weight, &affinity.Count); err != nil {
			meetupRows.Close()
			return snapshot, fmt.Errorf("scan meetup category affinity: %w", err)
		}
		snapshot.CategoryWeights[affinity.Category] += affinity.Weight
		merged := false
		for index := range snapshot.Affinities {
			if snapshot.Affinities[index].Category == affinity.Category {
				snapshot.Affinities[index].Weight += affinity.Weight
				snapshot.Affinities[index].Count += affinity.Count
				merged = true
				break
			}
		}
		if !merged {
			snapshot.Affinities = append(snapshot.Affinities, affinity)
		}
	}
	if err := meetupRows.Close(); err != nil {
		return snapshot, fmt.Errorf("close meetup category affinity: %w", err)
	}

	userRows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT other_user_id, COALESCE(SUM(weight), 0)::int
		FROM (
			SELECT other_user_id, weight
			FROM recommendation_interaction_signals
			WHERE user_id = $1 AND revoked_at IS NULL AND created_at >= $2
			UNION ALL
			SELECT other_user_id, weight
			FROM recommendation_learning_signals
			WHERE user_id = $1 AND revoked_at IS NULL AND created_at >= $2
		) weighted
		GROUP BY other_user_id
	`, userID, after)
	if err != nil {
		return snapshot, fmt.Errorf("list recommendation user affinity: %w", err)
	}
	for userRows.Next() {
		var otherUserID int64
		var weight int
		if err := userRows.Scan(&otherUserID, &weight); err != nil {
			userRows.Close()
			return snapshot, fmt.Errorf("scan recommendation user affinity: %w", err)
		}
		snapshot.UserWeights[otherUserID] = weight
	}
	if err := userRows.Close(); err != nil {
		return snapshot, fmt.Errorf("close recommendation user affinity: %w", err)
	}

	if recentLimit <= 0 {
		recentLimit = 12
	}
	recentRows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT signal_type, category, weight, source, created_at
		FROM (
			SELECT signal_type, category, weight, 'interaction'::text AS source, updated_at AS created_at
			FROM recommendation_interaction_signals
			WHERE user_id = $1 AND revoked_at IS NULL AND created_at >= $2
			UNION ALL
			SELECT signal.signal_type, meetup.category, signal.weight, 'meetup'::text AS source,
			       signal.updated_at AS created_at
			FROM recommendation_learning_signals signal
			JOIN meetups meetup ON meetup.id = signal.source_meetup_id
			WHERE signal.user_id = $1 AND signal.revoked_at IS NULL AND signal.created_at >= $2
		) recent
		ORDER BY created_at DESC
		LIMIT $3
	`, userID, after, recentLimit)
	if err != nil {
		return snapshot, fmt.Errorf("list recent recommendation signals: %w", err)
	}
	defer recentRows.Close()
	for recentRows.Next() {
		var signal RecommendationRecentSignal
		if err := recentRows.Scan(
			&signal.SignalType,
			&signal.Category,
			&signal.Weight,
			&signal.Source,
			&signal.CreatedAt,
		); err != nil {
			return snapshot, fmt.Errorf("scan recent recommendation signal: %w", err)
		}
		snapshot.RecentSignals = append(snapshot.RecentSignals, signal)
	}
	if err := recentRows.Err(); err != nil {
		return snapshot, fmt.Errorf("iterate recent recommendation signals: %w", err)
	}
	sort.Slice(snapshot.Affinities, func(i, j int) bool {
		if snapshot.Affinities[i].Weight == snapshot.Affinities[j].Weight {
			return snapshot.Affinities[i].Category < snapshot.Affinities[j].Category
		}
		return snapshot.Affinities[i].Weight > snapshot.Affinities[j].Weight
	})
	return snapshot, nil
}

func (s *Store) ListCurrentDailyRecommendations(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	sourceIntentID int64,
	sourceIntentUpdatedAt time.Time,
	now time.Time,
) ([]DailyRecommendation, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		WITH latest_batch AS (
			SELECT MAX(created_at) AS created_at
			FROM daily_recommendations
			WHERE user_id = $1 AND source_intent_id = $2
			  AND source_intent_updated_at = $3 AND expires_at > $4
		)
		SELECT recommendation.id, recommendation.user_id, recommendation.source_intent_id,
		       recommendation.candidate_intent_id, recommendation.rank, recommendation.score,
		       recommendation.fit_reason, recommendation.uncertainty,
		       recommendation.shared_context::text, recommendation.algorithm_version,
		       recommendation.inference_source, recommendation.model_name,
		       recommendation.confidence::float8, recommendation.prompt_version,
		       recommendation.source_intent_updated_at, recommendation.created_at,
		       recommendation.expires_at
		FROM daily_recommendations recommendation
		JOIN latest_batch ON latest_batch.created_at = recommendation.created_at
		WHERE recommendation.user_id = $1 AND recommendation.source_intent_id = $2
		ORDER BY recommendation.rank
	`, userID, sourceIntentID, sourceIntentUpdatedAt, now)
	if err != nil {
		return nil, fmt.Errorf("list current daily recommendations: %w", err)
	}
	defer rows.Close()
	return scanDailyRecommendationRows(rows)
}

func (s *Store) LockDailyRecommendations(ctx context.Context, tx *sql.Tx, userID int64) error {
	if tx == nil {
		return fmt.Errorf("lock daily recommendations requires a transaction")
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return fmt.Errorf("lock daily recommendations: %w", err)
	}
	return nil
}

func (s *Store) ExpireActiveDailyRecommendations(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	sourceIntentID int64,
	now time.Time,
) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE daily_recommendations
		SET expires_at = $3
		WHERE user_id = $1
		  AND ($2 = 0 OR source_intent_id = $2)
		  AND expires_at > $3
		  AND created_at < $3
	`, userID, sourceIntentID, now); err != nil {
		return fmt.Errorf("expire active daily recommendations: %w", err)
	}
	return nil
}

func (s *Store) ListRecentDailyCandidateIDs(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	since time.Time,
) (map[int64]struct{}, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT DISTINCT candidate_intent_id
		FROM daily_recommendations
		WHERE user_id = $1 AND created_at >= $2
	`, userID, since)
	if err != nil {
		return nil, fmt.Errorf("list recent daily candidates: %w", err)
	}
	defer rows.Close()
	ids := make(map[int64]struct{})
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan recent daily candidate: %w", err)
		}
		ids[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent daily candidates: %w", err)
	}
	return ids, nil
}

func (s *Store) CreateDailyRecommendation(
	ctx context.Context,
	tx *sql.Tx,
	params CreateDailyRecommendationParams,
) error {
	sharedContext, err := marshalStringList(params.SharedContext)
	if err != nil {
		return err
	}
	_, err = executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO daily_recommendations (
			user_id, source_intent_id, candidate_intent_id, rank, score,
			fit_reason, uncertainty, shared_context, algorithm_version,
			inference_source, model_name, confidence, prompt_version,
			source_intent_updated_at, created_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9,
			$10, $11, $12, $13, $14, $15, $16
		)
	`, params.UserID, params.SourceIntentID, params.CandidateIntentID, params.Rank,
		params.Score, params.FitReason, params.Uncertainty, sharedContext,
		params.AlgorithmVersion, params.InferenceSource, params.ModelName,
		params.Confidence, params.PromptVersion, params.SourceIntentUpdatedAt,
		params.Now, params.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create daily recommendation: %w", err)
	}
	return nil
}

func scanAIUserPreferences(row rowScanner) (AIUserPreferences, error) {
	var preferences AIUserPreferences
	if err := row.Scan(
		&preferences.UserID,
		&preferences.RecommendationLearningEnabled,
		&preferences.AssistantEnabled,
		&preferences.LearningResetAt,
		&preferences.AssistantMemoryResetAt,
		&preferences.CreatedAt,
		&preferences.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return AIUserPreferences{}, ErrNotFound
		}
		return AIUserPreferences{}, fmt.Errorf("scan AI user preferences: %w", err)
	}
	return preferences, nil
}

func scanDailyRecommendationRows(rows *sql.Rows) ([]DailyRecommendation, error) {
	recommendations := make([]DailyRecommendation, 0, 3)
	for rows.Next() {
		var recommendation DailyRecommendation
		var rawSharedContext string
		if err := rows.Scan(
			&recommendation.ID,
			&recommendation.UserID,
			&recommendation.SourceIntentID,
			&recommendation.CandidateIntentID,
			&recommendation.Rank,
			&recommendation.Score,
			&recommendation.FitReason,
			&recommendation.Uncertainty,
			&rawSharedContext,
			&recommendation.AlgorithmVersion,
			&recommendation.InferenceSource,
			&recommendation.ModelName,
			&recommendation.Confidence,
			&recommendation.PromptVersion,
			&recommendation.SourceIntentUpdatedAt,
			&recommendation.CreatedAt,
			&recommendation.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("scan daily recommendation: %w", err)
		}
		if err := unmarshalStringList(rawSharedContext, &recommendation.SharedContext); err != nil {
			return nil, err
		}
		recommendations = append(recommendations, recommendation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate daily recommendations: %w", err)
	}
	return recommendations, nil
}
