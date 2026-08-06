package connection

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

type interpretRequest struct {
	Text string `json:"text"`
}

type interestRequest struct {
	Message string `json:"message"`
}

type roomPlanRequest struct {
	ScheduledAt *string `json:"scheduledAt"`
	Place       string  `json:"place"`
}

type roomStatusRequest struct {
	Status string `json:"status"`
}

type contactSharingRequest struct {
	Shared bool `json:"shared"`
}

type roomMessageRequest struct {
	Content string `json:"content"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Interpret(w http.ResponseWriter, r *http.Request) {
	if _, ok := currentUserID(w, r); !ok {
		return
	}
	var req interpretRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.Interpret(r.Context(), req.Text)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetMatchingProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.GetMatchingProfile(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) UpdateMatchingProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var req MatchingProfileInput
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.UpdateMatchingProfile(r.Context(), userID, req)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ListFeed(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListFeed(r.Context(), userID, r.URL.Query().Get("category"))
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ListSaved(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListSavedIntents(r.Context(), userID, r.URL.Query().Get("category"))
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListMyIntents(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) CreateIntent(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var req IntentInput
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.CreateIntent(r.Context(), userID, req)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "intent created", result)
}

func (h *Handler) GetIntent(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	result, err := h.service.GetIntent(r.Context(), userID, intentID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetContextProfile(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	subjectID, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil || subjectID <= 0 {
		httpapi.Error(w, invalidParams("用户参数无效"))
		return
	}
	result, err := h.service.GetContextProfile(r.Context(), userID, intentID, subjectID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) UpdateIntent(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	var req IntentInput
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.UpdateIntent(r.Context(), userID, intentID, req)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ExpireIntent(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	if err := h.service.ExpireIntent(r.Context(), userID, intentID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, map[string]bool{"expired": true})
}

func (h *Handler) SaveIntent(w http.ResponseWriter, r *http.Request) {
	h.setIntentSaved(w, r, true)
}

func (h *Handler) UnsaveIntent(w http.ResponseWriter, r *http.Request) {
	h.setIntentSaved(w, r, false)
}

func (h *Handler) DismissIntent(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	if err := h.service.DismissIntent(r.Context(), userID, intentID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, map[string]bool{"dismissed": true})
}

func (h *Handler) setIntentSaved(w http.ResponseWriter, r *http.Request, saved bool) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	result, err := h.service.SetIntentSaved(r.Context(), userID, intentID, saved)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ExpressInterest(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	var req interestRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.SetIntentInterest(r.Context(), userID, intentID, req.Message)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) WithdrawInterest(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	result, err := h.service.WithdrawIntentInterest(r.Context(), userID, intentID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ListInterests(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	result, err := h.service.ListIntentInterests(r.Context(), userID, intentID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) AcceptInterest(w http.ResponseWriter, r *http.Request) {
	h.respondToInterest(w, r, true)
}

func (h *Handler) DeclineInterest(w http.ResponseWriter, r *http.Request) {
	h.respondToInterest(w, r, false)
}

func (h *Handler) respondToInterest(w http.ResponseWriter, r *http.Request, accepted bool) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	candidateID, err := parsePositiveID(chi.URLParam(r, "candidateId"))
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	room, err := h.service.RespondToInterest(r.Context(), userID, intentID, candidateID, accepted)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, map[string]any{
		"accepted": accepted,
		"room":     room,
	})
}

func (h *Handler) Recommendations(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "intentId")
	if !ok {
		return
	}
	result, err := h.service.Recommendations(r.Context(), userID, intentID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) DailyRecommendations(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.DailyRecommendations(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetRecommendationPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.GetRecommendationPreferences(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) UpdateRecommendationPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var req RecommendationPreferenceInput
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.UpdateRecommendationPreferences(r.Context(), userID, req)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ResetRecommendationPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.ResetRecommendationPreferences(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ListMatchRooms(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListMatchRooms(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetMatchRoom(w http.ResponseWriter, r *http.Request) {
	userID, roomID, ok := currentUserAndRouteID(w, r, "roomId")
	if !ok {
		return
	}
	result, err := h.service.GetMatchRoom(r.Context(), userID, roomID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) UpdateMatchRoomPlan(w http.ResponseWriter, r *http.Request) {
	userID, roomID, ok := currentUserAndRouteID(w, r, "roomId")
	if !ok {
		return
	}
	var req roomPlanRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	var scheduledAt *time.Time
	if req.ScheduledAt != nil && strings.TrimSpace(*req.ScheduledAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ScheduledAt))
		if err != nil {
			httpapi.Error(w, invalidParams("见面时间格式错误"))
			return
		}
		scheduledAt = &parsed
	}
	result, err := h.service.UpdateMatchRoomPlan(r.Context(), userID, roomID, scheduledAt, req.Place)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) UpdateMatchRoomStatus(w http.ResponseWriter, r *http.Request) {
	userID, roomID, ok := currentUserAndRouteID(w, r, "roomId")
	if !ok {
		return
	}
	var req roomStatusRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.UpdateMatchRoomStatus(r.Context(), userID, roomID, req.Status)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) SetContactSharing(w http.ResponseWriter, r *http.Request) {
	userID, roomID, ok := currentUserAndRouteID(w, r, "roomId")
	if !ok {
		return
	}
	var req contactSharingRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.SetMatchRoomContactShared(r.Context(), userID, roomID, req.Shared)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ListRoomMessages(w http.ResponseWriter, r *http.Request) {
	userID, roomID, ok := currentUserAndRouteID(w, r, "roomId")
	if !ok {
		return
	}
	result, err := h.service.ListMatchRoomMessages(r.Context(), userID, roomID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) CreateRoomMessage(w http.ResponseWriter, r *http.Request) {
	userID, roomID, ok := currentUserAndRouteID(w, r, "roomId")
	if !ok {
		return
	}
	var req roomMessageRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.CreateMatchRoomMessage(r.Context(), userID, roomID, req.Content)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "message created", result)
}

func currentUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return 0, false
	}
	return user.ID, true
}

func currentUserAndRouteID(w http.ResponseWriter, r *http.Request, key string) (int64, int64, bool) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return 0, 0, false
	}
	routeID, err := parsePositiveID(chi.URLParam(r, key))
	if err != nil {
		httpapi.Error(w, err)
		return 0, 0, false
	}
	return userID, routeID, true
}

func parsePositiveID(raw string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0, invalidParams("无效的资源编号")
	}
	return value, nil
}
