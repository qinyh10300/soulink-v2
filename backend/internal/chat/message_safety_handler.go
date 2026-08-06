package chat

import (
	"net/http"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type checkMessageSafetyRequest struct {
	Content string `json:"content"`
}

func (h *Handler) CheckDirectMessageSafety(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	conversationID, err := parseChatRouteID(r, "conversationId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var request checkMessageSafetyRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.CheckDirectMessageSafety(r.Context(), user.ID, conversationID, request.Content)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) CheckGroupMessageSafety(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := parseChatRouteID(r, "groupId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var request checkMessageSafetyRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.CheckGroupMessageSafety(r.Context(), user.ID, groupID, request.Content)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}
