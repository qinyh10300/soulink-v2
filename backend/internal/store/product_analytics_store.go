package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type ProductEventParams struct {
	EventName   string
	ActorUserID sql.NullInt64
	SubjectType string
	SubjectID   sql.NullInt64
	RequestID   string
	DedupeKey   sql.NullString
	Properties  map[string]any
	OccurredAt  time.Time
}

type ProductEventMetric struct {
	EventName    string
	Count        int64
	UniqueActors int64
}

type DailyProductMetric struct {
	Date                  string
	ActiveUsers           int64
	IntentsCreated        int64
	RecommendationsViewed int64
	InterestsExpressed    int64
	MatchesCreated        int64
	MeetupsConfirmed      int64
	MeetupsCompleted      int64
	SafetyReports         int64
}

type ProductMetricsSnapshot struct {
	RegisteredUsers   int64
	ActiveUsers       int64
	IntentCreators    int64
	RepeatIntentUsers int64
	OpenSafetyReports int64
	EventMetrics      []ProductEventMetric
	DailyMetrics      []DailyProductMetric
}

func (s *Store) RecordProductEvent(
	ctx context.Context,
	tx *sql.Tx,
	params ProductEventParams,
) error {
	properties := params.Properties
	if properties == nil {
		properties = map[string]any{}
	}
	rawProperties, err := json.Marshal(properties)
	if err != nil {
		return fmt.Errorf("marshal product event properties: %w", err)
	}

	_, err = executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO product_events (
			event_name, actor_user_id, subject_type, subject_id,
			request_id, dedupe_key, properties, occurred_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8)
		ON CONFLICT (dedupe_key) WHERE dedupe_key IS NOT NULL DO NOTHING
	`,
		params.EventName,
		params.ActorUserID,
		params.SubjectType,
		params.SubjectID,
		params.RequestID,
		params.DedupeKey,
		string(rawProperties),
		params.OccurredAt,
	)
	if err != nil {
		return fmt.Errorf("record product event: %w", err)
	}
	return nil
}

func (s *Store) GetProductMetrics(
	ctx context.Context,
	tx *sql.Tx,
	from time.Time,
	to time.Time,
) (ProductMetricsSnapshot, error) {
	exec := executor(s.db, tx)
	var snapshot ProductMetricsSnapshot

	if err := exec.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users WHERE deleted_at IS NULL),
			COUNT(DISTINCT actor_user_id) FILTER (WHERE actor_user_id IS NOT NULL),
			COUNT(DISTINCT actor_user_id) FILTER (WHERE event_name = 'intent_created'),
			(
				SELECT COUNT(*)
				FROM (
					SELECT actor_user_id
					FROM product_events
					WHERE occurred_at >= $1 AND occurred_at < $2
					  AND event_name = 'intent_created'
					  AND actor_user_id IS NOT NULL
					GROUP BY actor_user_id
					HAVING COUNT(*) >= 2
				) repeated
			),
			(SELECT COUNT(*) FROM safety_reports WHERE status IN ('open', 'reviewing'))
		FROM product_events
		WHERE occurred_at >= $1 AND occurred_at < $2
	`, from, to).Scan(
		&snapshot.RegisteredUsers,
		&snapshot.ActiveUsers,
		&snapshot.IntentCreators,
		&snapshot.RepeatIntentUsers,
		&snapshot.OpenSafetyReports,
	); err != nil {
		return ProductMetricsSnapshot{}, fmt.Errorf("get product metric summary: %w", err)
	}

	eventRows, err := queryContext(exec, ctx, `
		SELECT event_name,
		       COUNT(*),
		       COUNT(DISTINCT actor_user_id) FILTER (WHERE actor_user_id IS NOT NULL)
		FROM product_events
		WHERE occurred_at >= $1 AND occurred_at < $2
		GROUP BY event_name
		ORDER BY event_name
	`, from, to)
	if err != nil {
		return ProductMetricsSnapshot{}, fmt.Errorf("list product event metrics: %w", err)
	}
	defer eventRows.Close()

	for eventRows.Next() {
		var metric ProductEventMetric
		if err := eventRows.Scan(&metric.EventName, &metric.Count, &metric.UniqueActors); err != nil {
			return ProductMetricsSnapshot{}, fmt.Errorf("scan product event metric: %w", err)
		}
		snapshot.EventMetrics = append(snapshot.EventMetrics, metric)
	}
	if err := eventRows.Err(); err != nil {
		return ProductMetricsSnapshot{}, fmt.Errorf("iterate product event metrics: %w", err)
	}

	dailyRows, err := queryContext(exec, ctx, `
		SELECT TO_CHAR(day, 'YYYY-MM-DD'),
		       COUNT(DISTINCT actor_user_id) FILTER (WHERE actor_user_id IS NOT NULL),
		       COUNT(*) FILTER (WHERE event_name = 'intent_created'),
		       COUNT(*) FILTER (WHERE event_name = 'recommendations_viewed'),
		       COUNT(*) FILTER (WHERE event_name = 'interest_expressed'),
		       COUNT(*) FILTER (WHERE event_name = 'match_created'),
		       COUNT(*) FILTER (WHERE event_name = 'meetup_confirmed'),
		       COUNT(*) FILTER (WHERE event_name = 'meetup_completed'),
		       COUNT(*) FILTER (WHERE event_name = 'safety_report_submitted')
		FROM (
			SELECT DATE_TRUNC('day', occurred_at) AS day,
			       event_name,
			       actor_user_id
			FROM product_events
			WHERE occurred_at >= $1 AND occurred_at < $2
		) events
		GROUP BY day
		ORDER BY day
	`, from, to)
	if err != nil {
		return ProductMetricsSnapshot{}, fmt.Errorf("list daily product metrics: %w", err)
	}
	defer dailyRows.Close()

	for dailyRows.Next() {
		var metric DailyProductMetric
		if err := dailyRows.Scan(
			&metric.Date,
			&metric.ActiveUsers,
			&metric.IntentsCreated,
			&metric.RecommendationsViewed,
			&metric.InterestsExpressed,
			&metric.MatchesCreated,
			&metric.MeetupsConfirmed,
			&metric.MeetupsCompleted,
			&metric.SafetyReports,
		); err != nil {
			return ProductMetricsSnapshot{}, fmt.Errorf("scan daily product metric: %w", err)
		}
		snapshot.DailyMetrics = append(snapshot.DailyMetrics, metric)
	}
	if err := dailyRows.Err(); err != nil {
		return ProductMetricsSnapshot{}, fmt.Errorf("iterate daily product metrics: %w", err)
	}

	return snapshot, nil
}
