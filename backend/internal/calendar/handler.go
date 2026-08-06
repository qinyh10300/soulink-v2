package calendar

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type Handler struct {
	service *Service
}

type eventRequest struct {
	EventType   string  `json:"eventType"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	StartsAt    string  `json:"startsAt"`
	EndsAt      string  `json:"endsAt"`
	Visibility  string  `json:"visibility"`
	InviteeIDs  []int64 `json:"inviteeIds"`
}

type rsvpRequest struct {
	Status string `json:"status"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	from, to, err := parseRange(r)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	events, err := h.service.ListEvents(r.Context(), user.ID, from, to)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, events)
}

func (h *Handler) ListUserSchedule(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	targetID, err := parseRouteID(r, "userId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	from, to, err := parseRange(r)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	events, err := h.service.ListUserSchedule(r.Context(), user.ID, targetID, from, to)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, events)
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	input, err := decodeEventRequest(r)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	event, err := h.service.CreateEvent(r.Context(), user.ID, input)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "event created", event)
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	eventID, err := parseRouteID(r, "eventId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	input, err := decodeEventRequest(r)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	event, err := h.service.UpdateEvent(r.Context(), user.ID, eventID, input)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, event)
}

func (h *Handler) CancelEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	eventID, err := parseRouteID(r, "eventId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	if err := h.service.CancelEvent(r.Context(), user.ID, eventID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}

func (h *Handler) Respond(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(w, r)
	if !ok {
		return
	}
	eventID, err := parseRouteID(r, "eventId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var request rsvpRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	event, err := h.service.Respond(r.Context(), user.ID, eventID, request.Status)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, event)
}

func currentUser(w http.ResponseWriter, r *http.Request) (store.User, bool) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
	}
	return user, ok
}

func decodeEventRequest(r *http.Request) (EventInput, error) {
	var request eventRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		return EventInput{}, err
	}
	startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
	if err != nil {
		return EventInput{}, invalidParams("开始时间不正确")
	}
	endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
	if err != nil {
		return EventInput{}, invalidParams("结束时间不正确")
	}
	return EventInput{
		EventType:   request.EventType,
		Title:       request.Title,
		Description: request.Description,
		Location:    request.Location,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Visibility:  request.Visibility,
		InviteeIDs:  request.InviteeIDs,
	}, nil
}

func parseRange(r *http.Request) (time.Time, time.Time, error) {
	fromRaw := r.URL.Query().Get("from")
	toRaw := r.URL.Query().Get("to")
	if fromRaw == "" || toRaw == "" {
		now := time.Now()
		return now, now.Add(7 * 24 * time.Hour), nil
	}
	from, err := time.Parse(time.RFC3339, fromRaw)
	if err != nil {
		return time.Time{}, time.Time{}, invalidParams("开始范围不正确")
	}
	to, err := time.Parse(time.RFC3339, toRaw)
	if err != nil {
		return time.Time{}, time.Time{}, invalidParams("结束范围不正确")
	}
	return from, to, nil
}

func parseRouteID(r *http.Request, key string) (int64, error) {
	value, err := strconv.ParseInt(chi.URLParam(r, key), 10, 64)
	if err != nil || value <= 0 {
		return 0, invalidParams("参数错误")
	}
	return value, nil
}
