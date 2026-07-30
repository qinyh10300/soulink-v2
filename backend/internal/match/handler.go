package match

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	service *Service
}

func NewHandler(service ...*Service) *Handler {
	if len(service) == 0 {
		return &Handler{}
	}
	return &Handler{service: service[0]}
}

func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("matches/current"))
		return
	}

	result, err := h.service.GetCurrent(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("matches/history"))
		return
	}

	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
			return
		}
		limit = parsed
	}

	result, err := h.service.GetHistory(r.Context(), user.ID, limit)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetDetail(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("matches/{matchId}"))
		return
	}

	matchID, err := strconv.ParseInt(chi.URLParam(r, "matchId"), 10, 64)
	if err != nil || matchID <= 0 {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误"))
		return
	}

	result, err := h.service.GetDetail(r.Context(), user.ID, matchID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ContactMatchedPartner(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("matches/{matchId}/contact"))
		return
	}

	matchID, err := strconv.ParseInt(chi.URLParam(r, "matchId"), 10, 64)
	if err != nil || matchID <= 0 {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid match id"))
		return
	}

	result, err := h.service.ContactMatchedPartner(r.Context(), user.ID, matchID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetCycleStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("matches/cycle/status"))
		return
	}

	result, err := h.service.GetCycleStatus(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}
