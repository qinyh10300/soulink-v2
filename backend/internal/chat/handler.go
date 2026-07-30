package chat

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	service *Service
	hub     *Hub
}

func NewHandler(service *Service, hub *Hub) *Handler {
	return &Handler{service: service, hub: hub}
}

func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	convs, err := h.service.ListConversations(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, convs)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	convID, err := strconv.ParseInt(chi.URLParam(r, "conversationId"), 10, 64)
	if err != nil || convID <= 0 {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
		return
	}

	var beforeID int64
	if raw := r.URL.Query().Get("before"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
			return
		}
		beforeID = parsed
	}

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
			return
		}
		limit = parsed
	}

	result, err := h.service.ListMessages(r.Context(), user.ID, convID, beforeID, limit)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	convID, err := strconv.ParseInt(chi.URLParam(r, "conversationId"), 10, 64)
	if err != nil || convID <= 0 {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
		return
	}

	if err := h.service.MarkRead(r.Context(), user.ID, convID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}

func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	count, err := h.service.GetUnreadCount(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, map[string]int{"unreadCount": count})
}
