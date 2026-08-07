package playground

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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

type createPlazaGroupInviteRequest struct {
	InviteType     string `json:"inviteType"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	MinMembers     int    `json:"minMembers"`
	MaxMembers     int    `json:"maxMembers"`
	RevealStrategy string `json:"revealStrategy"`
	Deadline       string `json:"deadline"`
	GroupExpiresAt string `json:"groupExpiresAt"`
}

type respondPlazaGroupApplicationRequest struct {
	Status string `json:"status"`
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

func (h *Handler) ListPostPage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}

	limit := defaultPostPageSize
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 || parsed > maxPostPageSize {
			httpapi.Error(w, invalidParams())
			return
		}
		limit = parsed
	}
	page, err := h.service.ListPostPage(r.Context(), user.ID, ListPostsInput{
		Topic:  strings.TrimSpace(r.URL.Query().Get("topic")),
		Query:  strings.TrimSpace(r.URL.Query().Get("q")),
		Sort:   strings.TrimSpace(r.URL.Query().Get("sort")),
		Cursor: strings.TrimSpace(r.URL.Query().Get("cursor")),
		Limit:  limit,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, page)
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

func (h *Handler) FavoritePost(w http.ResponseWriter, r *http.Request) {
	h.setPostFavorite(w, r, true)
}

func (h *Handler) UnfavoritePost(w http.ResponseWriter, r *http.Request) {
	h.setPostFavorite(w, r, false)
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

func (h *Handler) CreateGroupInvite(w http.ResponseWriter, r *http.Request) {
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
	var req createPlazaGroupInviteRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	deadline, err := parseOptionalRFC3339(req.Deadline)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	groupExpiresAt, err := parseOptionalRFC3339(req.GroupExpiresAt)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	invite, err := h.service.CreatePlazaGroupInvite(r.Context(), user.ID, CreatePlazaGroupInviteInput{
		PostID: postID, InviteType: req.InviteType,
		Title: req.Title, Summary: req.Summary,
		MinMembers: req.MinMembers, MaxMembers: req.MaxMembers,
		RevealStrategy: req.RevealStrategy,
		Deadline:       deadline, GroupExpiresAt: groupExpiresAt,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "plaza group invite created", invite)
}

func (h *Handler) ListMyGroupInvites(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}
	invites, err := h.service.ListMyPlazaGroupInvites(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, invites)
}

func (h *Handler) GetGroupInvite(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}
	inviteID, err := parseRouteID(r, "inviteId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	invite, err := h.service.GetPlazaGroupInvite(r.Context(), user.ID, inviteID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, invite)
}

func (h *Handler) ApplyToGroupInvite(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}
	inviteID, err := parseRouteID(r, "inviteId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	invite, err := h.service.ApplyToPlazaGroup(r.Context(), user.ID, inviteID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "plaza group application created", invite)
}

func (h *Handler) RespondToGroupApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}
	inviteID, err := parseRouteID(r, "inviteId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	applicationID, err := parseRouteID(r, "applicationId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var req respondPlazaGroupApplicationRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, err)
		return
	}
	invite, err := h.service.RespondToPlazaGroupApplication(
		r.Context(), user.ID, inviteID, applicationID, req.Status,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, invite)
}

func (h *Handler) FormGroupInvite(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}
	inviteID, err := parseRouteID(r, "inviteId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	invite, err := h.service.FormPlazaGroup(r.Context(), user.ID, inviteID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, invite)
}

func (h *Handler) KeepGroupInvite(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "unauthorized"))
		return
	}
	inviteID, err := parseRouteID(r, "inviteId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	invite, err := h.service.KeepPlazaGroup(r.Context(), user.ID, inviteID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, invite)
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

func (h *Handler) setPostFavorite(w http.ResponseWriter, r *http.Request, favorited bool) {
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

	result, err := h.service.SetPostFavorite(r.Context(), user.ID, postID, favorited)
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

func parseOptionalRFC3339(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid time")
	}
	return parsed, nil
}
