package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/gate"
	"soulink-backend/internal/match"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/profile"
	"soulink-backend/internal/questionnaire"
)

type fakeGateChecker struct {
	status gate.Status
	err    error
}

func (f fakeGateChecker) Check(_ context.Context, _ string) (gate.Status, error) {
	return f.status, f.err
}

func TestGateWindowMiddlewareBlocksActiveWindow(t *testing.T) {
	t.Parallel()

	checker := fakeGateChecker{status: gate.Status{
		GateKey:  gate.GateKeyMatchingWrite,
		StartsAt: time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
		Reason:   "matching running",
		Active:   true,
	}}
	middleware := GateWindowMiddleware(slog.New(slog.NewTextHandler(io.Discard, nil)), checker, gate.GateKeyMatchingWrite, "fallback")

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/profile", nil)
	rec := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("gate middleware status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestNewRouterMatchWriteGuardBlocksSelectedWriteRoutes(t *testing.T) {
	t.Parallel()

	blockGuard := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})
	}

	router := NewRouter(Dependencies{
		Logger:               slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedOrigins:       []string{"http://localhost:5173"},
		AuthMiddleware:       func(next http.Handler) http.Handler { return next },
		MatchWriteGuard:      blockGuard,
		AuthHandler:          auth.NewHandler(nil),
		ProfileHandler:       profile.NewHandler(nil),
		MetaHandler:          meta.NewHandler(meta.NewService()),
		QuestionnaireHandler: questionnaire.NewHandler(),
		MatchHandler:         match.NewHandler(),
	})

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "profile_update", method: http.MethodPut, path: "/api/v1/users/me/profile"},
		{name: "contact_upsert", method: http.MethodPut, path: "/api/v1/users/me/contacts/wechat"},
		{name: "contact_delete", method: http.MethodDelete, path: "/api/v1/users/me/contacts/wechat"},
		{name: "questionnaire_batch", method: http.MethodPut, path: "/api/v1/questionnaires/current/answers"},
		{name: "questionnaire_single", method: http.MethodPut, path: "/api/v1/questionnaires/current/answers/1"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("%s %s status = %d, want %d", tc.method, tc.path, rec.Code, http.StatusForbidden)
			}
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/questionnaires/current", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code == http.StatusForbidden {
		t.Fatalf("GET /api/v1/questionnaires/current status = %d, should not be blocked by match write guard", rec.Code)
	}
}
