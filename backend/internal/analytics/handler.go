package analytics

import (
	"net/http"
	"strconv"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	days := 28
	if rawDays := r.URL.Query().Get("days"); rawDays != "" {
		parsedDays, err := strconv.Atoi(rawDays)
		if err != nil {
			httpapi.Error(w, httpapi.NewError(
				http.StatusBadRequest,
				httpapi.CodeInvalidParams,
				"统计周期格式错误",
			))
			return
		}
		days = parsedDays
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("product analytics"))
		return
	}
	dashboard, err := h.service.Dashboard(r.Context(), user.Email, days)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, dashboard)
}
