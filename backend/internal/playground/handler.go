package playground

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type Handler struct {
	service *Service
}

type createPostRequest struct {
	Topic     string   `json:"topic"`
	Content   string   `json:"content"`
	ImageURLs []string `json:"imageUrls"`
}

type createCommentRequest struct {
	Content         string `json:"content"`
	ParentCommentID int64  `json:"parentCommentId"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	posts, err := h.service.ListPosts(r.Context(), user.ID, ListPostsInput{
		Topic: strings.TrimSpace(r.URL.Query().Get("topic")),
		Query: strings.TrimSpace(r.URL.Query().Get("q")),
		Sort:  strings.TrimSpace(r.URL.Query().Get("sort")),
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, posts)
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	var req createPostRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	post, err := h.service.CreatePost(r.Context(), user.ID, CreatePostInput{
		Topic:     req.Topic,
		Content:   req.Content,
		ImageURLs: req.ImageURLs,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "post created", post)
}

func (h *Handler) LikePost(w http.ResponseWriter, r *http.Request) {
	h.setPostLike(w, r, true)
}

func (h *Handler) UnlikePost(w http.ResponseWriter, r *http.Request) {
	h.setPostLike(w, r, false)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	postID, err := parseRouteID(r, "postId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	var req createCommentRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}

	comment, err := h.service.CreateComment(r.Context(), user.ID, CreateCommentInput{
		PostID:          postID,
		Content:         req.Content,
		ParentCommentID: req.ParentCommentID,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "comment created", comment)
}

func (h *Handler) LikeComment(w http.ResponseWriter, r *http.Request) {
	h.setCommentLike(w, r, true)
}

func (h *Handler) UnlikeComment(w http.ResponseWriter, r *http.Request) {
	h.setCommentLike(w, r, false)
}

func (h *Handler) setPostLike(w http.ResponseWriter, r *http.Request, liked bool) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	postID, err := parseRouteID(r, "postId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	result, err := h.service.SetPostLike(r.Context(), user.ID, postID, liked)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) setCommentLike(w http.ResponseWriter, r *http.Request, liked bool) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	commentID, err := parseRouteID(r, "commentId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}

	result, err := h.service.SetCommentLike(r.Context(), user.ID, commentID, liked)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func parseRouteID(r *http.Request, key string) (int64, error) {
	value, err := strconv.ParseInt(chi.URLParam(r, key), 10, 64)
	if err != nil || value <= 0 {
		return 0, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid params")
	}
	return value, nil
}
