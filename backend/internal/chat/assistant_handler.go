package chat

import (
	"net/http"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type updateAssistantPreferencesRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *Handler) GetAssistantPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := assistantCurrentUser(w, r)
	if !ok {
		return
	}
	preferences, err := h.service.GetAssistantPreferences(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, preferences)
}

func (h *Handler) UpdateAssistantPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := assistantCurrentUser(w, r)
	if !ok {
		return
	}
	var request updateAssistantPreferencesRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	preferences, err := h.service.SetAssistantEnabled(r.Context(), user.ID, request.Enabled)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, preferences)
}

func (h *Handler) ResetAssistantMemory(w http.ResponseWriter, r *http.Request) {
	user, ok := assistantCurrentUser(w, r)
	if !ok {
		return
	}
	preferences, err := h.service.ResetAssistantMemory(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, preferences)
}

func (h *Handler) GetDirectAssistant(w http.ResponseWriter, r *http.Request) {
	h.getAssistantWorkspace(w, r, assistantContextDirect, "conversationId")
}

func (h *Handler) GenerateDirectAssistantStarters(w http.ResponseWriter, r *http.Request) {
	h.generateAssistantArtifact(w, r, assistantContextDirect, "conversationId", "starters")
}

func (h *Handler) GenerateDirectAssistantSummary(w http.ResponseWriter, r *http.Request) {
	h.generateAssistantArtifact(w, r, assistantContextDirect, "conversationId", "summary")
}

func (h *Handler) GenerateDirectAssistantActions(w http.ResponseWriter, r *http.Request) {
	h.generateAssistantArtifact(w, r, assistantContextDirect, "conversationId", "actions")
}

func (h *Handler) DeleteDirectAssistantArtifact(w http.ResponseWriter, r *http.Request) {
	h.deleteAssistantArtifact(w, r, assistantContextDirect, "conversationId")
}

func (h *Handler) GetGroupAssistant(w http.ResponseWriter, r *http.Request) {
	h.getAssistantWorkspace(w, r, assistantContextGroup, "groupId")
}

func (h *Handler) GenerateGroupAssistantSummary(w http.ResponseWriter, r *http.Request) {
	h.generateAssistantArtifact(w, r, assistantContextGroup, "groupId", "summary")
}

func (h *Handler) GenerateGroupAssistantActions(w http.ResponseWriter, r *http.Request) {
	h.generateAssistantArtifact(w, r, assistantContextGroup, "groupId", "actions")
}

func (h *Handler) DeleteGroupAssistantArtifact(w http.ResponseWriter, r *http.Request) {
	h.deleteAssistantArtifact(w, r, assistantContextGroup, "groupId")
}

func (h *Handler) getAssistantWorkspace(w http.ResponseWriter, r *http.Request, contextType, routeKey string) {
	user, ok := assistantCurrentUser(w, r)
	if !ok {
		return
	}
	contextID, err := parseChatRouteID(r, routeKey)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	workspace, err := h.service.GetAssistantWorkspace(r.Context(), user.ID, contextType, contextID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, workspace)
}

func (h *Handler) generateAssistantArtifact(
	w http.ResponseWriter,
	r *http.Request,
	contextType, routeKey, artifactType string,
) {
	user, ok := assistantCurrentUser(w, r)
	if !ok {
		return
	}
	contextID, err := parseChatRouteID(r, routeKey)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	artifact, err := h.service.GenerateAssistantArtifact(
		r.Context(), user.ID, contextType, contextID, artifactType,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "assistant draft created", artifact)
}

func (h *Handler) deleteAssistantArtifact(w http.ResponseWriter, r *http.Request, contextType, routeKey string) {
	user, ok := assistantCurrentUser(w, r)
	if !ok {
		return
	}
	contextID, err := parseChatRouteID(r, routeKey)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	artifactID, err := parseChatRouteID(r, "artifactId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	if err := h.service.DeleteAssistantArtifact(
		r.Context(), user.ID, contextType, contextID, artifactID,
	); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}

func assistantCurrentUser(w http.ResponseWriter, r *http.Request) (store.User, bool) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return store.User{}, false
	}
	return user, true
}
