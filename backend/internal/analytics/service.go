package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type MetricsStore interface {
	RecordProductEvent(context.Context, *sql.Tx, store.ProductEventParams) error
	GetProductMetrics(context.Context, *sql.Tx, time.Time, time.Time) (store.ProductMetricsSnapshot, error)
}

type Service struct {
	store       MetricsStore
	adminEmails map[string]struct{}
	nowFunc     func() time.Time
}

var supportedEvents = map[string]struct{}{
	EventIntentCreated:            {},
	EventIntentPublished:          {},
	EventRecommendationsViewed:    {},
	EventInterestExpressed:        {},
	EventInterestAccepted:         {},
	EventMatchCreated:             {},
	EventMeetupProposed:           {},
	EventMeetupConfirmed:          {},
	EventMeetupCompleted:          {},
	EventMeetupNoShow:             {},
	EventMeetingFeedbackSubmitted: {},
	EventPlazaGroupCreated:        {},
	EventGroupMatchCreated:        {},
	EventSafetyReportSubmitted:    {},
	EventUserBlocked:              {},
}

var allowedPropertyKeys = map[string]struct{}{
	"algorithmVersion": {},
	"category":         {},
	"memberCount":      {},
	"source":           {},
	"status":           {},
}

func NewService(dataStore MetricsStore, adminEmails []string) *Service {
	admins := make(map[string]struct{}, len(adminEmails))
	for _, email := range adminEmails {
		normalized := normalizeEmail(email)
		if normalized != "" {
			admins[normalized] = struct{}{}
		}
	}
	return &Service{
		store:       dataStore,
		adminEmails: admins,
		nowFunc:     time.Now,
	}
}

func (s *Service) Record(ctx context.Context, input EventInput) error {
	if s == nil || s.store == nil {
		return httpapi.FeatureNotReady("product analytics")
	}
	if _, ok := supportedEvents[input.Name]; !ok {
		return fmt.Errorf("unsupported product event %q", input.Name)
	}
	if (input.SubjectType == "") != (input.SubjectID <= 0) {
		return fmt.Errorf("product event subject type and id must be provided together")
	}

	actorID := sql.NullInt64{Int64: input.ActorUserID, Valid: input.ActorUserID > 0}
	subjectID := sql.NullInt64{Int64: input.SubjectID, Valid: input.SubjectID > 0}
	dedupeKey := sql.NullString{}
	if requestID := strings.TrimSpace(input.RequestID); requestID != "" {
		dedupeKey = sql.NullString{
			String: fmt.Sprintf("request:%s:%s:%s:%d", requestID, input.Name, input.SubjectType, input.SubjectID),
			Valid:  true,
		}
	}
	properties := make(map[string]any)
	for key, value := range input.Properties {
		if _, ok := allowedPropertyKeys[key]; ok {
			properties[key] = value
		}
	}
	occurredAt := input.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = s.now()
	}

	return s.store.RecordProductEvent(ctx, nil, store.ProductEventParams{
		EventName:   input.Name,
		ActorUserID: actorID,
		SubjectType: input.SubjectType,
		SubjectID:   subjectID,
		RequestID:   strings.TrimSpace(input.RequestID),
		DedupeKey:   dedupeKey,
		Properties:  properties,
		OccurredAt:  occurredAt,
	})
}

func (s *Service) Dashboard(ctx context.Context, adminEmail string, days int) (Dashboard, error) {
	if !s.isAdmin(adminEmail) {
		return Dashboard{}, httpapi.NewError(
			http.StatusForbidden,
			httpapi.CodeForbidden,
			"仅管理员可以查看产品指标",
		)
	}
	if s.store == nil {
		return Dashboard{}, httpapi.FeatureNotReady("product analytics")
	}
	if days != 7 && days != 28 && days != 90 {
		return Dashboard{}, httpapi.NewError(
			http.StatusBadRequest,
			httpapi.CodeInvalidParams,
			"统计周期仅支持 7、28 或 90 天",
		)
	}

	to := s.now()
	from := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC).
		AddDate(0, 0, -(days - 1))
	snapshot, err := s.store.GetProductMetrics(ctx, nil, from, to.Add(time.Nanosecond))
	if err != nil {
		return Dashboard{}, httpapi.InternalError(err)
	}

	eventCounts := make(map[string]store.ProductEventMetric, len(snapshot.EventMetrics))
	events := make([]EventMetric, 0, len(snapshot.EventMetrics))
	for _, metric := range snapshot.EventMetrics {
		eventCounts[metric.EventName] = metric
		events = append(events, EventMetric{
			EventName:    metric.EventName,
			Count:        metric.Count,
			UniqueActors: metric.UniqueActors,
		})
	}

	funnelDefinitions := []struct {
		key   string
		label string
	}{
		{EventIntentCreated, "创建意图"},
		{EventRecommendationsViewed, "查看推荐"},
		{EventInterestExpressed, "表达兴趣"},
		{EventMatchCreated, "双向匹配"},
		{EventMeetupConfirmed, "确认约见"},
		{EventMeetupCompleted, "完成见面"},
	}
	funnel := make([]FunnelStep, 0, len(funnelDefinitions))
	var previousCount int64
	for index, definition := range funnelDefinitions {
		count := eventCounts[definition.key].UniqueActors
		conversion := float64(0)
		if index == 0 {
			conversion = 100
		} else if previousCount > 0 {
			conversion = percentage(count, previousCount)
		}
		funnel = append(funnel, FunnelStep{
			Key:                    definition.key,
			Label:                  definition.label,
			Count:                  count,
			ConversionFromPrevious: conversion,
		})
		previousCount = count
	}

	dailyByDate := make(map[string]store.DailyProductMetric, len(snapshot.DailyMetrics))
	for _, metric := range snapshot.DailyMetrics {
		dailyByDate[metric.Date] = metric
	}
	daily := make([]DailyMetric, 0, days)
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		metric := dailyByDate[date]
		daily = append(daily, DailyMetric{
			Date:                  date,
			ActiveUsers:           metric.ActiveUsers,
			IntentsCreated:        metric.IntentsCreated,
			RecommendationsViewed: metric.RecommendationsViewed,
			InterestsExpressed:    metric.InterestsExpressed,
			MatchesCreated:        metric.MatchesCreated,
			MeetupsConfirmed:      metric.MeetupsConfirmed,
			MeetupsCompleted:      metric.MeetupsCompleted,
			SafetyReports:         metric.SafetyReports,
		})
	}

	matches := eventCounts[EventMatchCreated].Count
	meetupsConfirmed := eventCounts[EventMeetupConfirmed].Count
	meetupsCompleted := eventCounts[EventMeetupCompleted].Count
	reports := eventCounts[EventSafetyReportSubmitted].Count

	return Dashboard{
		Period: Period{Days: days, From: from, To: to},
		NorthStar: NorthStarMetrics{
			MeetupsCompleted:  meetupsCompleted,
			MeetupsConfirmed:  meetupsConfirmed,
			MatchToMeetupRate: percentage(meetupsCompleted, matches),
			RepeatIntentRate:  percentage(snapshot.RepeatIntentUsers, snapshot.IntentCreators),
		},
		Activity: ActivityMetrics{
			RegisteredUsers:   snapshot.RegisteredUsers,
			ActiveUsers:       snapshot.ActiveUsers,
			IntentCreators:    snapshot.IntentCreators,
			RepeatIntentUsers: snapshot.RepeatIntentUsers,
		},
		Safety: SafetyMetrics{
			OpenReports: snapshot.OpenSafetyReports,
			ReportRate:  percentage(reports, snapshot.ActiveUsers),
		},
		Funnel:      funnel,
		Events:      events,
		Daily:       daily,
		GeneratedAt: to,
	}, nil
}

func (s *Service) isAdmin(email string) bool {
	if s == nil {
		return false
	}
	_, ok := s.adminEmails[normalizeEmail(email)]
	return ok
}

func (s *Service) now() time.Time {
	if s == nil || s.nowFunc == nil {
		return time.Now().UTC()
	}
	return s.nowFunc().UTC()
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func percentage(numerator int64, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator) * 100
}
