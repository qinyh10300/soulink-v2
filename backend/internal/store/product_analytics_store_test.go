package store

import (
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestRecordProductEventUsesDedupeConflict(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	dataStore := New(db)
	now := time.Date(2026, time.August, 7, 10, 0, 0, 0, time.UTC)
	mock.ExpectExec("INSERT INTO product_events").
		WithArgs(
			"intent_created",
			sql.NullInt64{Int64: 7, Valid: true},
			"intent",
			sql.NullInt64{Int64: 12, Valid: true},
			"request-1",
			sql.NullString{String: "request:request-1:intent_created:intent:12", Valid: true},
			sqlmock.AnyArg(),
			now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = dataStore.RecordProductEvent(t.Context(), nil, ProductEventParams{
		EventName:   "intent_created",
		ActorUserID: sql.NullInt64{Int64: 7, Valid: true},
		SubjectType: "intent",
		SubjectID:   sql.NullInt64{Int64: 12, Valid: true},
		RequestID:   "request-1",
		DedupeKey:   sql.NullString{String: "request:request-1:intent_created:intent:12", Valid: true},
		Properties:  map[string]any{"source": "api"},
		OccurredAt:  now,
	})
	if err != nil {
		t.Fatalf("RecordProductEvent() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetProductMetricsScansSummaryEventsAndDailyRows(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()
	dataStore := New(db)
	from := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.August, 8, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT[[:space:]]+\\(SELECT COUNT\\(\\*\\) FROM users").
		WithArgs(from, to).
		WillReturnRows(sqlmock.NewRows([]string{
			"registered_users", "active_users", "intent_creators", "repeat_intent_users", "open_reports",
		}).AddRow(50, 20, 10, 3, 2))
	mock.ExpectQuery("SELECT event_name").
		WithArgs(from, to).
		WillReturnRows(sqlmock.NewRows([]string{"event_name", "count", "unique_actors"}).
			AddRow("intent_created", 12, 10))
	mock.ExpectQuery("SELECT TO_CHAR\\(day").
		WithArgs(from, to).
		WillReturnRows(sqlmock.NewRows([]string{
			"date", "active_users", "intents_created", "recommendations_viewed",
			"interests_expressed", "matches_created", "meetups_confirmed",
			"meetups_completed", "safety_reports",
		}).AddRow("2026-08-07", 5, 2, 3, 1, 1, 0, 0, 0))

	snapshot, err := dataStore.GetProductMetrics(t.Context(), nil, from, to)
	if err != nil {
		t.Fatalf("GetProductMetrics() error = %v", err)
	}
	if snapshot.RegisteredUsers != 50 || snapshot.RepeatIntentUsers != 3 {
		t.Fatalf("summary = %+v", snapshot)
	}
	if len(snapshot.EventMetrics) != 1 || snapshot.EventMetrics[0].Count != 12 {
		t.Fatalf("event metrics = %+v", snapshot.EventMetrics)
	}
	if len(snapshot.DailyMetrics) != 1 || snapshot.DailyMetrics[0].ActiveUsers != 5 {
		t.Fatalf("daily metrics = %+v", snapshot.DailyMetrics)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
