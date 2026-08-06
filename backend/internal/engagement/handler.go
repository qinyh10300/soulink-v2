package engagement

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type Handler struct {
	service *Service
}

type notificationPreferencesRequest struct {
	NotificationsEnabled bool   `json:"notificationsEnabled"`
	NewMessages          bool   `json:"newMessages"`
	MeetupInvites        bool   `json:"meetupInvites"`
	MeetupReminders      bool   `json:"meetupReminders"`
	DailyPicks           bool   `json:"dailyPicks"`
	PermissionState      string `json:"permissionState"`
	Timezone             string `json:"timezone"`
}

type onboardingRequest struct {
	SelectedScenes []string `json:"selectedScenes"`
}

type campusThemesRequest struct {
	ThemeKeys []string `json:"themeKeys"`
}

type updateCampusThemeRequest struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	DefaultPrompt string `json:"defaultPrompt"`
	ScheduleHint  string `json:"scheduleHint"`
	IconName      string `json:"iconName"`
	IsActive      bool   `json:"isActive"`
	DisplayOrder  int    `json:"displayOrder"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	preferences, err := h.service.GetPreferences(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, preferences)
}

func (h *Handler) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	var request notificationPreferencesRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	preferences, err := h.service.UpdateNotificationPreferences(
		r.Context(),
		user.ID,
		NotificationPreferencesInput{
			NotificationsEnabled: request.NotificationsEnabled,
			NewMessages:          request.NewMessages,
			MeetupInvites:        request.MeetupInvites,
			MeetupReminders:      request.MeetupReminders,
			DailyPicks:           request.DailyPicks,
			PermissionState:      request.PermissionState,
			Timezone:             request.Timezone,
		},
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, preferences)
}

func (h *Handler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	var request onboardingRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	preferences, err := h.service.CompleteOnboarding(
		r.Context(), user.ID, OnboardingInput{SelectedScenes: request.SelectedScenes},
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, preferences)
}

func (h *Handler) ListThemes(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	themes, err := h.service.ListThemes(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, themes)
}

func (h *Handler) ReplaceThemes(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	var request campusThemesRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	themes, err := h.service.ReplaceThemes(r.Context(), user.ID, request.ThemeKeys)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, themes)
}

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	beforeID, err := parseOptionalInt64(r.URL.Query().Get("before"))
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	limit := 20
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		limit, err = strconv.Atoi(rawLimit)
		if err != nil {
			httpapi.Error(w, invalidParams("通知分页参数不正确"))
			return
		}
	}
	page, err := h.service.ListNotifications(r.Context(), user.ID, beforeID, limit)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, page)
}

func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	notificationID, err := parseRouteID(r, "notificationId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	if err := h.service.MarkNotificationRead(r.Context(), user.ID, notificationID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}

func (h *Handler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	if err := h.service.MarkAllNotificationsRead(r.Context(), user.ID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}

func (h *Handler) MarkNotificationDelivered(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	notificationID, err := parseRouteID(r, "notificationId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	if err := h.service.MarkNotificationDelivered(r.Context(), user.ID, notificationID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}

func (h *Handler) ThemeMetrics(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	days := 28
	if rawDays := strings.TrimSpace(r.URL.Query().Get("days")); rawDays != "" {
		parsed, err := strconv.Atoi(rawDays)
		if err != nil {
			httpapi.Error(w, invalidParams("统计周期格式错误"))
			return
		}
		days = parsed
	}
	metrics, err := h.service.ThemeMetrics(r.Context(), user.Email, days)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, metrics)
}

func (h *Handler) ListAdminThemes(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	themes, err := h.service.ListAdminThemes(r.Context(), user.Email, user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, themes)
}

func (h *Handler) UpdateTheme(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	var request updateCampusThemeRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	theme, err := h.service.UpdateTheme(
		r.Context(), user.Email, chi.URLParam(r, "themeKey"),
		UpdateCampusThemeInput{
			Title: request.Title, Description: request.Description,
			DefaultPrompt: request.DefaultPrompt, ScheduleHint: request.ScheduleHint,
			IconName: request.IconName, IsActive: request.IsActive,
			DisplayOrder: request.DisplayOrder,
		},
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, theme)
}

func currentUser(w http.ResponseWriter, r *http.Request) (store.User, bool) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(
			http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录",
		))
	}
	return user, ok
}

func parseOptionalInt64(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, invalidParams("通知游标不正确")
	}
	return value, nil
}

func parseRouteID(r *http.Request, key string) (int64, error) {
	value, err := strconv.ParseInt(chi.URLParam(r, key), 10, 64)
	if err != nil || value <= 0 {
		return 0, invalidParams("通知编号不正确")
	}
	return value, nil
}
