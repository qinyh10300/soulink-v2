package analytics

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/store"
)

type captureRecorder struct {
	inputs []EventInput
}

func (r *captureRecorder) Record(_ context.Context, input EventInput) error {
	r.inputs = append(r.inputs, input)
	return nil
}

func TestTrackRecordsSuccessfulAuthenticatedRequest(t *testing.T) {
	t.Parallel()

	recorder := &captureRecorder{}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.With(Track(
		recorder,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		EventRecommendationsViewed,
		"intent",
		"intentId",
	)).Get("/intents/{intentId}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/intents/42", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.Context{
		User: store.User{ID: 7},
	}))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if len(recorder.inputs) != 1 {
		t.Fatalf("recorded inputs = %d, want 1", len(recorder.inputs))
	}
	input := recorder.inputs[0]
	if input.ActorUserID != 7 || input.SubjectID != 42 || input.SubjectType != "intent" {
		t.Fatalf("input = %+v", input)
	}
	if input.RequestID == "" {
		t.Fatal("RequestID is empty")
	}
}

func TestTrackSkipsFailedRequest(t *testing.T) {
	t.Parallel()

	recorder := &captureRecorder{}
	router := chi.NewRouter()
	router.With(Track(recorder, nil, EventIntentCreated, "", "")).Post("/intents", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	req := httptest.NewRequest(http.MethodPost, "/intents", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.Context{
		User: store.User{ID: 7},
	}))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if len(recorder.inputs) != 0 {
		t.Fatalf("recorded inputs = %d, want 0", len(recorder.inputs))
	}
}
