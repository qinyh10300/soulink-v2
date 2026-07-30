package app

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/feedback"
	"soulink-backend/internal/match"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/profile"
	"soulink-backend/internal/questionnaire"
	"soulink-backend/internal/school"
)

type routerUserCountStore struct {
	counts map[school.Name]int
}

func (s routerUserCountStore) CountActiveUsersByEmailSuffixes(_ context.Context, _ *sql.Tx, suffixes []string) (int, error) {
	for _, schoolName := range school.Names() {
		schoolSuffixes := school.EmailSuffixes(schoolName)
		if len(suffixes) == 0 || len(schoolSuffixes) == 0 {
			continue
		}
		if suffixes[0] == schoolSuffixes[0] {
			return s.counts[schoolName], nil
		}
	}
	return 0, nil
}

func TestNewRouterHealthz(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{
		Logger:               slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedOrigins:       []string{"http://localhost:5173"},
		AuthMiddleware:       func(next http.Handler) http.Handler { return next },
		AuthHandler:          auth.NewHandler(nil),
		FeedbackHandler:      feedback.NewHandler(nil),
		ProfileHandler:       profile.NewHandler(nil),
		MetaHandler:          meta.NewHandler(meta.NewService()),
		QuestionnaireHandler: questionnaire.NewHandler(),
		MatchHandler:         match.NewHandler(),
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestNewRouterMetaEndpoint(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{
		Logger:               slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedOrigins:       []string{"http://localhost:5173"},
		AuthMiddleware:       func(next http.Handler) http.Handler { return next },
		AuthHandler:          auth.NewHandler(nil),
		FeedbackHandler:      feedback.NewHandler(nil),
		ProfileHandler:       profile.NewHandler(nil),
		MetaHandler:          meta.NewHandler(meta.NewService()),
		QuestionnaireHandler: questionnaire.NewHandler(),
		MatchHandler:         match.NewHandler(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta/departments", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/meta/departments status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestNewRouterMetaEndpointSupportsSchoolEmailQuery(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{
		Logger:               slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedOrigins:       []string{"http://localhost:5173"},
		AuthMiddleware:       func(next http.Handler) http.Handler { return next },
		AuthHandler:          auth.NewHandler(nil),
		FeedbackHandler:      feedback.NewHandler(nil),
		ProfileHandler:       profile.NewHandler(nil),
		MetaHandler:          meta.NewHandler(meta.NewService()),
		QuestionnaireHandler: questionnaire.NewHandler(),
		MatchHandler:         match.NewHandler(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta/departments?email=test@pku.edu.cn", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/meta/departments?email=test@pku.edu.cn status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "计算机学院") {
		t.Fatalf("response body = %q, want PKU departments", rec.Body.String())
	}
}

func TestNewRouterRegisteredUserCountEndpoint(t *testing.T) {
	t.Parallel()

	metaService := meta.NewService(meta.WithUserCountStore(routerUserCountStore{
		counts: map[school.Name]int{school.Tsinghua: 42},
	}))
	router := NewRouter(Dependencies{
		Logger:               slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedOrigins:       []string{"http://localhost:5173"},
		AuthMiddleware:       func(next http.Handler) http.Handler { return next },
		AuthHandler:          auth.NewHandler(nil),
		FeedbackHandler:      feedback.NewHandler(nil),
		ProfileHandler:       profile.NewHandler(nil),
		MetaHandler:          meta.NewHandler(metaService),
		QuestionnaireHandler: questionnaire.NewHandler(),
		MatchHandler:         match.NewHandler(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta/registered-user-count", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/meta/registered-user-count status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "registeredUserCount") {
		t.Fatalf("response body = %q, want registeredUserCount field", rec.Body.String())
	}
}

func TestNewRouterPKURegisteredUserCountEndpoint(t *testing.T) {
	t.Parallel()

	metaService := meta.NewService(meta.WithUserCountStore(routerUserCountStore{
		counts: map[school.Name]int{school.Peking: 5},
	}))
	router := NewRouter(Dependencies{
		Logger:               slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedOrigins:       []string{"http://localhost:5173"},
		AuthMiddleware:       func(next http.Handler) http.Handler { return next },
		AuthHandler:          auth.NewHandler(nil),
		FeedbackHandler:      feedback.NewHandler(nil),
		ProfileHandler:       profile.NewHandler(nil),
		MetaHandler:          meta.NewHandler(metaService),
		QuestionnaireHandler: questionnaire.NewHandler(),
		MatchHandler:         match.NewHandler(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta/pku-registered-user-count", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/meta/pku-registered-user-count status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "registeredUserCount") {
		t.Fatalf("response body = %q, want registeredUserCount field", rec.Body.String())
	}
}

func TestNewRouterAdditionalRegisteredUserCountEndpoints(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		path       string
		schoolName school.Name
		count      int
		wantBody   string
	}{
		{
			name:       "bnu",
			path:       "/api/v1/meta/bnu-registered-user-count",
			schoolName: school.BeijingNormal,
			count:      3,
			wantBody:   "3",
		},
		{
			name:       "ruc",
			path:       "/api/v1/meta/ruc-registered-user-count",
			schoolName: school.Renmin,
			count:      2,
			wantBody:   "2",
		},
		{
			name:       "cuc",
			path:       "/api/v1/meta/cuc-registered-user-count",
			schoolName: school.Communication,
			count:      4,
			wantBody:   "4",
		},
		{
			name:       "bfsu",
			path:       "/api/v1/meta/bfsu-registered-user-count",
			schoolName: school.Beiwai,
			count:      1,
			wantBody:   "1",
		},
		{
			name:       "sjtu",
			path:       "/api/v1/meta/sjtu-registered-user-count",
			schoolName: school.ShanghaiJiao,
			count:      1,
			wantBody:   "2",
		},
		{
			name:       "fdu",
			path:       "/api/v1/meta/fdu-registered-user-count",
			schoolName: school.Fudan,
			count:      1,
			wantBody:   "3",
		},
		{
			name:       "nus",
			path:       "/api/v1/meta/nus-registered-user-count",
			schoolName: school.NUS,
			count:      1,
			wantBody:   "4",
		},
		{
			name:       "ntu",
			path:       "/api/v1/meta/ntu-registered-user-count",
			schoolName: school.NTU,
			count:      1,
			wantBody:   "5",
		},
		{
			name:       "zuel",
			path:       "/api/v1/meta/zuel-registered-user-count",
			schoolName: school.ZUEL,
			count:      1,
			wantBody:   "6",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			metaService := meta.NewService(meta.WithUserCountStore(routerUserCountStore{
				counts: map[school.Name]int{tc.schoolName: tc.count},
			}))
			router := NewRouter(Dependencies{
				Logger:               slog.New(slog.NewTextHandler(io.Discard, nil)),
				AllowedOrigins:       []string{"http://localhost:5173"},
				AuthMiddleware:       func(next http.Handler) http.Handler { return next },
				AuthHandler:          auth.NewHandler(nil),
				FeedbackHandler:      feedback.NewHandler(nil),
				ProfileHandler:       profile.NewHandler(nil),
				MetaHandler:          meta.NewHandler(metaService),
				QuestionnaireHandler: questionnaire.NewHandler(),
				MatchHandler:         match.NewHandler(),
			})

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", tc.path, rec.Code, http.StatusOK)
			}
			if !strings.Contains(rec.Body.String(), "registeredUserCount") {
				t.Fatalf("response body = %q, want registeredUserCount field", rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Fatalf("response body = %q, want count %s", rec.Body.String(), tc.wantBody)
			}
		})
	}
}

func TestNewRouterQuestionnairePlaceholder(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{
		Logger:               slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedOrigins:       []string{"http://localhost:5173"},
		AuthMiddleware:       func(next http.Handler) http.Handler { return next },
		AuthHandler:          auth.NewHandler(nil),
		FeedbackHandler:      feedback.NewHandler(nil),
		ProfileHandler:       profile.NewHandler(nil),
		MetaHandler:          meta.NewHandler(meta.NewService()),
		QuestionnaireHandler: questionnaire.NewHandler(),
		MatchHandler:         match.NewHandler(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/questionnaires/current", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("GET /api/v1/questionnaires/current status = %d, want %d", rec.Code, http.StatusNotImplemented)
	}
}
