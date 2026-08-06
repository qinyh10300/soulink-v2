package meetup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5/middleware"

	"soulink-backend/internal/analytics"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	maxAvailabilityWindows  = 24
	maxMeetupParticipants   = 20
	maxMeetupTitleLength    = 120
	maxMeetupLocationLength = 160
	maxFeedbackNotesLength  = 1000
	maxCancellationLength   = 240
	maxSuggestionRange      = 31 * 24 * time.Hour
	maxMeetupHorizon        = 180 * 24 * time.Hour
)

type Service struct {
	store    *store.Store
	recorder analytics.EventRecorder
	nowFunc  func() time.Time
}

func NewService(dataStore *store.Store, recorder analytics.EventRecorder) *Service {
	return &Service{store: dataStore, recorder: recorder, nowFunc: time.Now}
}

func (s *Service) ListAvailability(
	ctx context.Context,
	userID int64,
) ([]AvailabilityWindowView, error) {
	windows, err := s.store.ListAvailabilityWindows(ctx, nil, []int64{userID})
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]AvailabilityWindowView, 0, len(windows))
	for _, window := range windows {
		views = append(views, availabilityView(window))
	}
	return views, nil
}

func (s *Service) ReplaceAvailability(
	ctx context.Context,
	userID int64,
	inputs []AvailabilityWindowInput,
) ([]AvailabilityWindowView, error) {
	if len(inputs) > maxAvailabilityWindows {
		return nil, invalidParams("最多保存 24 个可约时间段")
	}
	now := s.now()
	normalized := make([]store.CreateAvailabilityWindowParams, 0, len(inputs))
	for _, input := range inputs {
		params, err := s.normalizeAvailabilityInput(userID, input, now)
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, params)
	}
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.store.DeleteAvailabilityWindows(ctx, tx, userID); err != nil {
			return err
		}
		for _, params := range normalized {
			if _, err := s.store.CreateAvailabilityWindow(ctx, tx, params); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return s.ListAvailability(ctx, userID)
}

func (s *Service) SuggestTimes(
	ctx context.Context,
	userID int64,
	input SuggestionInput,
) (SuggestionResult, error) {
	details, err := s.resolveContext(ctx, userID, input.Context)
	if err != nil {
		return SuggestionResult{}, err
	}
	if len(details.ParticipantIDs) < 2 || len(details.ParticipantIDs) > maxMeetupParticipants {
		return SuggestionResult{}, invalidParams("共同空闲仅支持 2 到 20 位参与者")
	}
	now := s.now()
	if input.From.Before(now.Add(-5 * time.Minute)) {
		input.From = now
	}
	if !input.To.After(input.From) || input.To.Sub(input.From) > maxSuggestionRange {
		return SuggestionResult{}, invalidParams("候选时间范围需要在 31 天内")
	}
	if input.DurationMinutes < 30 || input.DurationMinutes > 8*60 {
		return SuggestionResult{}, invalidParams("见面时长需要在 30 分钟到 8 小时之间")
	}

	windows, err := s.store.ListAvailabilityWindows(ctx, nil, details.ParticipantIDs)
	if err != nil {
		return SuggestionResult{}, httpapi.InternalError(err)
	}
	busy, err := s.store.ListBusyIntervals(
		ctx,
		nil,
		details.ParticipantIDs,
		input.From,
		input.To,
		0,
	)
	if err != nil {
		return SuggestionResult{}, httpapi.InternalError(err)
	}
	candidates, missing := buildCommonSuggestions(
		windows,
		busy,
		details.ParticipantIDs,
		input.From,
		input.To,
		time.Duration(input.DurationMinutes)*time.Minute,
	)
	return SuggestionResult{
		ParticipantCount:         len(details.ParticipantIDs),
		MissingAvailabilityCount: missing,
		Candidates:               candidates,
		Places:                   details.SuggestedPlaces,
	}, nil
}

func (s *Service) Create(
	ctx context.Context,
	userID int64,
	input CreateMeetupInput,
) (MeetupView, error) {
	details, err := s.resolveContext(ctx, userID, input.Context)
	if err != nil {
		return MeetupView{}, err
	}
	if len(details.ParticipantIDs) < 2 || len(details.ParticipantIDs) > maxMeetupParticipants {
		return MeetupView{}, invalidParams("约见需要 2 到 20 位参与者")
	}
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		input.Title = details.Title
	}
	input.Category = normalizeCategory(input.Category, details.Category)
	input.Location = strings.TrimSpace(input.Location)
	if err := s.validateMeetupDetails(
		input.Title,
		input.StartsAt,
		input.EndsAt,
		input.Location,
		0,
	); err != nil {
		return MeetupView{}, err
	}
	busy, err := s.store.ListBusyIntervals(
		ctx,
		nil,
		details.ParticipantIDs,
		input.StartsAt,
		input.EndsAt,
		0,
	)
	if err != nil {
		return MeetupView{}, httpapi.InternalError(err)
	}
	if hasBusyConflict(busy, input.StartsAt, input.EndsAt) {
		return MeetupView{}, httpapi.NewError(
			http.StatusConflict,
			httpapi.CodeConflict,
			"这个时段与至少一位参与者已接受的日程冲突，请重新选择",
		)
	}

	now := s.now()
	var created store.Meetup
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		createdMeetup, err := s.store.CreateMeetup(ctx, tx, store.CreateMeetupParams{
			ContextType: details.Ref.Type,
			ContextID:   details.Ref.ID,
			OrganizerID: userID,
			Title:       input.Title,
			Category:    input.Category,
			StartsAt:    input.StartsAt,
			EndsAt:      input.EndsAt,
			Location:    input.Location,
			Now:         now,
		})
		if err != nil {
			return err
		}
		created = createdMeetup
		for _, participantID := range details.ParticipantIDs {
			status := "pending"
			if participantID == userID {
				status = "accepted"
			}
			if err := s.store.AddMeetupParticipant(
				ctx,
				tx,
				created.ID,
				participantID,
				status,
				now,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return MeetupView{}, httpapi.InternalError(err)
	}
	s.record(ctx, analytics.EventMeetupProposed, userID, created.ID, input.Category, len(details.ParticipantIDs), "proposed")
	return s.Get(ctx, userID, created.ID)
}

func (s *Service) Get(ctx context.Context, userID, meetupID int64) (MeetupView, error) {
	meetupRecord, err := s.store.GetMeetup(ctx, nil, userID, meetupID)
	if err != nil {
		return MeetupView{}, mapMeetupStoreError(err)
	}
	return s.buildView(ctx, userID, meetupRecord)
}

func (s *Service) List(
	ctx context.Context,
	userID int64,
	contextRef *ContextRef,
	from time.Time,
	to time.Time,
) ([]MeetupView, error) {
	var records []store.Meetup
	var err error
	if contextRef != nil {
		if _, resolveErr := s.resolveContext(ctx, userID, *contextRef); resolveErr != nil {
			return nil, resolveErr
		}
		records, err = s.store.ListMeetupsForContext(
			ctx,
			nil,
			userID,
			contextRef.Type,
			contextRef.ID,
			50,
		)
	} else {
		if from.IsZero() {
			from = s.now().Add(-30 * 24 * time.Hour)
		}
		if to.IsZero() {
			to = s.now().Add(180 * 24 * time.Hour)
		}
		records, err = s.store.ListMeetupsForUser(ctx, nil, userID, from, to, 100)
	}
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]MeetupView, 0, len(records))
	for _, record := range records {
		view, buildErr := s.buildView(ctx, userID, record)
		if buildErr != nil {
			return nil, buildErr
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *Service) buildView(
	ctx context.Context,
	userID int64,
	record store.Meetup,
) (MeetupView, error) {
	participants, err := s.store.ListMeetupParticipants(ctx, nil, record.ID)
	if err != nil {
		return MeetupView{}, httpapi.InternalError(err)
	}
	participantViews := make([]ParticipantView, 0, len(participants))
	myStatus := ""
	for _, participant := range participants {
		if participant.UserID == userID {
			myStatus = participant.Status
		}
		participantViews = append(participantViews, ParticipantView{
			UserID:      participant.UserID,
			Nickname:    participant.Nickname,
			Department:  participant.Department,
			Grade:       participant.Grade,
			Status:      participant.Status,
			IsMe:        participant.UserID == userID,
			IsOrganizer: participant.UserID == record.OrganizerID,
		})
	}
	var feedbackView *FeedbackView
	feedback, feedbackErr := s.store.GetMeetupFeedback(ctx, nil, record.ID, userID)
	if feedbackErr == nil {
		feedbackView = toFeedbackView(feedback)
	} else if !errors.Is(feedbackErr, store.ErrNotFound) {
		return MeetupView{}, httpapi.InternalError(feedbackErr)
	}
	calendarEventID := nullableInt64Pointer(record.CalendarEventID)
	active := record.Status == "proposed" || record.Status == "reschedule_requested" ||
		record.Status == "confirmed"
	now := s.now()
	return MeetupView{
		ID:              record.ID,
		Context:         ContextRef{Type: record.ContextType, ID: record.ContextID},
		OrganizerID:     record.OrganizerID,
		Title:           record.Title,
		Category:        record.Category,
		StartsAt:        record.StartsAt.UTC().Format(time.RFC3339),
		EndsAt:          record.EndsAt.UTC().Format(time.RFC3339),
		Location:        record.Location,
		Status:          record.Status,
		Version:         record.Version,
		CalendarEventID: calendarEventID,
		Participants:    participantViews,
		MyStatus:        myStatus,
		CanRespond: (record.Status == "proposed" || record.Status == "reschedule_requested") &&
			myStatus == "pending",
		CanReschedule:   active,
		CanCancel:       active,
		CanComplete:     record.Status == "confirmed" && !now.Before(record.StartsAt),
		Reminder:        reminderFor(record, myStatus, feedbackView != nil, now),
		SuggestedPlaces: placesFor(record.Category),
		MyFeedback:      feedbackView,
		CreatedAt:       record.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       record.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) resolveContext(
	ctx context.Context,
	userID int64,
	ref ContextRef,
) (contextDetails, error) {
	ref.Type = strings.TrimSpace(ref.Type)
	if ref.ID <= 0 || (ref.Type != ContextMatchRoom && ref.Type != ContextDirect && ref.Type != ContextGroup) {
		return contextDetails{}, invalidParams("约见上下文不正确")
	}
	switch ref.Type {
	case ContextMatchRoom:
		room, err := s.store.GetMatchRoom(ctx, nil, userID, ref.ID)
		if err != nil {
			return contextDetails{}, mapMeetupStoreError(err)
		}
		category := normalizeCategory(room.IntentCategory, "other")
		places := make([]CampusPlace, 0, 5)
		seen := make(map[string]struct{})
		for _, name := range room.SuggestedPlaces {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			places = append(places, CampusPlace{Name: name, Area: "清华校内", BestFor: []string{category}})
			seen[name] = struct{}{}
		}
		for _, place := range placesFor(category) {
			if _, exists := seen[place.Name]; !exists {
				places = append(places, place)
			}
			if len(places) == 5 {
				break
			}
		}
		return contextDetails{
			Ref:             ref,
			Title:           room.IntentTitle,
			Category:        category,
			ParticipantIDs:  normalizedUserIDs([]int64{room.OwnerID, room.CandidateID}),
			SuggestedPlaces: places,
		}, nil
	case ContextDirect:
		conversation, err := s.store.GetConversation(ctx, nil, ref.ID)
		if err != nil {
			return contextDetails{}, mapMeetupStoreError(err)
		}
		if userID != conversation.User1ID && userID != conversation.User2ID {
			return contextDetails{}, forbidden("你不在这个会话中")
		}
		if conversation.FriendshipDisabledAt.Valid {
			return contextDetails{}, forbidden("好友关系已结束，不能继续发起约见")
		}
		return contextDetails{
			Ref:             ref,
			Title:           "一起在校内见面",
			Category:        "other",
			ParticipantIDs:  normalizedUserIDs([]int64{conversation.User1ID, conversation.User2ID}),
			SuggestedPlaces: placesFor("other"),
		}, nil
	case ContextGroup:
		if _, err := s.store.GetChatGroupMember(ctx, nil, ref.ID, userID); err != nil {
			return contextDetails{}, mapMeetupStoreError(err)
		}
		group, err := s.store.GetChatGroup(ctx, nil, ref.ID)
		if err != nil {
			return contextDetails{}, mapMeetupStoreError(err)
		}
		members, err := s.store.ListChatGroupMembers(ctx, nil, ref.ID)
		if err != nil {
			return contextDetails{}, httpapi.InternalError(err)
		}
		participantIDs := make([]int64, 0, len(members))
		for _, member := range members {
			participantIDs = append(participantIDs, member.UserID)
		}
		return contextDetails{
			Ref:             ref,
			Title:           group.Name,
			Category:        "other",
			ParticipantIDs:  normalizedUserIDs(participantIDs),
			SuggestedPlaces: placesFor("other"),
		}, nil
	default:
		return contextDetails{}, invalidParams("约见上下文不正确")
	}
}

func (s *Service) normalizeAvailabilityInput(
	userID int64,
	input AvailabilityWindowInput,
	now time.Time,
) (store.CreateAvailabilityWindowParams, error) {
	input.WindowType = strings.TrimSpace(input.WindowType)
	input.Timezone = strings.TrimSpace(input.Timezone)
	if input.Timezone == "" {
		input.Timezone = defaultTimezone
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return store.CreateAvailabilityWindowParams{}, invalidParams("时区不正确")
	}
	params := store.CreateAvailabilityWindowParams{
		UserID:     userID,
		WindowType: input.WindowType,
		Timezone:   input.Timezone,
		Now:        now,
	}
	switch input.WindowType {
	case "one_time":
		if input.StartsAt == nil || input.EndsAt == nil {
			return params, invalidParams("一次性可约时间需要开始和结束时间")
		}
		if input.EndsAt.Sub(*input.StartsAt) < 30*time.Minute ||
			input.EndsAt.Sub(*input.StartsAt) > 12*time.Hour ||
			input.StartsAt.Before(now.Add(-5*time.Minute)) ||
			input.StartsAt.After(now.Add(maxMeetupHorizon)) {
			return params, invalidParams("一次性可约时间需要在未来 180 天内，持续 30 分钟到 12 小时")
		}
		params.StartsAt = sql.NullTime{Time: input.StartsAt.UTC(), Valid: true}
		params.EndsAt = sql.NullTime{Time: input.EndsAt.UTC(), Valid: true}
	case "weekly":
		if input.Weekday < 1 || input.Weekday > 7 || input.StartMinute < 0 ||
			input.EndMinute > 1440 || input.EndMinute-input.StartMinute < 30 {
			return params, invalidParams("周期时间需要选择星期，并至少持续 30 分钟")
		}
		params.Weekday = sql.NullInt64{Int64: int64(input.Weekday), Valid: true}
		params.StartMinute = sql.NullInt64{Int64: int64(input.StartMinute), Valid: true}
		params.EndMinute = sql.NullInt64{Int64: int64(input.EndMinute), Valid: true}
		if input.ValidFrom != nil {
			params.ValidFrom = sql.NullTime{Time: *input.ValidFrom, Valid: true}
		}
		if input.ValidUntil != nil {
			params.ValidUntil = sql.NullTime{Time: *input.ValidUntil, Valid: true}
		}
		if params.ValidFrom.Valid && params.ValidUntil.Valid &&
			params.ValidUntil.Time.Before(params.ValidFrom.Time) {
			return params, invalidParams("周期时间的结束日期不能早于开始日期")
		}
	default:
		return params, invalidParams("可约时间类型不正确")
	}
	return params, nil
}

func (s *Service) validateMeetupDetails(
	title string,
	startsAt time.Time,
	endsAt time.Time,
	location string,
	excludeCalendarEventID int64,
) error {
	if strings.TrimSpace(title) == "" || utf8.RuneCountInString(title) > maxMeetupTitleLength {
		return invalidParams("约见标题需要在 1 到 120 个字之间")
	}
	if utf8.RuneCountInString(location) > maxMeetupLocationLength {
		return invalidParams("约见地点不能超过 160 个字")
	}
	duration := endsAt.Sub(startsAt)
	if startsAt.Before(s.now().Add(-5*time.Minute)) || startsAt.After(s.now().Add(maxMeetupHorizon)) ||
		duration < 30*time.Minute || duration > 8*time.Hour {
		return invalidParams("约见需要在未来 180 天内，持续 30 分钟到 8 小时")
	}
	_ = excludeCalendarEventID
	return nil
}

func (s *Service) now() time.Time {
	if s.nowFunc == nil {
		return time.Now()
	}
	return s.nowFunc()
}

func (s *Service) record(
	ctx context.Context,
	eventName string,
	actorID int64,
	meetupID int64,
	category string,
	memberCount int,
	status string,
) {
	if s.recorder == nil {
		return
	}
	_ = s.recorder.Record(ctx, analytics.EventInput{
		Name:        eventName,
		ActorUserID: actorID,
		SubjectType: "meetup",
		SubjectID:   meetupID,
		RequestID:   middleware.GetReqID(ctx),
		Properties: map[string]any{
			"source":      "api",
			"category":    category,
			"memberCount": memberCount,
			"status":      status,
		},
	})
}

func availabilityView(window store.AvailabilityWindow) AvailabilityWindowView {
	view := AvailabilityWindowView{
		ID:         window.ID,
		WindowType: window.WindowType,
		Timezone:   window.Timezone,
	}
	if window.StartsAt.Valid {
		value := window.StartsAt.Time.UTC().Format(time.RFC3339)
		view.StartsAt = &value
	}
	if window.EndsAt.Valid {
		value := window.EndsAt.Time.UTC().Format(time.RFC3339)
		view.EndsAt = &value
	}
	if window.Weekday.Valid {
		view.Weekday = int(window.Weekday.Int64)
	}
	if window.StartMinute.Valid {
		view.StartMinute = int(window.StartMinute.Int64)
	}
	if window.EndMinute.Valid {
		view.EndMinute = int(window.EndMinute.Int64)
	}
	if window.ValidFrom.Valid {
		value := window.ValidFrom.Time.Format("2006-01-02")
		view.ValidFrom = &value
	}
	if window.ValidUntil.Valid {
		value := window.ValidUntil.Time.Format("2006-01-02")
		view.ValidUntil = &value
	}
	return view
}

func normalizeCategory(category, fallback string) string {
	category = strings.TrimSpace(category)
	switch category {
	case "meal", "study", "sport", "event", "romantic", "project", "other":
		return category
	case "date":
		return "romantic"
	case "sports":
		return "sport"
	case "cofounder":
		return "project"
	}
	if fallback != "" && fallback != category {
		return normalizeCategory(fallback, "other")
	}
	return "other"
}

func normalizedUserIDs(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func toFeedbackView(feedback store.MeetupFeedback) *FeedbackView {
	view := &FeedbackView{
		Attendance:    feedback.Attendance,
		SafetyConcern: feedback.SafetyConcern,
		PrivateNotes:  feedback.PrivateNotes,
		SubmittedAt:   feedback.SubmittedAt.UTC().Format(time.RFC3339),
	}
	if feedback.Rating.Valid {
		value := int(feedback.Rating.Int64)
		view.Rating = &value
	}
	if feedback.WouldMeetAgain.Valid {
		value := feedback.WouldMeetAgain.Bool
		view.WouldMeetAgain = &value
	}
	return view
}

func nullableInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func reminderFor(record store.Meetup, myStatus string, hasFeedback bool, now time.Time) ReminderView {
	switch record.Status {
	case "proposed", "reschedule_requested":
		if myStatus == "pending" {
			return ReminderView{Stage: "awaiting_response", Text: "这项约见正在等你确认"}
		}
		return ReminderView{Stage: "awaiting_others", Text: "已回复，正在等待其他参与者"}
	case "confirmed":
		if !now.Before(record.EndsAt) {
			if hasFeedback {
				return ReminderView{Stage: "complete", Text: "约见已结束，反馈仅你可见"}
			}
			return ReminderView{Stage: "feedback_due", Text: "约见已结束，可以留下私密反馈"}
		}
		until := record.StartsAt.Sub(now)
		if until <= 0 {
			return ReminderView{Stage: "underway", Text: "约见时间已到，注意确认地点与安全"}
		}
		if until <= time.Hour {
			return ReminderView{Stage: "soon", Text: "约见将在 1 小时内开始"}
		}
		if until <= 24*time.Hour {
			return ReminderView{Stage: "day_before", Text: "约见将在 24 小时内开始"}
		}
		return ReminderView{Stage: "confirmed", Text: "所有参与者已确认，已加入日历"}
	case "completed", "no_show":
		if hasFeedback {
			return ReminderView{Stage: "complete", Text: "反馈已保存，仅你可见并可撤回"}
		}
		return ReminderView{Stage: "feedback_due", Text: "请记录这次见面的私密反馈"}
	case "declined":
		return ReminderView{Stage: "closed", Text: "有参与者拒绝了这项提议"}
	case "cancelled":
		return ReminderView{Stage: "closed", Text: "这项约见已取消"}
	default:
		return ReminderView{Stage: "none", Text: ""}
	}
}

func mapMeetupStoreError(err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "找不到这项约见")
	}
	return httpapi.InternalError(err)
}

func invalidParams(message string) error {
	return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, message)
}

func forbidden(message string) error {
	return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, message)
}

func conflict(message string) error {
	return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, message)
}

func requireExpectedVersion(record store.Meetup, expectedVersion int) error {
	if expectedVersion <= 0 || record.Version != expectedVersion {
		return conflict("约见已被其他参与者更新，请刷新后重试")
	}
	return nil
}

func calendarEventType(category string) string {
	switch category {
	case "sport":
		return "sports"
	case "romantic":
		return "date"
	case "meal", "study", "project", "other":
		return category
	default:
		return "meetup"
	}
}

func signalFromFeedback(input FeedbackInput) (string, int) {
	if input.Attendance == "no_show" {
		return "no_show", -3
	}
	if input.WouldMeetAgain != nil {
		if *input.WouldMeetAgain {
			return "meet_again", 2
		}
		return "do_not_repeat", -2
	}
	if input.Attendance == "attended" {
		return "positive_meetup", 1
	}
	return "", 0
}

func wrapStoreTxError(err error) error {
	var appErr *httpapi.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	if errors.Is(err, store.ErrNotFound) {
		return conflict("约见状态已经变化，请刷新后重试")
	}
	return httpapi.InternalError(fmt.Errorf("meetup transaction: %w", err))
}
