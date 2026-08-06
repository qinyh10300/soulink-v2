package engagement

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type DataStore interface {
	WithTx(context.Context, func(*sql.Tx) error) error
	GetEngagementPreferences(context.Context, *sql.Tx, int64) (store.EngagementPreferences, error)
	UpsertEngagementPreferences(context.Context, *sql.Tx, store.UpsertEngagementPreferencesParams) (store.EngagementPreferences, error)
	CompleteUserOnboarding(context.Context, *sql.Tx, int64, []string, bool, time.Time) (store.EngagementPreferences, error)
	ListCampusThemes(context.Context, *sql.Tx, int64, bool) ([]store.CampusTheme, error)
	ReplaceUserCampusThemes(context.Context, *sql.Tx, int64, []string, time.Time) error
	UpdateCampusTheme(context.Context, *sql.Tx, store.UpdateCampusThemeParams) (store.CampusTheme, error)
	MaterializeUserNotifications(context.Context, *sql.Tx, int64, string, time.Time) error
	ListUserNotifications(context.Context, *sql.Tx, int64, int64, int) ([]store.UserNotification, error)
	CountUnreadUserNotifications(context.Context, *sql.Tx, int64) (int64, error)
	MarkUserNotificationRead(context.Context, *sql.Tx, int64, int64, time.Time) error
	MarkAllUserNotificationsRead(context.Context, *sql.Tx, int64, time.Time) error
	MarkUserNotificationDelivered(context.Context, *sql.Tx, int64, int64, time.Time) error
	GetCampusThemeMetrics(context.Context, *sql.Tx, time.Time, time.Time) ([]store.CampusThemeMetric, error)
}

type TaskObserver interface {
	ObserveTask(context.Context, string, time.Duration, error)
}

type Service struct {
	store       DataStore
	adminEmails map[string]struct{}
	nowFunc     func() time.Time
	tasks       TaskObserver
}

var supportedScenes = map[string]struct{}{
	"meal": {}, "study": {}, "sport": {}, "romantic": {}, "project": {},
}

var supportedPermissionStates = map[string]struct{}{
	"default": {}, "granted": {}, "denied": {}, "unavailable": {},
}

var supportedThemeIcons = map[string]struct{}{
	"utensils": {}, "book-open": {}, "dumbbell": {}, "users-round": {},
	"sparkles": {}, "heart": {}, "calendar": {},
}

func NewService(dataStore DataStore, adminEmails []string) *Service {
	admins := make(map[string]struct{}, len(adminEmails))
	for _, email := range adminEmails {
		if normalized := normalizeEmail(email); normalized != "" {
			admins[normalized] = struct{}{}
		}
	}
	return &Service{store: dataStore, adminEmails: admins, nowFunc: time.Now}
}

func (s *Service) WithTaskObserver(observer TaskObserver) *Service {
	s.tasks = observer
	return s
}

func (s *Service) GetPreferences(ctx context.Context, userID int64) (PreferencesView, error) {
	preferences, err := s.store.GetEngagementPreferences(ctx, nil, userID)
	if errors.Is(err, store.ErrNotFound) {
		return defaultPreferences(), nil
	}
	if err != nil {
		return PreferencesView{}, httpapi.InternalError(err)
	}
	return toPreferencesView(preferences), nil
}

func (s *Service) UpdateNotificationPreferences(
	ctx context.Context,
	userID int64,
	input NotificationPreferencesInput,
) (PreferencesView, error) {
	input.PermissionState = strings.ToLower(strings.TrimSpace(input.PermissionState))
	input.Timezone = strings.TrimSpace(input.Timezone)
	if _, ok := supportedPermissionStates[input.PermissionState]; !ok {
		return PreferencesView{}, invalidParams("通知授权状态不正确")
	}
	if input.Timezone == "" {
		input.Timezone = defaultTimezone
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return PreferencesView{}, invalidParams("时区不正确")
	}

	preferences, err := s.store.UpsertEngagementPreferences(
		ctx,
		nil,
		store.UpsertEngagementPreferencesParams{
			UserID:               userID,
			NotificationsEnabled: input.NotificationsEnabled,
			NewMessages:          input.NewMessages,
			MeetupInvites:        input.MeetupInvites,
			MeetupReminders:      input.MeetupReminders,
			DailyPicks:           input.DailyPicks,
			PermissionState:      input.PermissionState,
			Timezone:             input.Timezone,
			Now:                  s.now(),
		},
	)
	if err != nil {
		return PreferencesView{}, httpapi.InternalError(err)
	}
	return toPreferencesView(preferences), nil
}

func (s *Service) CompleteOnboarding(
	ctx context.Context,
	userID int64,
	input OnboardingInput,
) (PreferencesView, error) {
	scenes, err := normalizeUniqueList(input.SelectedScenes, supportedScenes, 1, 5)
	if err != nil {
		return PreferencesView{}, err
	}
	romanticOptIn := contains(scenes, "romantic")
	var preferences store.EngagementPreferences
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		var completeErr error
		preferences, completeErr = s.store.CompleteUserOnboarding(
			ctx, tx, userID, scenes, romanticOptIn, s.now(),
		)
		return completeErr
	})
	if err != nil {
		return PreferencesView{}, httpapi.InternalError(err)
	}
	return toPreferencesView(preferences), nil
}

func (s *Service) ListThemes(ctx context.Context, userID int64) ([]CampusThemeView, error) {
	themes, err := s.store.ListCampusThemes(ctx, nil, userID, false)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return toCampusThemeViews(themes), nil
}

func (s *Service) ListAdminThemes(
	ctx context.Context,
	adminEmail string,
	userID int64,
) ([]CampusThemeView, error) {
	if !s.isAdmin(adminEmail) {
		return nil, forbidden("仅管理员可以配置校园主题")
	}
	themes, err := s.store.ListCampusThemes(ctx, nil, userID, true)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return toCampusThemeViews(themes), nil
}

func (s *Service) ReplaceThemes(
	ctx context.Context,
	userID int64,
	themeKeys []string,
) ([]CampusThemeView, error) {
	keys, err := normalizeUniqueList(themeKeys, nil, 0, 8)
	if err != nil {
		return nil, err
	}
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		return s.store.ReplaceUserCampusThemes(ctx, tx, userID, keys, s.now())
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, invalidParams("校园主题不存在或已停用")
	}
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return s.ListThemes(ctx, userID)
}

func (s *Service) ListNotifications(
	ctx context.Context,
	userID int64,
	beforeID int64,
	limit int,
) (NotificationPage, error) {
	if beforeID < 0 {
		return NotificationPage{}, invalidParams("通知游标不正确")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		return NotificationPage{}, invalidParams("每页通知不能超过 50 条")
	}

	preferences, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return NotificationPage{}, err
	}
	if preferences.NotificationsEnabled {
		location, locationErr := time.LoadLocation(preferences.Timezone)
		if locationErr != nil {
			location, _ = time.LoadLocation(defaultTimezone)
		}
		now := s.now()
		startedAt := time.Now()
		materializeErr := s.store.MaterializeUserNotifications(
			ctx, nil, userID, now.In(location).Format("2006-01-02"), now,
		)
		if s.tasks != nil {
			s.tasks.ObserveTask(ctx, "notifications.materialize", time.Since(startedAt), materializeErr)
		}
		if materializeErr != nil {
			return NotificationPage{}, httpapi.InternalError(materializeErr)
		}
	}

	items, err := s.store.ListUserNotifications(ctx, nil, userID, beforeID, limit+1)
	if err != nil {
		return NotificationPage{}, httpapi.InternalError(err)
	}
	unreadCount, err := s.store.CountUnreadUserNotifications(ctx, nil, userID)
	if err != nil {
		return NotificationPage{}, httpapi.InternalError(err)
	}

	var nextCursor *int64
	if len(items) > limit {
		items = items[:limit]
		cursor := items[len(items)-1].ID
		nextCursor = &cursor
	}
	views := make([]NotificationView, 0, len(items))
	for _, item := range items {
		views = append(views, toNotificationView(item))
	}
	return NotificationPage{Items: views, UnreadCount: unreadCount, NextCursor: nextCursor}, nil
}

func (s *Service) MarkNotificationRead(
	ctx context.Context,
	userID int64,
	notificationID int64,
) error {
	if notificationID <= 0 {
		return invalidParams("通知编号不正确")
	}
	err := s.store.MarkUserNotificationRead(ctx, nil, userID, notificationID, s.now())
	return mapNotificationMutationError(err)
}

func (s *Service) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	if err := s.store.MarkAllUserNotificationsRead(ctx, nil, userID, s.now()); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) MarkNotificationDelivered(
	ctx context.Context,
	userID int64,
	notificationID int64,
) error {
	if notificationID <= 0 {
		return invalidParams("通知编号不正确")
	}
	err := s.store.MarkUserNotificationDelivered(ctx, nil, userID, notificationID, s.now())
	return mapNotificationMutationError(err)
}

func (s *Service) ThemeMetrics(
	ctx context.Context,
	adminEmail string,
	days int,
) (CampusThemeMetricsView, error) {
	if !s.isAdmin(adminEmail) {
		return CampusThemeMetricsView{}, forbidden("仅管理员可以查看校园主题指标")
	}
	if days != 7 && days != 28 && days != 90 {
		return CampusThemeMetricsView{}, invalidParams("统计周期仅支持 7、28 或 90 天")
	}
	to := s.now()
	from := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC).
		AddDate(0, 0, -(days - 1))
	records, err := s.store.GetCampusThemeMetrics(ctx, nil, from, to.Add(time.Nanosecond))
	if err != nil {
		return CampusThemeMetricsView{}, httpapi.InternalError(err)
	}
	metrics := make([]CampusThemeMetricView, 0, len(records))
	for _, record := range records {
		metrics = append(metrics, CampusThemeMetricView{
			Key:                  record.Key,
			Title:                record.Title,
			Category:             record.Category,
			SupplyUsers:          record.SupplyUsers,
			DemandUsers:          record.DemandUsers,
			GroupIntents:         record.GroupIntents,
			GroupsFormed:         record.GroupsFormed,
			GroupFormationRate:   percentage(record.GroupsFormed, record.GroupIntents),
			MeetupsEligible:      record.MeetupsEligible,
			MeetupsCompleted:     record.MeetupsCompleted,
			MeetupCompletionRate: percentage(record.MeetupsCompleted, record.MeetupsEligible),
		})
	}
	return CampusThemeMetricsView{
		Days: days, From: from, To: to, Themes: metrics, GeneratedAt: to,
	}, nil
}

func (s *Service) UpdateTheme(
	ctx context.Context,
	adminEmail string,
	key string,
	input UpdateCampusThemeInput,
) (CampusThemeView, error) {
	if !s.isAdmin(adminEmail) {
		return CampusThemeView{}, forbidden("仅管理员可以配置校园主题")
	}
	key = strings.TrimSpace(key)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.DefaultPrompt = strings.TrimSpace(input.DefaultPrompt)
	input.ScheduleHint = strings.TrimSpace(input.ScheduleHint)
	input.IconName = strings.TrimSpace(input.IconName)
	if key == "" || utf8.RuneCountInString(input.Title) < 2 || utf8.RuneCountInString(input.Title) > 48 {
		return CampusThemeView{}, invalidParams("主题名称需为 2 到 48 个字符")
	}
	if utf8.RuneCountInString(input.Description) < 4 || utf8.RuneCountInString(input.Description) > 240 {
		return CampusThemeView{}, invalidParams("主题说明需为 4 到 240 个字符")
	}
	if utf8.RuneCountInString(input.DefaultPrompt) < 8 || utf8.RuneCountInString(input.DefaultPrompt) > 500 {
		return CampusThemeView{}, invalidParams("主题默认意图需为 8 到 500 个字符")
	}
	if utf8.RuneCountInString(input.ScheduleHint) > 80 {
		return CampusThemeView{}, invalidParams("主题时间提示不能超过 80 个字符")
	}
	if _, ok := supportedThemeIcons[input.IconName]; !ok {
		return CampusThemeView{}, invalidParams("主题图标不正确")
	}
	if input.DisplayOrder < 0 || input.DisplayOrder > 1000 {
		return CampusThemeView{}, invalidParams("主题排序不正确")
	}

	theme, err := s.store.UpdateCampusTheme(ctx, nil, store.UpdateCampusThemeParams{
		Key: key, Title: input.Title, Description: input.Description,
		DefaultPrompt: input.DefaultPrompt, ScheduleHint: input.ScheduleHint,
		IconName: input.IconName, IsActive: input.IsActive,
		DisplayOrder: input.DisplayOrder, Now: s.now(),
	})
	if errors.Is(err, store.ErrNotFound) {
		return CampusThemeView{}, httpapi.NewError(
			http.StatusNotFound, httpapi.CodeUserNotFound, "校园主题不存在",
		)
	}
	if err != nil {
		return CampusThemeView{}, httpapi.InternalError(err)
	}
	return toCampusThemeView(theme), nil
}

func defaultPreferences() PreferencesView {
	return PreferencesView{
		NotificationsEnabled: false,
		NewMessages:          true,
		MeetupInvites:        true,
		MeetupReminders:      true,
		DailyPicks:           true,
		PermissionState:      "default",
		Timezone:             defaultTimezone,
		SelectedScenes:       []string{},
		OnboardingCompleted:  false,
	}
}

func toPreferencesView(preferences store.EngagementPreferences) PreferencesView {
	view := PreferencesView{
		NotificationsEnabled: preferences.NotificationsEnabled,
		NewMessages:          preferences.NewMessages,
		MeetupInvites:        preferences.MeetupInvites,
		MeetupReminders:      preferences.MeetupReminders,
		DailyPicks:           preferences.DailyPicks,
		PermissionState:      preferences.PermissionState,
		Timezone:             preferences.Timezone,
		SelectedScenes:       append([]string(nil), preferences.SelectedScenes...),
		OnboardingCompleted:  preferences.OnboardingCompletedAt.Valid,
	}
	if preferences.OnboardingCompletedAt.Valid {
		value := preferences.OnboardingCompletedAt.Time.UTC().Format(time.RFC3339)
		view.OnboardingCompletedAt = &value
	}
	if !preferences.UpdatedAt.IsZero() {
		value := preferences.UpdatedAt.UTC().Format(time.RFC3339)
		view.UpdatedAt = &value
	}
	return view
}

func toCampusThemeViews(themes []store.CampusTheme) []CampusThemeView {
	views := make([]CampusThemeView, 0, len(themes))
	for _, theme := range themes {
		views = append(views, toCampusThemeView(theme))
	}
	return views
}

func toCampusThemeView(theme store.CampusTheme) CampusThemeView {
	return CampusThemeView{
		Key: theme.Key, Title: theme.Title, Description: theme.Description,
		Category: theme.Category, DefaultPrompt: theme.DefaultPrompt,
		ScheduleHint: theme.ScheduleHint, IconName: theme.IconName,
		IsActive: theme.IsActive, DisplayOrder: theme.DisplayOrder,
		Selected: theme.Selected, UpdatedAt: theme.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toNotificationView(item store.UserNotification) NotificationView {
	view := NotificationView{
		ID: item.ID, Kind: item.Kind, Title: item.Title, Body: item.Body,
		ActionPath:   item.ActionPath,
		ScheduledFor: item.ScheduledFor.UTC().Format(time.RFC3339),
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
	}
	if item.DeliveredAt.Valid {
		value := item.DeliveredAt.Time.UTC().Format(time.RFC3339)
		view.DeliveredAt = &value
	}
	if item.ReadAt.Valid {
		value := item.ReadAt.Time.UTC().Format(time.RFC3339)
		view.ReadAt = &value
	}
	return view
}

func normalizeUniqueList(
	values []string,
	allowed map[string]struct{},
	minimum int,
	maximum int,
) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if allowed != nil {
			if _, ok := allowed[value]; !ok {
				return nil, invalidParams("包含不支持的场景")
			}
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	if len(normalized) < minimum || len(normalized) > maximum {
		return nil, invalidParams("请选择合适数量的场景")
	}
	return normalized, nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func mapNotificationMutationError(err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "通知不存在")
	}
	if err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func invalidParams(message string) error {
	return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, message)
}

func forbidden(message string) error {
	return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, message)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *Service) isAdmin(email string) bool {
	_, ok := s.adminEmails[normalizeEmail(email)]
	return ok
}

func (s *Service) now() time.Time {
	if s == nil || s.nowFunc == nil {
		return time.Now().UTC()
	}
	return s.nowFunc().UTC()
}

func percentage(numerator int64, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator) * 100
}
