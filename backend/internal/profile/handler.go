package profile

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/school"
	"soulink-backend/internal/store"
)

type Handler struct {
	service *Service
}

type updateProfileRequest struct {
	Nickname     string `json:"nickname"`
	Department   string `json:"department"`
	Grade        string `json:"grade"`
	ContactType  string `json:"contactType"`
	ContactValue string `json:"contactValue"`
}

type upsertContactRequest struct {
	Value string `json:"value"`
}

type profileResponse struct {
	ID           int64             `json:"id,omitempty"`
	Email        string            `json:"email,omitempty"`
	Nickname     string            `json:"nickname"`
	SchoolName   string            `json:"schoolName"`
	Department   string            `json:"department"`
	Grade        string            `json:"grade"`
	ContactType  string            `json:"contactType"`
	ContactValue string            `json:"contactValue"`
	Contacts     []contactResponse `json:"contacts"`
	CreatedAt    string            `json:"createdAt,omitempty"`
}

type publicProfileResponse struct {
	ID         int64  `json:"id"`
	Nickname   string `json:"nickname"`
	SchoolName string `json:"schoolName"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
	CreatedAt  string `json:"createdAt"`
}

type contactResponse struct {
	Type      string `json:"type"`
	Label     string `json:"label"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updatedAt"`
	Readonly  bool   `json:"readonly,omitempty"`
}

// NewHandler 创建个人资料模块的 HTTP 处理器。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetProfile 返回当前登录用户的个人资料。
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("profile"))
		return
	}

	details, err := h.service.Get(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.OK(w, toProfileResponse(details, true))
}

func (h *Handler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("profile"))
		return
	}

	userID, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil || userID <= 0 {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid params"))
		return
	}

	details, err := h.service.GetPublic(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.OK(w, toPublicProfileResponse(details))
}

// UpdateProfile 处理个人资料更新请求。
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("profile"))
		return
	}

	var req updateProfileRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	updated, err := h.service.Update(r.Context(), user.ID, UpdateProfileInput{
		Nickname:     req.Nickname,
		Department:   req.Department,
		Grade:        req.Grade,
		ContactType:  req.ContactType,
		ContactValue: req.ContactValue,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "profile updated", toProfileResponse(updated, false))
}

// GetContacts 返回当前用户保存的联系方式列表；如果用户没有设置联系方式，则返回注册邮箱作为只读默认联系方式。
func (h *Handler) GetContacts(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("profile"))
		return
	}

	details, err := h.service.Get(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.OK(w, map[string]any{"contacts": contactsForResponse(details)})
}

// UpsertContact 保存或覆盖指定渠道的联系方式。
func (h *Handler) UpsertContact(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("profile"))
		return
	}

	var req upsertContactRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	updated, err := h.service.UpsertContact(r.Context(), user.ID, UpsertContactInput{
		ContactType:  chi.URLParam(r, "contactType"),
		ContactValue: req.Value,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "contact saved", map[string]any{"contacts": contactsForResponse(updated)})
}

// DeleteContact 删除指定渠道的联系方式。
func (h *Handler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("profile"))
		return
	}

	updated, err := h.service.DeleteContact(r.Context(), user.ID, chi.URLParam(r, "contactType"))
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "contact deleted", map[string]any{"contacts": contactsForResponse(updated)})
}

// toProfileResponse 将完整资料模型转换成接口响应。
func toProfileResponse(details Details, includeReadonly bool) profileResponse {
	response := profileResponse{
		Nickname:     details.User.Nickname,
		SchoolName:   schoolNameForEmail(details.User.Email),
		Department:   details.User.Department,
		Grade:        details.User.Grade,
		ContactType:  details.User.ContactType,
		ContactValue: details.User.ContactValue,
		Contacts:     contactsForResponse(details),
	}
	if includeReadonly {
		response.ID = details.User.ID
		response.Email = details.User.Email
		response.CreatedAt = details.User.CreatedAt.Format(time.RFC3339)
	}
	return response
}

func toPublicProfileResponse(details PublicDetails) publicProfileResponse {
	return publicProfileResponse{
		ID:         details.User.ID,
		Nickname:   details.User.Nickname,
		SchoolName: schoolNameForEmail(details.User.Email),
		Department: details.User.Department,
		Grade:      details.User.Grade,
		CreatedAt:  details.User.CreatedAt.Format(time.RFC3339),
	}
}

func contactsForResponse(details Details) []contactResponse {
	if len(details.Contacts) == 0 && details.User.Email != "" {
		return []contactResponse{{
			Type:      "email",
			Label:     meta.ContactTypeLabel("email"),
			Value:     details.User.Email,
			UpdatedAt: details.User.CreatedAt.Format(time.RFC3339),
			Readonly:  true,
		}}
	}
	return toContactResponses(details.Contacts)
}

func toContactResponses(contacts []store.ContactMethod) []contactResponse {
	responses := make([]contactResponse, 0, len(contacts))
	for _, contact := range contacts {
		responses = append(responses, contactResponse{
			Type:      contact.ContactType,
			Label:     meta.ContactTypeLabel(contact.ContactType),
			Value:     contact.ContactValue,
			UpdatedAt: contact.UpdatedAt.Format(time.RFC3339),
		})
	}
	return responses
}

func schoolNameForEmail(email string) string {
	displayName, ok := school.DisplayNameByEmail(email)
	if !ok {
		return ""
	}
	return displayName
}
