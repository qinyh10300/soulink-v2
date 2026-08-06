package operations

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type requestKey struct {
	Method string
	Route  string
}

type aggregate struct {
	Count         int64
	ClientErrors  int64
	ServerErrors  int64
	Failures      int64
	TotalDuration time.Duration
	MaxDuration   time.Duration
}

type Collector struct {
	mu        sync.RWMutex
	startedAt time.Time
	logger    *slog.Logger
	requests  map[requestKey]*aggregate
	tasks     map[string]*aggregate
}

type RequestSummary struct {
	Total             int64   `json:"total"`
	ClientErrors      int64   `json:"clientErrors"`
	ServerErrors      int64   `json:"serverErrors"`
	AverageDurationMS float64 `json:"averageDurationMs"`
	MaxDurationMS     float64 `json:"maxDurationMs"`
}

type RouteMetric struct {
	Method            string  `json:"method"`
	Route             string  `json:"route"`
	Count             int64   `json:"count"`
	ClientErrors      int64   `json:"clientErrors"`
	ServerErrors      int64   `json:"serverErrors"`
	AverageDurationMS float64 `json:"averageDurationMs"`
	MaxDurationMS     float64 `json:"maxDurationMs"`
}

type TaskMetric struct {
	Name              string  `json:"name"`
	Count             int64   `json:"count"`
	Failures          int64   `json:"failures"`
	AverageDurationMS float64 `json:"averageDurationMs"`
	MaxDurationMS     float64 `json:"maxDurationMs"`
}

type Snapshot struct {
	StartedAt     time.Time      `json:"startedAt"`
	UptimeSeconds int64          `json:"uptimeSeconds"`
	Requests      RequestSummary `json:"requests"`
	Routes        []RouteMetric  `json:"routes"`
	Tasks         []TaskMetric   `json:"tasks"`
	GeneratedAt   time.Time      `json:"generatedAt"`
}

func NewCollector(logger *slog.Logger) *Collector {
	if logger == nil {
		logger = slog.Default()
	}
	return &Collector{
		startedAt: time.Now().UTC(),
		logger:    logger,
		requests:  make(map[requestKey]*aggregate),
		tasks:     make(map[string]*aggregate),
	}
}

func (c *Collector) ObserveRequest(method, route string, status int, duration time.Duration) {
	if c == nil {
		return
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "UNKNOWN"
	}
	route = strings.TrimSpace(route)
	if route == "" {
		route = "unmatched"
	}
	c.mu.Lock()
	metric := c.requests[requestKey{Method: method, Route: route}]
	if metric == nil {
		metric = &aggregate{}
		c.requests[requestKey{Method: method, Route: route}] = metric
	}
	metric.Count++
	metric.TotalDuration += duration
	if duration > metric.MaxDuration {
		metric.MaxDuration = duration
	}
	if status >= 400 && status < 500 {
		metric.ClientErrors++
	}
	if status >= 500 {
		metric.ServerErrors++
		metric.Failures++
	}
	c.mu.Unlock()
}

func (c *Collector) ObserveTask(ctx context.Context, name string, duration time.Duration, taskErr error) {
	if c == nil {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "unknown"
	}
	c.mu.Lock()
	metric := c.tasks[name]
	if metric == nil {
		metric = &aggregate{}
		c.tasks[name] = metric
	}
	metric.Count++
	metric.TotalDuration += duration
	if duration > metric.MaxDuration {
		metric.MaxDuration = duration
	}
	if taskErr != nil {
		metric.Failures++
	}
	c.mu.Unlock()

	args := []any{
		"task", name,
		"success", taskErr == nil,
		"duration_ms", durationMilliseconds(duration),
	}
	if requestID := middleware.GetReqID(ctx); requestID != "" {
		args = append(args, "request_id", requestID)
	}
	if taskErr != nil {
		args = append(args, "error", taskErr)
		c.logger.Error("task completed", args...)
		return
	}
	c.logger.Info("task completed", args...)
}

func (c *Collector) Snapshot(now time.Time) Snapshot {
	if c == nil {
		return Snapshot{}
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshot := Snapshot{
		StartedAt:     c.startedAt,
		UptimeSeconds: max(0, int64(now.Sub(c.startedAt).Seconds())),
		Routes:        make([]RouteMetric, 0, len(c.requests)),
		Tasks:         make([]TaskMetric, 0, len(c.tasks)),
		GeneratedAt:   now,
	}
	var totalDuration time.Duration
	var maxDuration time.Duration
	for key, metric := range c.requests {
		snapshot.Requests.Total += metric.Count
		snapshot.Requests.ClientErrors += metric.ClientErrors
		snapshot.Requests.ServerErrors += metric.ServerErrors
		totalDuration += metric.TotalDuration
		if metric.MaxDuration > maxDuration {
			maxDuration = metric.MaxDuration
		}
		snapshot.Routes = append(snapshot.Routes, RouteMetric{
			Method: key.Method, Route: key.Route, Count: metric.Count,
			ClientErrors: metric.ClientErrors, ServerErrors: metric.ServerErrors,
			AverageDurationMS: averageDurationMilliseconds(metric),
			MaxDurationMS:     durationMilliseconds(metric.MaxDuration),
		})
	}
	if snapshot.Requests.Total > 0 {
		snapshot.Requests.AverageDurationMS = durationMilliseconds(totalDuration / time.Duration(snapshot.Requests.Total))
	}
	snapshot.Requests.MaxDurationMS = durationMilliseconds(maxDuration)
	for name, metric := range c.tasks {
		snapshot.Tasks = append(snapshot.Tasks, TaskMetric{
			Name: name, Count: metric.Count, Failures: metric.Failures,
			AverageDurationMS: averageDurationMilliseconds(metric),
			MaxDurationMS:     durationMilliseconds(metric.MaxDuration),
		})
	}
	sort.Slice(snapshot.Routes, func(i, j int) bool {
		if snapshot.Routes[i].Count != snapshot.Routes[j].Count {
			return snapshot.Routes[i].Count > snapshot.Routes[j].Count
		}
		if snapshot.Routes[i].Route != snapshot.Routes[j].Route {
			return snapshot.Routes[i].Route < snapshot.Routes[j].Route
		}
		return snapshot.Routes[i].Method < snapshot.Routes[j].Method
	})
	sort.Slice(snapshot.Tasks, func(i, j int) bool {
		return snapshot.Tasks[i].Name < snapshot.Tasks[j].Name
	})
	return snapshot
}

func averageDurationMilliseconds(metric *aggregate) float64 {
	if metric == nil || metric.Count == 0 {
		return 0
	}
	return durationMilliseconds(metric.TotalDuration / time.Duration(metric.Count))
}

func durationMilliseconds(duration time.Duration) float64 {
	return float64(duration.Microseconds()) / 1000
}
