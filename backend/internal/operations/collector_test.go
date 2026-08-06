package operations

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestCollectorAggregatesRequestsAndTasks(t *testing.T) {
	collector := NewCollector(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	collector.startedAt = time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC)
	collector.ObserveRequest("get", "/api/v1/items/{itemId}", 200, 10*time.Millisecond)
	collector.ObserveRequest("GET", "/api/v1/items/{itemId}", 404, 20*time.Millisecond)
	collector.ObserveRequest("GET", "/api/v1/items/{itemId}", 503, 30*time.Millisecond)
	collector.ObserveTask(context.Background(), "ai.generate", 40*time.Millisecond, nil)
	collector.ObserveTask(context.Background(), "ai.generate", 60*time.Millisecond, errors.New("timeout"))

	snapshot := collector.Snapshot(collector.startedAt.Add(2 * time.Minute))
	if snapshot.UptimeSeconds != 120 {
		t.Fatalf("uptime = %d, want 120", snapshot.UptimeSeconds)
	}
	if snapshot.Requests.Total != 3 || snapshot.Requests.ClientErrors != 1 || snapshot.Requests.ServerErrors != 1 {
		t.Fatalf("request summary = %+v", snapshot.Requests)
	}
	if snapshot.Requests.AverageDurationMS != 20 || snapshot.Requests.MaxDurationMS != 30 {
		t.Fatalf("request durations = %+v", snapshot.Requests)
	}
	if len(snapshot.Routes) != 1 || snapshot.Routes[0].Route != "/api/v1/items/{itemId}" {
		t.Fatalf("routes = %+v", snapshot.Routes)
	}
	if len(snapshot.Tasks) != 1 || snapshot.Tasks[0].Count != 2 || snapshot.Tasks[0].Failures != 1 ||
		snapshot.Tasks[0].AverageDurationMS != 50 {
		t.Fatalf("tasks = %+v", snapshot.Tasks)
	}
}

func TestCollectorUsesBoundedFallbackNames(t *testing.T) {
	collector := NewCollector(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	collector.ObserveRequest("", "", 500, time.Millisecond)
	collector.ObserveTask(context.Background(), "", time.Millisecond, nil)
	snapshot := collector.Snapshot(time.Now().Add(time.Second))
	if snapshot.Routes[0].Method != "UNKNOWN" || snapshot.Routes[0].Route != "unmatched" {
		t.Fatalf("fallback route = %+v", snapshot.Routes[0])
	}
	if snapshot.Tasks[0].Name != "unknown" {
		t.Fatalf("fallback task = %+v", snapshot.Tasks[0])
	}
}
