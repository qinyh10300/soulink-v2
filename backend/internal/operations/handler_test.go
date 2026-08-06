package operations

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/store"
)

func TestMetricsRequiresConfiguredAdmin(t *testing.T) {
	handler := NewHandler(
		NewCollector(slog.New(slog.NewJSONHandler(io.Discard, nil))),
		[]string{"admin@mails.tsinghua.edu.cn"},
	)
	for _, test := range []struct {
		name   string
		email  string
		status int
	}{
		{name: "admin", email: " ADMIN@mails.tsinghua.edu.cn ", status: http.StatusOK},
		{name: "member", email: "member@mails.tsinghua.edu.cn", status: http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/operational-metrics", nil)
			req = req.WithContext(auth.WithContext(req.Context(), auth.Context{
				User: store.User{ID: 1, Email: test.email},
			}))
			recorder := httptest.NewRecorder()
			handler.Metrics(recorder, req)
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, test.status, recorder.Body.String())
			}
		})
	}
}
