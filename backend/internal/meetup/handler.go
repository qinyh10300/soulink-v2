package meetup

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	service *Service
}

type availabilityWindowRequest struct {
	WindowType  string `json:"windowType"`
	Timezone    string `json:"timezone"`
	StartsAt    string `json:"startsAt"`
	EndsAt      string `json:"endsAt"`
	Weekday     int    `json:"weekday"`
	StartMinute int    `json:"startMinute"`
	EndMinute   int    `json:"endMinute"`
	ValidFrom   string `json:"validFrom"`
	ValidUntil  string `json:"validUntil"`
}

type replaceAvailabilityRequest struct {
	Windows []availabilityWindowRequest `json:"windows"`
}

type contextRequest struct {
	Type string `json:"type"`
	ID   int64  `json:"id"`
}

type suggestionRequest struct {
	Context         contextRequest `json:"context"`
	From            string         `json:"from"`
	To              string         `json:"to"`
	DurationMinutes int            `json:"durationMinutes"`
}

type createMeetupRequest struct {
	Context  contextRequest `json:"context"`
	Title    string         `json:"title"`
	Category string         `json:"category"`
	StartsAt string         `json:"startsAt"`
	EndsAt   string         `json:"endsAt"`
	Location string         `json:"location"`
}

type responseRequest struct {
	Action          string `json:"action"`
	ExpectedVersion int    `json:"expectedVersion"`
}

type rescheduleRequest struct {
	StartsAt        string `json:"startsAt"`
	EndsAt          string `json:"endsAt"`
	Location        string `json:"location"`
	ExpectedVersion int    `json:"expectedVersion"`
}

type cancelRequest struct {
	ExpectedVersion int    `json:"expectedVersion"`
	Reason          string `json:"reason"`
}

type completeRequest struct {
	ExpectedVersion int    `json:"expectedVersion"`
	Outcome         string `json:"outcome"`
}

type feedbackRequest struct {
	Attendance     string `json:"attendance"`
	Rating         *int   `json:"rating"`
	WouldMeetAgain *bool  `json:"wouldMeetAgain"`
	SafetyConcern  bool   `json:"safetyConcern"`
	PrivateNotes   string `json:"privateNotes"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListAvailability(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	windows, err := h.service.ListAvailability(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, windows)
}

func (h *Handler) ReplaceAvailability(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var request replaceAvailabilityRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	inputs := make([]AvailabilityWindowInput, 0, len(request.Windows))
	for _, window := range request.Windows {
		input, err := parseAvailabilityWindow(window)
		if err != nil {
			httpapi.Error(w, err)
			return
		}
		inputs = append(inputs, input)
	}
	windows, err := h.service.ReplaceAvailability(r.Context(), userID, inputs)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, windows)
}

func (h *Handler) SuggestTimes(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var request suggestionRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	from, err := parseRFC3339(request.From, "候选开始时间不正确")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	to, err := parseRFC3339(request.To, "候选结束时间不正确")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.SuggestTimes(r.Context(), userID, SuggestionInput{
		Context:         ContextRef{Type: request.Context.Type, ID: request.Context.ID},
		From:            from,
		To:              to,
		DurationMinutes: request.DurationMinutes,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var request createMeetupRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	startsAt, err := parseRFC3339(request.StartsAt, "约见开始时间不正确")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	endsAt, err := parseRFC3339(request.EndsAt, "约见结束时间不正确")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.Create(r.Context(), userID, CreateMeetupInput{
		Context:  ContextRef{Type: request.Context.Type, ID: request.Context.ID},
		Title:    request.Title,
		Category: request.Category,
		StartsAt: startsAt,
		EndsAt:   endsAt,
		Location: request.Location,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "meetup proposed", result)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var contextRef *ContextRef
	contextType := strings.TrimSpace(r.URL.Query().Get("contextType"))
	contextIDRaw := strings.TrimSpace(r.URL.Query().Get("contextId"))
	if contextType != "" || contextIDRaw != "" {
		contextID, err := strconv.ParseInt(contextIDRaw, 10, 64)
		if err != nil || contextID <= 0 || contextType == "" {
			httpapi.Error(w, invalidParams("约见上下文不正确"))
			return
		}
		contextRef = &ContextRef{Type: contextType, ID: contextID}
	}
	var from, to time.Time
	var err error
	if raw := strings.TrimSpace(r.URL.Query().Get("from")); raw != "" {
		from, err = parseRFC3339(raw, "开始范围不正确")
		if err != nil {
			httpapi.Error(w, err)
			return
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("to")); raw != "" {
		to, err = parseRFC3339(raw, "结束范围不正确")
		if err != nil {
			httpapi.Error(w, err)
			return
		}
	}
	result, err := h.service.List(r.Context(), userID, contextRef, from, to)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, meetupID, ok := meetupRequestIDs(w, r)
	if !ok {
		return
	}
	result, err := h.service.Get(r.Context(), userID, meetupID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) Respond(w http.ResponseWriter, r *http.Request) {
	userID, meetupID, ok := meetupRequestIDs(w, r)
	if !ok {
		return
	}
	var request responseRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.Respond(
		r.Context(),
		userID,
		meetupID,
		request.Action,
		request.ExpectedVersion,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) Reschedule(w http.ResponseWriter, r *http.Request) {
	userID, meetupID, ok := meetupRequestIDs(w, r)
	if !ok {
		return
	}
	var request rescheduleRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	startsAt, err := parseRFC3339(request.StartsAt, "改期开始时间不正确")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	endsAt, err := parseRFC3339(request.EndsAt, "改期结束时间不正确")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.Reschedule(r.Context(), userID, meetupID, RescheduleInput{
		StartsAt:        startsAt,
		EndsAt:          endsAt,
		Location:        request.Location,
		ExpectedVersion: request.ExpectedVersion,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, meetupID, ok := meetupRequestIDs(w, r)
	if !ok {
		return
	}
	var request cancelRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.Cancel(
		r.Context(),
		userID,
		meetupID,
		request.ExpectedVersion,
		request.Reason,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userID, meetupID, ok := meetupRequestIDs(w, r)
	if !ok {
		return
	}
	var request completeRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.Complete(
		r.Context(),
		userID,
		meetupID,
		request.ExpectedVersion,
		request.Outcome,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	userID, meetupID, ok := meetupRequestIDs(w, r)
	if !ok {
		return
	}
	var request feedbackRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.SubmitFeedback(r.Context(), userID, meetupID, FeedbackInput{
		Attendance:     request.Attendance,
		Rating:         request.Rating,
		WouldMeetAgain: request.WouldMeetAgain,
		SafetyConcern:  request.SafetyConcern,
		PrivateNotes:   request.PrivateNotes,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) RevokeFeedback(w http.ResponseWriter, r *http.Request) {
	userID, meetupID, ok := meetupRequestIDs(w, r)
	if !ok {
		return
	}
	result, err := h.service.RevokeFeedback(r.Context(), userID, meetupID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func parseAvailabilityWindow(request availabilityWindowRequest) (AvailabilityWindowInput, error) {
	input := AvailabilityWindowInput{
		WindowType:  request.WindowType,
		Timezone:    request.Timezone,
		Weekday:     request.Weekday,
		StartMinute: request.StartMinute,
		EndMinute:   request.EndMinute,
	}
	if request.StartsAt != "" {
		value, err := parseRFC3339(request.StartsAt, "可约开始时间不正确")
		if err != nil {
			return input, err
		}
		input.StartsAt = &value
	}
	if request.EndsAt != "" {
		value, err := parseRFC3339(request.EndsAt, "可约结束时间不正确")
		if err != nil {
			return input, err
		}
		input.EndsAt = &value
	}
	if request.ValidFrom != "" {
		value, err := time.Parse("2006-01-02", request.ValidFrom)
		if err != nil {
			return input, invalidParams("周期开始日期不正确")
		}
		input.ValidFrom = &value
	}
	if request.ValidUntil != "" {
		value, err := time.Parse("2006-01-02", request.ValidUntil)
		if err != nil {
			return input, invalidParams("周期结束日期不正确")
		}
		input.ValidUntil = &value
	}
	return input, nil
}

func parseRFC3339(value, message string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, invalidParams(message)
	}
	return parsed, nil
}

func currentUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return 0, false
	}
	return user.ID, true
}

func meetupRequestIDs(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return 0, 0, false
	}
	meetupID, err := strconv.ParseInt(chi.URLParam(r, "meetupId"), 10, 64)
	if err != nil || meetupID <= 0 {
		httpapi.Error(w, invalidParams("约见编号不正确"))
		return 0, 0, false
	}
	return userID, meetupID, true
}
