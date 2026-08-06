package safety

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

const maxPhotoEvidenceBytes = 2 * 1024 * 1024

type appealPhotoVerificationRequest struct {
	Reason string `json:"reason"`
}

type reviewPhotoVerificationRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

func (h *Handler) GetMyTrustProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := safetyCurrentUser(w, r)
	if !ok {
		return
	}
	profile, err := h.service.GetTrustProfile(r.Context(), user.ID, true)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, profile)
}

func (h *Handler) GetPublicTrustProfile(w http.ResponseWriter, r *http.Request) {
	if _, ok := safetyCurrentUser(w, r); !ok {
		return
	}
	userID, err := parseSafetyRouteID(r, "userId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	profile, err := h.service.GetTrustProfile(r.Context(), userID, false)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, profile)
}

func (h *Handler) SubmitPhotoVerification(w http.ResponseWriter, r *http.Request) {
	user, ok := safetyCurrentUser(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(maxPhotoEvidenceBytes + 1024); err != nil {
		httpapi.Error(w, invalidParams("照片上传失败，请选择 2MB 以内的 JPEG 或 PNG"))
		return
	}
	file, _, err := r.FormFile("photo")
	if err != nil {
		httpapi.Error(w, invalidParams("请选择用于人工审核的本人照片"))
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, maxPhotoEvidenceBytes+1))
	if err != nil || len(raw) == 0 || len(raw) > maxPhotoEvidenceBytes {
		httpapi.Error(w, invalidParams("照片大小需要在 2MB 以内"))
		return
	}
	mimeType, normalized, err := normalizePhotoEvidence(raw)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	profile, err := h.service.SubmitPhotoVerification(r.Context(), user.ID, mimeType, normalized)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "photo verification submitted", profile)
}

func (h *Handler) AppealPhotoVerification(w http.ResponseWriter, r *http.Request) {
	user, ok := safetyCurrentUser(w, r)
	if !ok {
		return
	}
	requestID, err := parseSafetyRouteID(r, "requestId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var request appealPhotoVerificationRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	profile, err := h.service.AppealPhotoVerification(r.Context(), user.ID, requestID, request.Reason)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, profile)
}

func (h *Handler) ListPhotoVerifications(w http.ResponseWriter, r *http.Request) {
	user, ok := safetyCurrentUser(w, r)
	if !ok {
		return
	}
	requests, err := h.service.ListPhotoVerifications(r.Context(), user.Email, r.URL.Query().Get("status"))
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, requests)
}

func (h *Handler) PhotoVerificationEvidence(w http.ResponseWriter, r *http.Request) {
	user, ok := safetyCurrentUser(w, r)
	if !ok {
		return
	}
	requestID, err := parseSafetyRouteID(r, "requestId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	mimeType, evidence, err := h.service.PhotoVerificationEvidence(r.Context(), user.Email, requestID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "no-store, private")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(evidence)
}

func (h *Handler) ReviewPhotoVerification(w http.ResponseWriter, r *http.Request) {
	user, ok := safetyCurrentUser(w, r)
	if !ok {
		return
	}
	requestID, err := parseSafetyRouteID(r, "requestId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var request reviewPhotoVerificationRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.ReviewPhotoVerification(
		r.Context(), user.ID, user.Email, requestID, request.Status, request.Notes,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) ListReviewEvents(w http.ResponseWriter, r *http.Request) {
	user, ok := safetyCurrentUser(w, r)
	if !ok {
		return
	}
	entityID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("entityId")), 10, 64)
	if err != nil || entityID <= 0 {
		httpapi.Error(w, invalidParams("审核日志参数不正确"))
		return
	}
	events, err := h.service.ReviewEvents(
		r.Context(), user.Email, strings.TrimSpace(r.URL.Query().Get("entityType")), entityID,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, events)
}

func normalizePhotoEvidence(raw []byte) (string, []byte, error) {
	contentType := http.DetectContentType(raw)
	if contentType != "image/jpeg" && contentType != "image/png" {
		return "", nil, invalidParams("照片仅支持 JPEG 或 PNG")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || config.Width <= 0 || config.Height <= 0 ||
		config.Width > 4096 || config.Height > 4096 || config.Width*config.Height > 16_000_000 {
		return "", nil, invalidParams("照片无法读取或尺寸过大")
	}
	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return "", nil, invalidParams("照片无法读取")
	}
	var output bytes.Buffer
	switch format {
	case "jpeg":
		contentType = "image/jpeg"
		err = jpeg.Encode(&output, decoded, &jpeg.Options{Quality: 88})
	case "png":
		contentType = "image/png"
		err = png.Encode(&output, decoded)
	default:
		return "", nil, invalidParams("照片仅支持 JPEG 或 PNG")
	}
	if err != nil || output.Len() == 0 || output.Len() > maxPhotoEvidenceBytes {
		return "", nil, invalidParams("处理后的照片超过 2MB，请压缩后重试")
	}
	return contentType, output.Bytes(), nil
}

func safetyCurrentUser(w http.ResponseWriter, r *http.Request) (user struct {
	ID    int64
	Email string
}, ok bool) {
	current, found := auth.CurrentUser(r)
	if !found {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return user, false
	}
	user.ID = current.ID
	user.Email = current.Email
	return user, true
}

func parseSafetyRouteID(r *http.Request, key string) (int64, error) {
	value, err := strconv.ParseInt(chi.URLParam(r, key), 10, 64)
	if err != nil || value <= 0 {
		return 0, invalidParams("无效的资源编号")
	}
	return value, nil
}
