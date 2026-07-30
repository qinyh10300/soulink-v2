package questionnaire

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	service *Service
}

type saveAnswerRequest struct {
	Value      json.RawMessage `json:"value"`
	Importance *bool           `json:"importance,omitempty"`
}

type batchAnswerRequestItem struct {
	QuestionID string          `json:"questionId"`
	Value      json.RawMessage `json:"value"`
	Importance *bool           `json:"importance,omitempty"`
}

type saveAnswersRequest struct {
	Answers []batchAnswerRequestItem `json:"answers"`
	Action  string                   `json:"action"`
}

type updateParticipationRequest struct {
	ParticipatingInMatch bool `json:"participatingInMatch"`
}

func NewHandler(service ...*Service) *Handler {
	var svc *Service
	if len(service) > 0 {
		svc = service[0]
	}
	return &Handler{service: svc}
}

func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("questionnaire/current"))
		return
	}

	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	result, err := h.service.GetCurrent(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("questionnaire/current/progress"))
		return
	}

	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	result, err := h.service.GetProgress(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) GetPortrait(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("questionnaire/current/portrait"))
		return
	}

	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	result, err := h.service.GetPortrait(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) SaveAnswer(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("questionnaire/current/answers/{questionId}"))
		return
	}

	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	var req saveAnswerRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	result, err := h.service.SaveAnswer(r.Context(), user.ID, SaveAnswerInput{
		QuestionID: chi.URLParam(r, "questionId"),
		Value:      req.Value,
		Importance: req.Importance,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "answer saved", result)
}

func (h *Handler) SaveAnswers(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("questionnaire/current/answers"))
		return
	}

	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	var req saveAnswersRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	answers := make([]SaveAnswerInput, 0, len(req.Answers))
	for _, item := range req.Answers {
		answers = append(answers, SaveAnswerInput{
			QuestionID: item.QuestionID,
			Value:      item.Value,
			Importance: item.Importance,
		})
	}

	result, err := h.service.SaveAnswers(r.Context(), user.ID, SaveAnswersInput{
		Answers: answers,
		Action:  req.Action,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "answers saved", result)
}

func (h *Handler) UpdateParticipation(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("questionnaire/current/participation"))
		return
	}

	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	var req updateParticipationRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	result, err := h.service.UpdateParticipation(r.Context(), user.ID, req.ParticipatingInMatch)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "match participation updated", result)
}
