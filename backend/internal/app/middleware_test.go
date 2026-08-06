package app

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type requestObservation struct {
	method   string
	route    string
	status   int
	duration time.Duration
}

type requestObserverStub struct {
	last  requestObservation
	calls int
}

func (observer *requestObserverStub) ObserveRequest(method, route string, status int, duration time.Duration) {
	observer.last = requestObservation{method: method, route: route, status: status, duration: duration}
	observer.calls++
}

func TestRequestLoggerReturnsCorrelationIDAndUsesRoutePattern(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	observer := &requestObserverStub{}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(RequestLogger(logger, observer))
	router.Get("/items/{itemId}", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "failed", http.StatusServiceUnavailable)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/items/42", nil))
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("X-Request-ID response header is empty")
	}
	if observer.last.method != http.MethodGet || observer.last.route != "/items/{itemId}" ||
		observer.last.status != http.StatusServiceUnavailable || observer.last.duration <= 0 {
		t.Fatalf("observation = %+v", observer.last)
	}
	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("log is not JSON: %v; output = %s", err, output.String())
	}
	if entry["route"] != "/items/{itemId}" || entry["level"] != "ERROR" {
		t.Fatalf("log entry = %+v", entry)
	}
	if _, leaked := entry["path"]; leaked {
		t.Fatalf("log entry contains raw path: %+v", entry)
	}
}

func TestRequestLoggerExcludesWebSocketLifetimeFromRequestMetrics(t *testing.T) {
	observer := &requestObserverStub{}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(RequestLogger(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), observer))
	router.Get("/ws", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusSwitchingProtocols)
	})

	request := httptest.NewRequest(http.MethodGet, "/ws", nil)
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "websocket")
	router.ServeHTTP(httptest.NewRecorder(), request)

	if observer.calls != 0 {
		t.Fatalf("observer calls = %d, want 0", observer.calls)
	}
}
