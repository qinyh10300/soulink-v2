package questionnaire

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

func TestBuildAnswerSnapshotIgnoresQuestionsOutsideCurrentCatalog(t *testing.T) {
	catalog := &Catalog{
		Key:     "campus-match",
		Version: "v1.2.1",
		Questions: []Question{
			{ID: "1", Order: 1, Title: "gender", Required: true},
		},
		RequiredCount: 1,
		byID: map[string]Question{
			"1": {ID: "1", Order: 1, Title: "gender", Required: true},
		},
	}
	service := &Service{catalog: catalog}
	now := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)

	validValue, err := json.Marshal("male")
	if err != nil {
		t.Fatalf("json.Marshal(validValue) error = %v", err)
	}
	staleValue, err := json.Marshal("stale-answer")
	if err != nil {
		t.Fatalf("json.Marshal(staleValue) error = %v", err)
	}

	answeredCount, requiredAnsweredCount, answers, err := service.buildAnswerSnapshot([]store.QuestionnaireAnswer{
		{
			QuestionnaireKey:     "campus-match",
			QuestionnaireVersion: "v1.2.1",
			QuestionID:           "1",
			ValueJSON:            validValue,
			UpdatedAt:            now,
		},
		{
			QuestionnaireKey:     "campus-match",
			QuestionnaireVersion: "v1.0.0",
			QuestionID:           "999",
			ValueJSON:            staleValue,
			UpdatedAt:            now.Add(-time.Hour),
		},
	})
	if err != nil {
		t.Fatalf("buildAnswerSnapshot() error = %v", err)
	}

	if answeredCount != 1 {
		t.Fatalf("answeredCount = %d, want %d", answeredCount, 1)
	}
	if requiredAnsweredCount != 1 {
		t.Fatalf("requiredAnsweredCount = %d, want %d", requiredAnsweredCount, 1)
	}
	if len(answers) != 1 {
		t.Fatalf("len(answers) = %d, want %d", len(answers), 1)
	}
	if answers[0].QuestionID != "1" {
		t.Fatalf("answers[0].QuestionID = %q, want %q", answers[0].QuestionID, "1")
	}
}

func TestSaveAnswerRejectsWhenCollectionHasNotStarted(t *testing.T) {
	start := time.Date(2026, 3, 26, 9, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	question := Question{
		ID:       "1",
		Order:    1,
		Title:    "gender",
		Type:     QuestionTypeSingleChoice,
		Required: true,
		Options: []Option{
			{Label: "male", Value: "male"},
			{Label: "female", Value: "female"},
		},
	}
	service := &Service{
		store: &store.Store{},
		catalog: &Catalog{
			Key:                "campus-match",
			Version:            "v2.0.0",
			CollectionStartsAt: &start,
			CollectionEndsAt:   &end,
			CollectionTimezone: "Asia/Shanghai",
			Questions:          []Question{question},
			RequiredCount:      1,
			byID: map[string]Question{
				"1": question,
			},
		},
		nowFunc: func() time.Time {
			return start.Add(-time.Minute)
		},
	}

	_, err := service.SaveAnswer(context.Background(), 1001, SaveAnswerInput{
		QuestionID: "1",
		Value:      json.RawMessage(`"male"`),
	})
	if err == nil {
		t.Fatal("SaveAnswer() error = nil, want collection closed error")
	}
	appErr := httpapi.AsAppError(err)
	if appErr.HTTPStatus != http.StatusForbidden {
		t.Fatalf("SaveAnswer() status = %d, want %d", appErr.HTTPStatus, http.StatusForbidden)
	}
	if appErr.Code != httpapi.CodeForbidden {
		t.Fatalf("SaveAnswer() code = %d, want %d", appErr.Code, httpapi.CodeForbidden)
	}
}

func TestUpdateParticipationRequiresCompletedQuestionnaire(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	service := NewService(store.New(db), &Catalog{Key: "campus-match", Version: "v1"}, WithNowFunc(func() time.Time { return now }))

	mock.ExpectQuery(`SELECT user_id, questionnaire_key, questionnaire_title, questionnaire_version, total_count,
               required_count, answered_count, required_answered_count, progress_percent,
               has_completed_questionnaire, participating_in_match, last_action, last_saved_at, result_reveal_at,
               created_at, updated_at
        FROM user_questionnaire_states`).
		WithArgs(int64(1001), "campus-match", "v1").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "questionnaire_key", "questionnaire_title", "questionnaire_version",
			"total_count", "required_count", "answered_count", "required_answered_count",
			"progress_percent", "has_completed_questionnaire", "participating_in_match",
			"last_action", "last_saved_at", "result_reveal_at", "created_at", "updated_at",
		}).AddRow(
			int64(1001), "campus-match", "Campus Match", "v1",
			10, 10, 5, 5,
			50, false, false,
			"save", now, sql.NullTime{}, now, now,
		))

	_, err = service.UpdateParticipation(context.Background(), 1001, true)
	if err == nil {
		t.Fatal("UpdateParticipation() error = nil, want forbidden error")
	}
	appErr := httpapi.AsAppError(err)
	if appErr.HTTPStatus != http.StatusForbidden {
		t.Fatalf("UpdateParticipation() status = %d, want %d", appErr.HTTPStatus, http.StatusForbidden)
	}
	if appErr.Code != httpapi.CodeForbidden {
		t.Fatalf("UpdateParticipation() code = %d, want %d", appErr.Code, httpapi.CodeForbidden)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestUpdateParticipationPersistsChoiceForCompletedQuestionnaire(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	revealAt := now.Add(9 * time.Hour)
	service := NewService(store.New(db), &Catalog{Key: "campus-match", Version: "v1"}, WithNowFunc(func() time.Time { return now }))

	mock.ExpectQuery(`SELECT user_id, questionnaire_key, questionnaire_title, questionnaire_version, total_count,
               required_count, answered_count, required_answered_count, progress_percent,
               has_completed_questionnaire, participating_in_match, last_action, last_saved_at, result_reveal_at,
               created_at, updated_at
        FROM user_questionnaire_states`).
		WithArgs(int64(1001), "campus-match", "v1").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "questionnaire_key", "questionnaire_title", "questionnaire_version",
			"total_count", "required_count", "answered_count", "required_answered_count",
			"progress_percent", "has_completed_questionnaire", "participating_in_match",
			"last_action", "last_saved_at", "result_reveal_at", "created_at", "updated_at",
		}).AddRow(
			int64(1001), "campus-match", "Campus Match", "v1",
			10, 10, 10, 10,
			100, true, false,
			"submit", now, revealAt, now, now,
		))
	mock.ExpectExec(`UPDATE user_questionnaire_states`).
		WithArgs(int64(1001), "campus-match", "v1", true, now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := service.UpdateParticipation(context.Background(), 1001, true)
	if err != nil {
		t.Fatalf("UpdateParticipation() error = %v", err)
	}
	if !result.HasCompletedQuestionnaire {
		t.Fatal("UpdateParticipation() hasCompletedQuestionnaire = false, want true")
	}
	if !result.ParticipatingInMatch {
		t.Fatal("UpdateParticipation() participatingInMatch = false, want true")
	}
	if result.ResultRevealAt != revealAt.Format(time.RFC3339) {
		t.Fatalf("UpdateParticipation() resultRevealAt = %q, want %q", result.ResultRevealAt, revealAt.Format(time.RFC3339))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}
