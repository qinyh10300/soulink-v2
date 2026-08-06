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
	ContextType    string `json:"contextType"`
	ContextID      int64  `json:"contextId"`
	Category       string `json:"category"`
	Details        string `json:"details"`
	BlockUser      bool   `json:"blockUser"`
	ReportedUserID int64  `json:"reportedUserId"`
}

type updateReportRequest struct {
	Status      string `json:"status"`
	ReviewNotes string `json:"reviewNotes"`
	AssignToMe  bool   `json:"assignToMe"`
}

type appealReportRequest struct {
	Reason string `json:"reason"`
}

type reviewReportAppealRequest struct {
	Status      string `json:"status"`
	ReviewNotes string `json:"reviewNotes"`
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
		ContextType:    req.ContextType,
		ContextID:      req.ContextID,
		Category:       req.Category,
		Details:        req.Details,
		BlockUser:      req.BlockUser,
		ReportedUserID: req.ReportedUserID,
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
	report, err := h.service.UpdateReport(
		r.Context(),
		user.ID,
		user.Email,
		reportID,
		ReportUpdateInput{
			Status:      req.Status,
			ReviewNotes: req.ReviewNotes,
			AssignToMe:  req.AssignToMe,
		},
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, report)
}

func (h *Handler) ListMyCases(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	cases, err := h.service.ListMyCases(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, cases)
}

func (h *Handler) AppealReport(w http.ResponseWriter, r *http.Request) {
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
	var request appealReportRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	appeal, err := h.service.AppealReport(
		r.Context(), user.ID, reportID, ReportAppealInput{Reason: request.Reason},
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "safety report appealed", appeal)
}

func (h *Handler) ReviewReportAppeal(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	appealID, err := strconv.ParseInt(chi.URLParam(r, "appealId"), 10, 64)
	if err != nil || appealID <= 0 {
		httpapi.Error(w, invalidParams("无效的申诉编号"))
		return
	}
	var request reviewReportAppealRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	appeal, err := h.service.ReviewReportAppeal(
		r.Context(), user.ID, user.Email, appealID,
		ReportAppealReviewInput{Status: request.Status, ReviewNotes: request.ReviewNotes},
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, appeal)
}
