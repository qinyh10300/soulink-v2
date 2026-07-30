package feedback

import (
	"net/http"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	service *Service
}

type submitRequest struct {
	Category     string `json:"category"`
	Rating       int    `json:"rating"`
	Details      string `json:"details"`
	ContactEmail string `json:"contactEmail"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("feedback"))
		return
	}

	var req submitRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	result, err := h.service.Submit(r.Context(), SubmitInput{
		UserID:       user.ID,
		UserEmail:    user.Email,
		Category:     req.Category,
		Rating:       req.Rating,
		Details:      req.Details,
		ContactEmail: req.ContactEmail,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "feedback submitted", result)
}
