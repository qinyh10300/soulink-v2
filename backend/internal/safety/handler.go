package safety

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

type submitReportRequest struct {
	ContextType string `json:"contextType"`
	ContextID   int64  `json:"contextId"`
	Category    string `json:"category"`
	Details     string `json:"details"`
	BlockUser   bool   `json:"blockUser"`
}

type updateReportRequest struct {
	Status string `json:"status"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Capabilities(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	httpapi.OK(w, h.service.Capabilities(user.Email))
}

func (h *Handler) SubmitReport(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	var req submitReportRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.SubmitReport(r.Context(), user.ID, ReportInput{
		ContextType: req.ContextType,
		ContextID:   req.ContextID,
		Category:    req.Category,
		Details:     req.Details,
		BlockUser:   req.BlockUser,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "safety report submitted", result)
}

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	reports, err := h.service.ListReports(r.Context(), user.Email, r.URL.Query().Get("status"))
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, reports)
}

func (h *Handler) UpdateReport(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	reportID, err := strconv.ParseInt(chi.URLParam(r, "reportId"), 10, 64)
	if err != nil || reportID <= 0 {
		httpapi.Error(w, invalidParams("无效的举报编号"))
		return
	}
	var req updateReportRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	report, err := h.service.UpdateReportStatus(
		r.Context(),
		user.ID,
		user.Email,
		reportID,
		req.Status,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, report)
}
