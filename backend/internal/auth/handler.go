package auth

import (
	"net/http"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/school"
	"soulink-backend/internal/store"
)

type Handler struct {
	service *Service
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type sendCodeRequest struct {
	Email string `json:"email"`
}

type verifyCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type registerRequest struct {
	Email      string `json:"email"`
	Code       string `json:"code"`
	Password   string `json:"password"`
	Nickname   string `json:"nickname"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

type changePasswordRequest struct {
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

type deleteAccountRequest struct {
	Code string `json:"code"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

type userResponse struct {
	ID                        int64  `json:"id"`
	Email                     string `json:"email"`
	Nickname                  string `json:"nickname"`
	SchoolName                string `json:"schoolName"`
	Department                string `json:"department"`
	Grade                     string `json:"grade"`
	HasCompletedQuestionnaire bool   `json:"hasCompletedQuestionnaire"`
}

type meResponse struct {
	ID                        int64  `json:"id"`
	Email                     string `json:"email"`
	Nickname                  string `json:"nickname"`
	SchoolName                string `json:"schoolName"`
	Department                string `json:"department"`
	Grade                     string `json:"grade"`
	ContactType               string `json:"contactType"`
	ContactValue              string `json:"contactValue"`
	HasCompletedQuestionnaire bool   `json:"hasCompletedQuestionnaire"`
}

// NewHandler 创建认证模块的 HTTP 处理器。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login 处理登录请求。
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	token, user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.OK(w, authResponse{Token: token, User: toUserResponse(user)})
}

// Logout 处理退出登录请求。
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, ok := CurrentSession(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	if err := h.service.Logout(r.Context(), session.ID); err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "logout success", nil)
}

// Me 返回当前登录用户的基础信息。
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	httpapi.OK(w, meResponse{
		ID:                        user.ID,
		Email:                     user.Email,
		Nickname:                  user.Nickname,
		SchoolName:                schoolNameForEmail(user.Email),
		Department:                user.Department,
		Grade:                     user.Grade,
		ContactType:               user.ContactType,
		ContactValue:              user.ContactValue,
		HasCompletedQuestionnaire: user.HasCompletedQuestionnaire,
	})
}

func toUserResponse(user store.User) userResponse {
	return userResponse{
		ID:                        user.ID,
		Email:                     user.Email,
		Nickname:                  user.Nickname,
		SchoolName:                schoolNameForEmail(user.Email),
		Department:                user.Department,
		Grade:                     user.Grade,
		HasCompletedQuestionnaire: user.HasCompletedQuestionnaire,
	}
}

func schoolNameForEmail(email string) string {
	displayName, ok := school.DisplayNameByEmail(email)
	if !ok {
		return ""
	}
	return displayName
}

// SendRegisterCode 处理注册验证码发送请求。
func (h *Handler) SendRegisterCode(w http.ResponseWriter, r *http.Request) {
	var req sendCodeRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	expireSeconds, err := h.service.SendRegisterCode(r.Context(), req.Email)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "verification code sent", map[string]int{
		"expireSeconds":   expireSeconds,
		"cooldownSeconds": h.service.VerificationCodeCooldownSeconds(),
	})
}

// VerifyRegisterCode 处理注册验证码校验请求。
func (h *Handler) VerifyRegisterCode(w http.ResponseWriter, r *http.Request) {
	var req verifyCodeRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	if err := h.service.VerifyRegisterCode(r.Context(), req.Email, req.Code); err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "code verified", map[string]bool{"verified": true})
}

// Register 处理注册并自动登录请求。
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	token, user, err := h.service.Register(r.Context(), RegisterInput{
		Email:      req.Email,
		Code:       req.Code,
		Password:   req.Password,
		Nickname:   req.Nickname,
		Department: req.Department,
		Grade:      req.Grade,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "register success", authResponse{Token: token, User: toUserResponse(user)})
}

// SendForgotPasswordCode 处理找回密码验证码发送请求。
func (h *Handler) SendForgotPasswordCode(w http.ResponseWriter, r *http.Request) {
	var req sendCodeRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	expireSeconds, err := h.service.SendForgotPasswordCode(r.Context(), req.Email)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "verification code sent", map[string]int{
		"expireSeconds":   expireSeconds,
		"cooldownSeconds": h.service.VerificationCodeCooldownSeconds(),
	})
}

// ResetForgotPassword 处理忘记密码重置请求。
func (h *Handler) ResetForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	if err := h.service.ResetForgotPassword(r.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "password reset success", nil)
}

// SendChangePasswordCode 处理登录态下的修改密码验证码发送请求。
func (h *Handler) SendChangePasswordCode(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	expireSeconds, err := h.service.SendChangePasswordCode(r.Context(), user.Email)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "verification code sent", map[string]any{
		"email":           user.Email,
		"expireSeconds":   expireSeconds,
		"cooldownSeconds": h.service.VerificationCodeCooldownSeconds(),
	})
}

// ChangePassword 处理登录态下的修改密码请求。
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	var req changePasswordRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	if err := h.service.ChangePassword(r.Context(), user.ID, user.Email, req.Code, req.NewPassword); err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "password updated", nil)
}

// SendDeleteAccountCode 处理注销账号验证码发送请求。
func (h *Handler) SendDeleteAccountCode(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	expireSeconds, err := h.service.SendDeleteAccountCode(r.Context(), user.Email)
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "verification code sent", map[string]any{
		"email":           user.Email,
		"expireSeconds":   expireSeconds,
		"cooldownSeconds": h.service.VerificationCodeCooldownSeconds(),
	})
}

// DeleteAccount 处理账号注销请求。
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}

	var req deleteAccountRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	if err := h.service.DeleteAccount(r.Context(), user.ID, user.Email, req.Code); err != nil {
		httpapi.Error(w, err)
		return
	}

	httpapi.Respond(w, http.StatusOK, "account deleted", nil)
}
