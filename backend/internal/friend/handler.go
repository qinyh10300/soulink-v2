package friend

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SendRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	var body struct {
		Nickname string `json:"nickname"`
	}
	if err := httpapi.DecodeJSON(r, &body); err != nil {
		httpapi.Error(w, err)
		return
	}
	body.Nickname = strings.TrimSpace(body.Nickname)
	if body.Nickname == "" {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
		return
	}

	friendship, err := h.service.SendRequest(r.Context(), user.ID, body.Nickname)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, map[string]int64{"friendshipId": friendship.ID})
}

func (h *Handler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	var body struct {
		FriendshipID int64 `json:"friendshipId"`
	}
	if err := httpapi.DecodeJSON(r, &body); err != nil {
		httpapi.Error(w, err)
		return
	}
	if body.FriendshipID <= 0 {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
		return
	}

	result, err := h.service.AcceptRequest(r.Context(), user.ID, body.FriendshipID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	var body struct {
		FriendshipID int64 `json:"friendshipId"`
	}
	if err := httpapi.DecodeJSON(r, &body); err != nil {
		httpapi.Error(w, err)
		return
	}
	if body.FriendshipID <= 0 {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
		return
	}

	if err := h.service.RejectRequest(r.Context(), user.ID, body.FriendshipID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}

func (h *Handler) ListFriends(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	friends, err := h.service.ListFriends(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, friends)
}

func (h *Handler) ListReceivedRequests(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	reqs, err := h.service.ListReceivedRequests(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, reqs)
}

func (h *Handler) ListSentRequests(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	reqs, err := h.service.ListSentRequests(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, reqs)
}

func (h *Handler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	friendshipID, err := strconv.ParseInt(chi.URLParam(r, "friendshipId"), 10, 64)
	if err != nil || friendshipID <= 0 {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
		return
	}

	if err := h.service.RemoveFriend(r.Context(), user.ID, friendshipID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}
