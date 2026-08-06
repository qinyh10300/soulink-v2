package connection

import (
	"net/http"

	"soulink-backend/internal/httpapi"
)

type groupMatchInviteRequest struct {
	CandidateIntentID int64 `json:"candidateIntentId"`
}

type groupMatchResponseRequest struct {
	Status string `json:"status"`
}

func (h *Handler) CreateGroupMatchIntent(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var req GroupMatchIntentInput
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.CreateGroupMatchIntent(r.Context(), userID, req)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "group intent created", result)
}

func (h *Handler) ListMyGroupMatchIntents(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListMyGroupMatchIntents(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetGroupMatchIntent(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "groupIntentId")
	if !ok {
		return
	}
	result, err := h.service.GetGroupMatchIntent(r.Context(), userID, intentID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GroupMatchRecommendations(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "groupIntentId")
	if !ok {
		return
	}
	result, err := h.service.GroupMatchRecommendations(r.Context(), userID, intentID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) InviteGroupMatchCandidate(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "groupIntentId")
	if !ok {
		return
	}
	var req groupMatchInviteRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	if req.CandidateIntentID <= 0 {
		httpapi.Error(w, invalidParams("候选参数无效"))
		return
	}
	result, err := h.service.InviteGroupMatchCandidate(
		r.Context(), userID, intentID, req.CandidateIntentID,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "group invitation created", result)
}

func (h *Handler) ListMyGroupMatchInvitations(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListMyGroupMatchInvitations(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) RespondGroupMatchInvitation(w http.ResponseWriter, r *http.Request) {
	userID, invitationID, ok := currentUserAndRouteID(w, r, "invitationId")
	if !ok {
		return
	}
	var req groupMatchResponseRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.RespondGroupMatchInvitation(
		r.Context(), userID, invitationID, req.Status,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) FormGroupMatchIntent(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "groupIntentId")
	if !ok {
		return
	}
	result, err := h.service.FormGroupMatchIntent(r.Context(), userID, intentID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) LeaveGroupMatchIntent(w http.ResponseWriter, r *http.Request) {
	userID, intentID, ok := currentUserAndRouteID(w, r, "groupIntentId")
	if !ok {
		return
	}
	result, err := h.service.LeaveGroupMatchIntent(r.Context(), userID, intentID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}
