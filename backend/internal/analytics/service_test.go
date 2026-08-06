package analytics

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type fakeMetricsStore struct {
	recorded  store.ProductEventParams
	snapshot  store.ProductMetricsSnapshot
	recordErr error
	queryErr  error
}

func (s *fakeMetricsStore) RecordProductEvent(
	_ context.Context,
	_ *sql.Tx,
	params store.ProductEventParams,
) error {
	s.recorded = params
	return s.recordErr
}

func (s *fakeMetricsStore) GetProductMetrics(
	_ context.Context,
	_ *sql.Tx,
	_ time.Time,
	_ time.Time,
) (store.ProductMetricsSnapshot, error) {
	return s.snapshot, s.queryErr
}

func TestRecordSanitizesPropertiesAndBuildsDedupeKey(t *testing.T) {
	t.Parallel()

	dataStore := &fakeMetricsStore{}
	service := NewService(dataStore, nil)
	now := time.Date(2026, time.August, 7, 8, 30, 0, 0, time.UTC)
	service.nowFunc = func() time.Time { return now }

	err := service.Record(t.Context(), EventInput{
		Name:        EventRecommendationsViewed,
		ActorUserID: 12,
		SubjectType: "intent",
		SubjectID:   34,
		RequestID:   "request-1",
		Properties: map[string]any{
			"algorithmVersion": "rules-v1",
			"messageBody":      "must not be stored",
		},
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	if dataStore.recorded.EventName != EventRecommendationsViewed {
		t.Fatalf("EventName = %q", dataStore.recorded.EventName)
	}
	if !dataStore.recorded.ActorUserID.Valid || dataStore.recorded.ActorUserID.Int64 != 12 {
		t.Fatalf("ActorUserID = %+v", dataStore.recorded.ActorUserID)
	}
	if dataStore.recorded.DedupeKey.String != "request:request-1:recommendations_viewed:intent:34" {
		t.Fatalf("DedupeKey = %q", dataStore.recorded.DedupeKey.String)
	}
	if len(dataStore.recorded.Properties) != 1 || dataStore.recorded.Properties["algorithmVersion"] != "rules-v1" {
		t.Fatalf("Properties = %#v", dataStore.recorded.Properties)
	}
	if !dataStore.recorded.OccurredAt.Equal(now) {
		t.Fatalf("OccurredAt = %v, want %v", dataStore.recorded.OccurredAt, now)
	}
}

func TestRecordRejectsUnknownEventAndIncompleteSubject(t *testing.T) {
	t.Parallel()

	service := NewService(&fakeMetricsStore{}, nil)
	if err := service.Record(t.Context(), EventInput{Name: "custom_client_event"}); err == nil {
		t.Fatal("Record() unknown event error = nil")
	}
	if err := service.Record(t.Context(), EventInput{
		Name:        EventInterestExpressed,
		SubjectType: "intent",
	}); err == nil {
		t.Fatal("Record() incomplete subject error = nil")
	}
}

func TestDashboardBuildsFunnelAndRates(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)
	dataStore := &fakeMetricsStore{snapshot: store.ProductMetricsSnapshot{
		RegisteredUsers:   100,
		ActiveUsers:       20,
		IntentCreators:    10,
		RepeatIntentUsers: 2,
		OpenSafetyReports: 3,
		EventMetrics: []store.ProductEventMetric{
			{EventName: EventIntentCreated, Count: 12, UniqueActors: 10},
			{EventName: EventRecommendationsViewed, Count: 16, UniqueActors: 8},
			{EventName: EventInterestExpressed, Count: 5, UniqueActors: 4},
			{EventName: EventMatchCreated, Count: 2, UniqueActors: 2},
			{EventName: EventMeetupConfirmed, Count: 1, UniqueActors: 1},
			{EventName: EventMeetupCompleted, Count: 1, UniqueActors: 1},
			{EventName: EventSafetyReportSubmitted, Count: 2, UniqueActors: 2},
		},
		DailyMetrics: []store.DailyProductMetric{
			{Date: "2026-08-07", ActiveUsers: 5, IntentsCreated: 2, MeetupsCompleted: 1},
		},
	}}
	service := NewService(dataStore, []string{"admin@mails.tsinghua.edu.cn"})
	service.nowFunc = func() time.Time { return now }

	dashboard, err := service.Dashboard(t.Context(), " ADMIN@mails.tsinghua.edu.cn ", 7)
	if err != nil {
		t.Fatalf("Dashboard() error = %v", err)
	}
	if dashboard.NorthStar.MeetupsCompleted != 1 || dashboard.NorthStar.MatchToMeetupRate != 50 {
		t.Fatalf("NorthStar = %+v", dashboard.NorthStar)
	}
	if dashboard.NorthStar.RepeatIntentRate != 20 {
		t.Fatalf("RepeatIntentRate = %v", dashboard.NorthStar.RepeatIntentRate)
	}
	if dashboard.Safety.ReportRate != 10 {
		t.Fatalf("ReportRate = %v", dashboard.Safety.ReportRate)
	}
	if len(dashboard.Funnel) != 6 || dashboard.Funnel[1].ConversionFromPrevious != 80 {
		t.Fatalf("Funnel = %+v", dashboard.Funnel)
	}
	if len(dashboard.Daily) != 7 || dashboard.Daily[6].Date != "2026-08-07" || dashboard.Daily[6].ActiveUsers != 5 {
		t.Fatalf("Daily = %+v", dashboard.Daily)
	}
}

func TestDashboardRequiresAdminAndSupportedPeriod(t *testing.T) {
	t.Parallel()

	service := NewService(&fakeMetricsStore{}, []string{"admin@mails.tsinghua.edu.cn"})
	_, err := service.Dashboard(t.Context(), "student@mails.tsinghua.edu.cn", 28)
	if appErr := httpapi.AsAppError(err); appErr.Code != httpapi.CodeForbidden {
		t.Fatalf("non-admin code = %d, want %d", appErr.Code, httpapi.CodeForbidden)
	}
	_, err = service.Dashboard(t.Context(), "admin@mails.tsinghua.edu.cn", 30)
	if appErr := httpapi.AsAppError(err); appErr.Code != httpapi.CodeInvalidParams {
		t.Fatalf("invalid period code = %d, want %d", appErr.Code, httpapi.CodeInvalidParams)
	}
}
