package engagement

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"soulink-backend/internal/store"
)

type engagementStoreStub struct {
	preferences          store.EngagementPreferences
	preferencesErr       error
	upsertInput          store.UpsertEngagementPreferencesParams
	completedScenes      []string
	completedRomantic    bool
	themes               []store.CampusTheme
	materializedDates    []string
	materializedByDedupe map[string]store.UserNotification
	notificationItems    []store.UserNotification
	unreadCount          int64
	metrics              []store.CampusThemeMetric
}

func (s *engagementStoreStub) WithTx(_ context.Context, fn func(*sql.Tx) error) error {
	return fn(nil)
}

func (s *engagementStoreStub) GetEngagementPreferences(
	_ context.Context,
	_ *sql.Tx,
	_ int64,
) (store.EngagementPreferences, error) {
	return s.preferences, s.preferencesErr
}

func (s *engagementStoreStub) UpsertEngagementPreferences(
	_ context.Context,
	_ *sql.Tx,
	params store.UpsertEngagementPreferencesParams,
) (store.EngagementPreferences, error) {
	s.upsertInput = params
	result := s.preferences
	result.UserID = params.UserID
	result.NotificationsEnabled = params.NotificationsEnabled
	result.NewMessages = params.NewMessages
	result.MeetupInvites = params.MeetupInvites
	result.MeetupReminders = params.MeetupReminders
	result.DailyPicks = params.DailyPicks
	result.PermissionState = params.PermissionState
	result.Timezone = params.Timezone
	result.UpdatedAt = params.Now
	return result, nil
}

func (s *engagementStoreStub) CompleteUserOnboarding(
	_ context.Context,
	_ *sql.Tx,
	userID int64,
	scenes []string,
	romantic bool,
	now time.Time,
) (store.EngagementPreferences, error) {
	s.completedScenes = append([]string(nil), scenes...)
	s.completedRomantic = romantic
	result := s.preferences
	result.UserID = userID
	result.SelectedScenes = append([]string(nil), scenes...)
	result.OnboardingCompletedAt = sql.NullTime{Time: now, Valid: true}
	result.UpdatedAt = now
	return result, nil
}

func (s *engagementStoreStub) ListCampusThemes(
	_ context.Context,
	_ *sql.Tx,
	_ int64,
	_ bool,
) ([]store.CampusTheme, error) {
	return s.themes, nil
}

func (s *engagementStoreStub) ReplaceUserCampusThemes(
	_ context.Context,
	_ *sql.Tx,
	_ int64,
	_ []string,
	_ time.Time,
) error {
	return nil
}

func (s *engagementStoreStub) UpdateCampusTheme(
	_ context.Context,
	_ *sql.Tx,
	params store.UpdateCampusThemeParams,
) (store.CampusTheme, error) {
	return store.CampusTheme{
		Key: params.Key, Title: params.Title, Description: params.Description,
		DefaultPrompt: params.DefaultPrompt, ScheduleHint: params.ScheduleHint,
		IconName: params.IconName, IsActive: params.IsActive,
		DisplayOrder: params.DisplayOrder, UpdatedAt: params.Now,
	}, nil
}

func (s *engagementStoreStub) MaterializeUserNotifications(
	_ context.Context,
	_ *sql.Tx,
	userID int64,
	localDate string,
	now time.Time,
) error {
	s.materializedDates = append(s.materializedDates, localDate)
	if s.materializedByDedupe == nil {
		s.materializedByDedupe = make(map[string]store.UserNotification)
	}
	dedupe := "daily-picks:" + localDate
	if _, exists := s.materializedByDedupe[dedupe]; !exists {
		s.materializedByDedupe[dedupe] = store.UserNotification{
			ID: 7, UserID: userID, Kind: "daily_picks", Title: "今日精选",
			Body: "打开 SouLink 查看", ActionPath: "/connect",
			ScheduledFor: now, CreatedAt: now,
		}
	}
	return nil
}

func (s *engagementStoreStub) ListUserNotifications(
	_ context.Context,
	_ *sql.Tx,
	_ int64,
	_ int64,
	_ int,
) ([]store.UserNotification, error) {
	if len(s.notificationItems) > 0 {
		return s.notificationItems, nil
	}
	items := make([]store.UserNotification, 0, len(s.materializedByDedupe))
	for _, item := range s.materializedByDedupe {
		items = append(items, item)
	}
	return items, nil
}

func (s *engagementStoreStub) CountUnreadUserNotifications(
	_ context.Context,
	_ *sql.Tx,
	_ int64,
) (int64, error) {
	return s.unreadCount, nil
}

func (s *engagementStoreStub) MarkUserNotificationRead(
	context.Context,
	*sql.Tx,
	int64,
	int64,
	time.Time,
) error {
	return nil
}

func (s *engagementStoreStub) MarkAllUserNotificationsRead(
	context.Context,
	*sql.Tx,
	int64,
	time.Time,
) error {
	return nil
}

func (s *engagementStoreStub) MarkUserNotificationDelivered(
	context.Context,
	*sql.Tx,
	int64,
	int64,
	time.Time,
) error {
	return nil
}

func (s *engagementStoreStub) GetCampusThemeMetrics(
	context.Context,
	*sql.Tx,
	time.Time,
	time.Time,
) ([]store.CampusThemeMetric, error) {
	return s.metrics, nil
}

func TestUpdateNotificationPreferencesRejectsInvalidTimezone(t *testing.T) {
	t.Parallel()

	dataStore := &engagementStoreStub{}
	service := NewService(dataStore, nil)
	_, err := service.UpdateNotificationPreferences(context.Background(), 1, NotificationPreferencesInput{
		PermissionState: "granted",
		Timezone:        "Mars/Tsinghua",
	})
	if err == nil {
		t.Fatal("UpdateNotificationPreferences() error = nil, want invalid timezone")
	}
	if dataStore.upsertInput.UserID != 0 {
		t.Fatal("invalid timezone reached persistence")
	}
}

func TestNotificationOptOutStopsMaterialization(t *testing.T) {
	t.Parallel()

	dataStore := &engagementStoreStub{preferences: store.EngagementPreferences{
		UserID: 1, NotificationsEnabled: false, NewMessages: true,
		MeetupInvites: true, MeetupReminders: true, DailyPicks: true,
		PermissionState: "denied", Timezone: defaultTimezone,
	}}
	service := NewService(dataStore, nil)
	page, err := service.ListNotifications(context.Background(), 1, 0, 20)
	if err != nil {
		t.Fatalf("ListNotifications() error = %v", err)
	}
	if len(dataStore.materializedDates) != 0 {
		t.Fatalf("materialization calls = %d, want 0 after opt-out", len(dataStore.materializedDates))
	}
	if len(page.Items) != 0 {
		t.Fatalf("items = %#v, want empty", page.Items)
	}
}

func TestCompleteOnboardingMakesRomanticOptInExplicit(t *testing.T) {
	t.Parallel()

	dataStore := &engagementStoreStub{preferences: store.EngagementPreferences{
		PermissionState: "default", Timezone: defaultTimezone,
	}}
	service := NewService(dataStore, nil)
	if _, err := service.CompleteOnboarding(
		context.Background(), 1, OnboardingInput{SelectedScenes: []string{"meal", "project"}},
	); err != nil {
		t.Fatalf("CompleteOnboarding() error = %v", err)
	}
	if dataStore.completedRomantic {
		t.Fatal("romantic opt-in = true without explicit romantic selection")
	}

	if _, err := service.CompleteOnboarding(
		context.Background(), 1, OnboardingInput{SelectedScenes: []string{"romantic"}},
	); err != nil {
		t.Fatalf("CompleteOnboarding(romantic) error = %v", err)
	}
	if !dataStore.completedRomantic {
		t.Fatal("romantic opt-in = false after explicit romantic selection")
	}
}

func TestRepeatedNotificationRefreshUsesLocalDateDedupe(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 8, 8, 1, 30, 0, 0, time.UTC)
	dataStore := &engagementStoreStub{preferences: store.EngagementPreferences{
		UserID: 1, NotificationsEnabled: true, NewMessages: true,
		MeetupInvites: true, MeetupReminders: true, DailyPicks: true,
		PermissionState: "granted", Timezone: "America/New_York",
	}}
	service := NewService(dataStore, nil)
	service.nowFunc = func() time.Time { return fixedNow }

	first, err := service.ListNotifications(context.Background(), 1, 0, 20)
	if err != nil {
		t.Fatalf("first ListNotifications() error = %v", err)
	}
	second, err := service.ListNotifications(context.Background(), 1, 0, 20)
	if err != nil {
		t.Fatalf("second ListNotifications() error = %v", err)
	}
	if len(dataStore.materializedDates) != 2 || dataStore.materializedDates[0] != "2026-08-07" {
		t.Fatalf("materialized dates = %#v, want local date 2026-08-07", dataStore.materializedDates)
	}
	if len(first.Items) != 1 || len(second.Items) != 1 {
		t.Fatalf("item counts = %d, %d, want deduped 1", len(first.Items), len(second.Items))
	}
}

func TestThemeMetricsRequiresAdminAndCalculatesRates(t *testing.T) {
	t.Parallel()

	dataStore := &engagementStoreStub{metrics: []store.CampusThemeMetric{{
		Key: "week-study", Title: "本周自习", Category: "study",
		SupplyUsers: 9, DemandUsers: 12, GroupIntents: 4, GroupsFormed: 3,
		MeetupsEligible: 3, MeetupsCompleted: 2,
	}}}
	service := NewService(dataStore, []string{"admin@mails.tsinghua.edu.cn"})
	if _, err := service.ThemeMetrics(context.Background(), "student@mails.tsinghua.edu.cn", 28); err == nil {
		t.Fatal("ThemeMetrics(non-admin) error = nil, want forbidden")
	}
	view, err := service.ThemeMetrics(context.Background(), "ADMIN@mails.tsinghua.edu.cn", 28)
	if err != nil {
		t.Fatalf("ThemeMetrics(admin) error = %v", err)
	}
	if len(view.Themes) != 1 || view.Themes[0].GroupFormationRate != 75 {
		t.Fatalf("theme metrics = %#v, want 75%% formation rate", view.Themes)
	}
	if view.Themes[0].MeetupCompletionRate < 66.6 || view.Themes[0].MeetupCompletionRate > 66.7 {
		t.Fatalf("meetup completion rate = %v, want about 66.7", view.Themes[0].MeetupCompletionRate)
	}
}
