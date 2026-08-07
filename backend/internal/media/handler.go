package media

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UploadPlaza(w http.ResponseWriter, r *http.Request) {
	h.upload(w, r, false)
}

func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	h.upload(w, r, true)
}

func (h *Handler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	if err := h.service.DeleteAvatar(r.Context(), user.ID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, map[string]bool{"deleted": true})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	asset, err := h.service.GetReadable(r.Context(), user.ID, chi.URLParam(r, "assetKey"))
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	w.Header().Set("Content-Type", asset.MIMEType)
	w.Header().Set("Content-Length", strconv.Itoa(asset.ByteSize))
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(asset.Content)
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request, avatar bool) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadBytes+(1<<20))
	file, header, err := r.FormFile("file")
	if err != nil {
		httpapi.Error(w, invalidMedia("请选择要上传的图片"))
		return
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, MaxUploadBytes+1))
	if err != nil {
		httpapi.Error(w, httpapi.InternalError(fmt.Errorf("read media upload: %w", err)))
		return
	}
	input := UploadInput{Filename: header.Filename, Content: content}
	var view AssetView
	if avatar {
		view, err = h.service.UploadAvatar(r.Context(), user.ID, input)
	} else {
		view, err = h.service.UploadPlaza(r.Context(), user.ID, input)
	}
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "media uploaded", view)
}
