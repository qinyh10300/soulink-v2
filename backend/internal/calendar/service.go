package calendar

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	maxEventTitleLength       = 120
	maxEventDescriptionLength = 2000
	maxEventLocationLength    = 160
	maxEventAttendees         = 20
	maxEventDuration          = 7 * 24 * time.Hour
	maxCalendarRange          = 93 * 24 * time.Hour
)

var validEventTypes = map[string]struct{}{
	"meetup": {}, "study": {}, "sports": {}, "meal": {},
	"date": {}, "project": {}, "other": {},
}

var validVisibilities = map[string]struct{}{
	"public": {}, "busy_only": {}, "private": {},
}

type Service struct {
	store *store.Store
}

type EventInput struct {
	EventType   string
	Title       string
	Description string
	Location    string
	StartsAt    time.Time
	EndsAt      time.Time
	Visibility  string
	InviteeIDs  []int64
}

type AttendeeView struct {
	UserID     int64  `json:"userId"`
	Nickname   string `json:"nickname"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
	Status     string `json:"status"`
}

type EventView struct {
	ID                  int64          `json:"id"`
	OrganizerID         int64          `json:"organizerId,omitempty"`
	OrganizerNickname   string         `json:"organizerNickname,omitempty"`
	OrganizerDepartment string         `json:"organizerDepartment,omitempty"`
	EventType           string         `json:"eventType"`
	Title               string         `json:"title"`
	Description         string         `json:"description"`
	Location            string         `json:"location"`
	StartsAt            string         `json:"startsAt"`
	EndsAt              string         `json:"endsAt"`
	Visibility          string         `json:"visibility"`
	IsOrganizer         bool           `json:"isOrganizer"`
	AttendeeStatus      string         `json:"attendeeStatus,omitempty"`
	IsBusyOnly          bool           `json:"isBusyOnly"`
	Attendees           []AttendeeView `json:"attendees"`
	CreatedAt           string         `json:"createdAt"`
}

func NewService(dataStore *store.Store) *Service {
	return &Service{store: dataStore}
}

func (s *Service) ListEvents(ctx context.Context, userID int64, from, to time.Time) ([]EventView, error) {
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	events, err := s.store.ListCalendarEvents(ctx, nil, userID, from, to)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]EventView, 0, len(events))
	for _, event := range events {
		attendees, err := s.store.ListCalendarAttendees(ctx, nil, event.ID)
		if err != nil {
			return nil, httpapi.InternalError(err)
		}
		views = append(views, fullEventView(event, attendees, userID))
	}
	return views, nil
}

func (s *Service) ListUserSchedule(ctx context.Context, viewerID, targetUserID int64, from, to time.Time) ([]EventView, error) {
	if targetUserID <= 0 {
		return nil, invalidParams("用户不存在")
	}
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	if _, err := s.store.GetUserByID(ctx, nil, targetUserID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "用户不存在")
		}
		return nil, httpapi.InternalError(err)
	}
	if viewerID != targetUserID {
		settings, err := s.store.GetUserProfileSettings(ctx, nil, targetUserID)
		if errors.Is(err, store.ErrNotFound) || (err == nil && !settings.SchedulePublic) {
			return nil, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "该用户尚未公开时间表")
		}
		if err != nil {
			return nil, httpapi.InternalError(err)
		}
	}
	events, err := s.store.ListUserCalendarAvailability(ctx, nil, viewerID, targetUserID, from, to)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]EventView, 0, len(events))
	for _, event := range events {
		canSeeDetails := viewerID == targetUserID || event.ViewerIsParticipant || event.Visibility == "public"
		if !canSeeDetails {
			views = append(views, busyEventView(event))
			continue
		}
		attendees, err := s.store.ListCalendarAttendees(ctx, nil, event.ID)
		if err != nil {
			return nil, httpapi.InternalError(err)
		}
		view := fullEventView(event, attendees, viewerID)
		if viewerID != targetUserID && !event.ViewerIsParticipant {
			view.Attendees = []AttendeeView{}
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *Service) CreateEvent(ctx context.Context, organizerID int64, input EventInput) (EventView, error) {
	normalized, err := normalizeEventInput(input, true)
	if err != nil {
		return EventView{}, err
	}
	invitees, err := s.normalizeInvitees(ctx, organizerID, normalized.InviteeIDs)
	if err != nil {
		return EventView{}, err
	}

	var event store.CalendarEvent
	now := time.Now()
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		created, err := s.store.CreateCalendarEvent(ctx, tx, store.CreateCalendarEventParams{
			OrganizerID: organizerID,
			EventType:   normalized.EventType,
			Title:       normalized.Title,
			Description: normalized.Description,
			Location:    normalized.Location,
			StartsAt:    normalized.StartsAt,
			EndsAt:      normalized.EndsAt,
			Visibility:  normalized.Visibility,
			Now:         now,
		})
		if err != nil {
			return httpapi.InternalError(err)
		}
		event = created
		for _, inviteeID := range invitees {
			if err := s.store.AddCalendarAttendee(ctx, tx, event.ID, inviteeID, now); err != nil {
				return httpapi.InternalError(err)
			}
		}
		return nil
	})
	if err != nil {
		return EventView{}, err
	}
	attendees, err := s.store.ListCalendarAttendees(ctx, nil, event.ID)
	if err != nil {
		return EventView{}, httpapi.InternalError(err)
	}
	return fullEventView(event, attendees, organizerID), nil
}

func (s *Service) UpdateEvent(ctx context.Context, userID, eventID int64, input EventInput) (EventView, error) {
	event, err := s.requireEvent(ctx, userID, eventID)
	if err != nil {
		return EventView{}, err
	}
	if event.OrganizerID != userID {
		return EventView{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "仅发起人可以修改日程")
	}
	normalized, err := normalizeEventInput(input, false)
	if err != nil {
		return EventView{}, err
	}
	updated, err := s.store.UpdateCalendarEvent(ctx, nil, userID, store.UpdateCalendarEventParams{
		EventID:     eventID,
		EventType:   normalized.EventType,
		Title:       normalized.Title,
		Description: normalized.Description,
		Location:    normalized.Location,
		StartsAt:    normalized.StartsAt,
		EndsAt:      normalized.EndsAt,
		Visibility:  normalized.Visibility,
		Now:         time.Now(),
	})
	if err != nil {
		return EventView{}, httpapi.InternalError(err)
	}
	attendees, err := s.store.ListCalendarAttendees(ctx, nil, eventID)
	if err != nil {
		return EventView{}, httpapi.InternalError(err)
	}
	return fullEventView(updated, attendees, userID), nil
}

func (s *Service) CancelEvent(ctx context.Context, userID, eventID int64) error {
	event, err := s.requireEvent(ctx, userID, eventID)
	if err != nil {
		return err
	}
	if event.OrganizerID != userID {
		return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "仅发起人可以取消日程")
	}
	if err := s.store.CancelCalendarEvent(ctx, nil, eventID, time.Now()); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) Respond(ctx context.Context, userID, eventID int64, status string) (EventView, error) {
	if status != "accepted" && status != "declined" {
		return EventView{}, invalidParams("请选择接受或拒绝")
	}
	event, err := s.requireEvent(ctx, userID, eventID)
	if err != nil {
		return EventView{}, err
	}
	if event.OrganizerID == userID || !event.ViewerAttendeeStatus.Valid {
		return EventView{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无法回复该邀请")
	}
	if err := s.store.UpdateCalendarAttendeeStatus(ctx, nil, eventID, userID, status, time.Now()); err != nil {
		return EventView{}, httpapi.InternalError(err)
	}
	updated, err := s.store.GetCalendarEvent(ctx, nil, userID, eventID)
	if err != nil {
		return EventView{}, httpapi.InternalError(err)
	}
	attendees, err := s.store.ListCalendarAttendees(ctx, nil, eventID)
	if err != nil {
		return EventView{}, httpapi.InternalError(err)
	}
	return fullEventView(updated, attendees, userID), nil
}

func (s *Service) requireEvent(ctx context.Context, viewerID, eventID int64) (store.CalendarEvent, error) {
	event, err := s.store.GetCalendarEvent(ctx, nil, viewerID, eventID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.CalendarEvent{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "日程不存在")
		}
		return store.CalendarEvent{}, httpapi.InternalError(err)
	}
	if event.OrganizerID != viewerID && !event.ViewerAttendeeStatus.Valid {
		return store.CalendarEvent{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权访问该日程")
	}
	return event, nil
}

func (s *Service) normalizeInvitees(ctx context.Context, organizerID int64, raw []int64) ([]int64, error) {
	seen := make(map[int64]struct{}, len(raw))
	invitees := make([]int64, 0, len(raw))
	for _, inviteeID := range raw {
		if inviteeID <= 0 {
			return nil, invalidParams("邀请对象不存在")
		}
		if inviteeID == organizerID {
			continue
		}
		if _, ok := seen[inviteeID]; ok {
			continue
		}
		if _, err := s.store.GetUserByID(ctx, nil, inviteeID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return nil, invalidParams("邀请对象不存在")
			}
			return nil, httpapi.InternalError(err)
		}
		seen[inviteeID] = struct{}{}
		invitees = append(invitees, inviteeID)
	}
	if len(invitees) > maxEventAttendees {
		return nil, invalidParams("单个日程最多邀请 20 人")
	}
	sort.Slice(invitees, func(i, j int) bool { return invitees[i] < invitees[j] })
	return invitees, nil
}

func normalizeEventInput(input EventInput, creating bool) (EventInput, error) {
	input.EventType = strings.TrimSpace(input.EventType)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Location = strings.TrimSpace(input.Location)
	input.Visibility = strings.TrimSpace(input.Visibility)
	if _, ok := validEventTypes[input.EventType]; !ok {
		return EventInput{}, invalidParams("日程类型不正确")
	}
	if _, ok := validVisibilities[input.Visibility]; !ok {
		return EventInput{}, invalidParams("可见范围不正确")
	}
	if utf8.RuneCountInString(input.Title) == 0 || utf8.RuneCountInString(input.Title) > maxEventTitleLength {
		return EventInput{}, invalidParams("日程标题需为 1 到 120 个字符")
	}
	if utf8.RuneCountInString(input.Description) > maxEventDescriptionLength || utf8.RuneCountInString(input.Location) > maxEventLocationLength {
		return EventInput{}, invalidParams("日程内容过长")
	}
	if input.StartsAt.IsZero() || input.EndsAt.IsZero() || !input.EndsAt.After(input.StartsAt) {
		return EventInput{}, invalidParams("日程时间不正确")
	}
	if input.EndsAt.Sub(input.StartsAt) > maxEventDuration {
		return EventInput{}, invalidParams("单个日程不能超过 7 天")
	}
	if creating && input.StartsAt.Before(time.Now().Add(-5*time.Minute)) {
		return EventInput{}, invalidParams("不能创建已经开始的日程")
	}
	return input, nil
}

func validateRange(from, to time.Time) error {
	if from.IsZero() || to.IsZero() || !to.After(from) || to.Sub(from) > maxCalendarRange {
		return invalidParams("日历时间范围不正确")
	}
	return nil
}

func fullEventView(event store.CalendarEvent, attendees []store.CalendarAttendee, viewerID int64) EventView {
	attendeeViews := make([]AttendeeView, 0, len(attendees))
	for _, attendee := range attendees {
		attendeeViews = append(attendeeViews, AttendeeView{
			UserID:     attendee.UserID,
			Nickname:   attendee.Nickname,
			Department: attendee.Department,
			Grade:      attendee.Grade,
			Status:     attendee.Status,
		})
	}
	view := EventView{
		ID:                  event.ID,
		OrganizerID:         event.OrganizerID,
		OrganizerNickname:   event.OrganizerNickname,
		OrganizerDepartment: event.OrganizerDepartment,
		EventType:           event.EventType,
		Title:               event.Title,
		Description:         event.Description,
		Location:            event.Location,
		StartsAt:            event.StartsAt.Format(time.RFC3339),
		EndsAt:              event.EndsAt.Format(time.RFC3339),
		Visibility:          event.Visibility,
		IsOrganizer:         event.OrganizerID == viewerID,
		Attendees:           attendeeViews,
		CreatedAt:           event.CreatedAt.Format(time.RFC3339),
	}
	if event.ViewerAttendeeStatus.Valid {
		view.AttendeeStatus = event.ViewerAttendeeStatus.String
	}
	return view
}

func busyEventView(event store.CalendarEvent) EventView {
	return EventView{
		ID:         event.ID,
		EventType:  "busy",
		Title:      "忙碌",
		StartsAt:   event.StartsAt.Format(time.RFC3339),
		EndsAt:     event.EndsAt.Format(time.RFC3339),
		Visibility: event.Visibility,
		IsBusyOnly: true,
		Attendees:  []AttendeeView{},
		CreatedAt:  event.CreatedAt.Format(time.RFC3339),
	}
}

func invalidParams(message string) error {
	return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, message)
}
