package playground

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/school"
	"soulink-backend/internal/store"
)

const (
	maxPostContentLength    = 2000
	maxCommentContentLength = 1000
	maxPostImageCount       = 9
	maxImageURLLength       = 2048
)

type Service struct {
	store *store.Store
}

type CreatePostInput struct {
	Content   string
	ImageURLs []string
}

type CreateCommentInput struct {
	PostID          int64
	Content         string
	ParentCommentID int64
}

type AuthorView struct {
	UserID     int64  `json:"userId"`
	Nickname   string `json:"nickname"`
	SchoolName string `json:"schoolName"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
}

type CommentView struct {
	ID              int64         `json:"id"`
	PostID          int64         `json:"postId"`
	Author          AuthorView    `json:"author"`
	ParentCommentID *int64        `json:"parentCommentId,omitempty"`
	Content         string        `json:"content"`
	LikeCount       int           `json:"likeCount"`
	ReplyCount      int           `json:"replyCount"`
	LikedByMe       bool          `json:"likedByMe"`
	CreatedAt       string        `json:"createdAt"`
	Replies         []CommentView `json:"replies"`
}

type PostView struct {
	ID           int64         `json:"id"`
	Author       AuthorView    `json:"author"`
	Content      string        `json:"content"`
	ImageURLs    []string      `json:"imageUrls"`
	LikeCount    int           `json:"likeCount"`
	CommentCount int           `json:"commentCount"`
	LikedByMe    bool          `json:"likedByMe"`
	CreatedAt    string        `json:"createdAt"`
	Comments     []CommentView `json:"comments"`
}

type LikeResult struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"likeCount"`
}

func NewService(dataStore *store.Store) *Service {
	return &Service{store: dataStore}
}

func (s *Service) ListPosts(ctx context.Context, viewerID, authorID int64) ([]PostView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	posts, err := s.store.ListPlaygroundPosts(ctx, nil, viewerID, authorID, 50)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}

	views := make([]PostView, 0, len(posts))
	for _, post := range posts {
		comments, err := s.store.ListPlaygroundComments(ctx, nil, viewerID, post.ID)
		if err != nil {
			return nil, httpapi.InternalError(err)
		}
		views = append(views, toPostView(post, buildCommentTree(comments)))
	}
	return views, nil
}

func (s *Service) CreatePost(ctx context.Context, authorID int64, input CreatePostInput) (PostView, error) {
	if err := s.ensureReady(); err != nil {
		return PostView{}, err
	}

	content, imageURLs, err := normalizePostInput(input)
	if err != nil {
		return PostView{}, err
	}

	var post store.PlaygroundPost
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		postID, err := s.store.CreatePlaygroundPost(ctx, tx, store.CreatePlaygroundPostParams{
			AuthorID:  authorID,
			Content:   content,
			ImageURLs: imageURLs,
			Now:       time.Now(),
		})
		if err != nil {
			return httpapi.InternalError(err)
		}

		created, err := s.store.GetPlaygroundPost(ctx, tx, authorID, postID)
		if err != nil {
			return mapPlaygroundStoreError(err)
		}
		post = created
		return nil
	})
	if err != nil {
		return PostView{}, err
	}

	return toPostView(post, nil), nil
}

func (s *Service) SetPostLike(ctx context.Context, userID, postID int64, liked bool) (LikeResult, error) {
	if err := s.ensureReady(); err != nil {
		return LikeResult{}, err
	}

	var post store.PlaygroundPost
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := s.store.GetPlaygroundPost(ctx, tx, userID, postID); err != nil {
			return mapPlaygroundStoreError(err)
		}

		var changed bool
		var err error
		if liked {
			changed, err = s.store.CreatePlaygroundPostLike(ctx, tx, postID, userID, time.Now())
			if err == nil && changed {
				err = s.store.AddPlaygroundPostLikeCount(ctx, tx, postID, 1)
			}
		} else {
			changed, err = s.store.DeletePlaygroundPostLike(ctx, tx, postID, userID)
			if err == nil && changed {
				err = s.store.AddPlaygroundPostLikeCount(ctx, tx, postID, -1)
			}
		}
		if err != nil {
			return httpapi.InternalError(err)
		}

		updated, err := s.store.GetPlaygroundPost(ctx, tx, userID, postID)
		if err != nil {
			return mapPlaygroundStoreError(err)
		}
		post = updated
		return nil
	})
	if err != nil {
		return LikeResult{}, err
	}

	return LikeResult{Liked: post.LikedByMe, LikeCount: post.LikeCount}, nil
}

func (s *Service) CreateComment(ctx context.Context, authorID int64, input CreateCommentInput) (CommentView, error) {
	if err := s.ensureReady(); err != nil {
		return CommentView{}, err
	}
	if input.PostID <= 0 {
		return CommentView{}, invalidParams()
	}

	content := strings.TrimSpace(input.Content)
	if content == "" || utf8.RuneCountInString(content) > maxCommentContentLength {
		return CommentView{}, invalidParams()
	}

	var comment store.PlaygroundComment
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := s.store.GetPlaygroundPost(ctx, tx, authorID, input.PostID); err != nil {
			return mapPlaygroundStoreError(err)
		}

		var parentID sql.NullInt64
		if input.ParentCommentID > 0 {
			parent, err := s.store.GetPlaygroundComment(ctx, tx, authorID, input.ParentCommentID)
			if err != nil {
				return mapPlaygroundStoreError(err)
			}
			if parent.PostID != input.PostID {
				return invalidParams()
			}
			if parent.ParentCommentID.Valid {
				return invalidParams()
			}
			parentID = sql.NullInt64{Int64: input.ParentCommentID, Valid: true}
		}

		commentID, err := s.store.CreatePlaygroundComment(ctx, tx, store.CreatePlaygroundCommentParams{
			PostID:          input.PostID,
			AuthorID:        authorID,
			ParentCommentID: parentID,
			Content:         content,
			Now:             time.Now(),
		})
		if err != nil {
			return httpapi.InternalError(err)
		}

		if err := s.store.AddPlaygroundPostCommentCount(ctx, tx, input.PostID, 1); err != nil {
			return httpapi.InternalError(err)
		}
		if parentID.Valid {
			if err := s.store.AddPlaygroundCommentReplyCount(ctx, tx, parentID.Int64, 1); err != nil {
				return httpapi.InternalError(err)
			}
		}

		created, err := s.store.GetPlaygroundComment(ctx, tx, authorID, commentID)
		if err != nil {
			return mapPlaygroundStoreError(err)
		}
		comment = created
		return nil
	})
	if err != nil {
		return CommentView{}, err
	}

	return toCommentView(comment), nil
}

func (s *Service) SetCommentLike(ctx context.Context, userID, commentID int64, liked bool) (LikeResult, error) {
	if err := s.ensureReady(); err != nil {
		return LikeResult{}, err
	}

	var comment store.PlaygroundComment
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := s.store.GetPlaygroundComment(ctx, tx, userID, commentID); err != nil {
			return mapPlaygroundStoreError(err)
		}

		var changed bool
		var err error
		if liked {
			changed, err = s.store.CreatePlaygroundCommentLike(ctx, tx, commentID, userID, time.Now())
			if err == nil && changed {
				err = s.store.AddPlaygroundCommentLikeCount(ctx, tx, commentID, 1)
			}
		} else {
			changed, err = s.store.DeletePlaygroundCommentLike(ctx, tx, commentID, userID)
			if err == nil && changed {
				err = s.store.AddPlaygroundCommentLikeCount(ctx, tx, commentID, -1)
			}
		}
		if err != nil {
			return httpapi.InternalError(err)
		}

		updated, err := s.store.GetPlaygroundComment(ctx, tx, userID, commentID)
		if err != nil {
			return mapPlaygroundStoreError(err)
		}
		comment = updated
		return nil
	})
	if err != nil {
		return LikeResult{}, err
	}

	return LikeResult{Liked: comment.LikedByMe, LikeCount: comment.LikeCount}, nil
}

func (s *Service) ensureReady() error {
	if s == nil || s.store == nil {
		return httpapi.FeatureNotReady("playground")
	}
	return nil
}

func normalizePostInput(input CreatePostInput) (string, []string, error) {
	content := strings.TrimSpace(input.Content)
	if utf8.RuneCountInString(content) > maxPostContentLength {
		return "", nil, invalidParams()
	}

	imageURLs := make([]string, 0, len(input.ImageURLs))
	seen := make(map[string]struct{}, len(input.ImageURLs))
	for _, raw := range input.ImageURLs {
		imageURL := strings.TrimSpace(raw)
		if imageURL == "" {
			continue
		}
		if utf8.RuneCountInString(imageURL) > maxImageURLLength {
			return "", nil, invalidParams()
		}
		parsed, err := url.ParseRequestURI(imageURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return "", nil, invalidParams()
		}
		if _, ok := seen[imageURL]; ok {
			continue
		}
		seen[imageURL] = struct{}{}
		imageURLs = append(imageURLs, imageURL)
	}

	if content == "" && len(imageURLs) == 0 {
		return "", nil, invalidParams()
	}
	if len(imageURLs) > maxPostImageCount {
		return "", nil, invalidParams()
	}

	return content, imageURLs, nil
}

func buildCommentTree(comments []store.PlaygroundComment) []CommentView {
	viewsByID := make(map[int64]*CommentView, len(comments))
	roots := make([]CommentView, 0)

	for _, comment := range comments {
		view := toCommentView(comment)
		viewsByID[view.ID] = &view
		if !comment.ParentCommentID.Valid {
			roots = append(roots, view)
			continue
		}
		parent, ok := viewsByID[comment.ParentCommentID.Int64]
		if !ok {
			roots = append(roots, view)
			continue
		}
		parent.Replies = append(parent.Replies, view)
	}

	// Re-copy roots so each root includes replies appended through the map pointers.
	result := make([]CommentView, 0, len(roots))
	for _, root := range roots {
		if updated, ok := viewsByID[root.ID]; ok {
			result = append(result, *updated)
		} else {
			result = append(result, root)
		}
	}
	return result
}

func toPostView(post store.PlaygroundPost, comments []CommentView) PostView {
	if comments == nil {
		comments = []CommentView{}
	}
	if post.ImageURLs == nil {
		post.ImageURLs = []string{}
	}
	return PostView{
		ID:           post.ID,
		Author:       toAuthorView(post.Author),
		Content:      post.Content,
		ImageURLs:    post.ImageURLs,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		LikedByMe:    post.LikedByMe,
		CreatedAt:    post.CreatedAt.Format(time.RFC3339),
		Comments:     comments,
	}
}

func toCommentView(comment store.PlaygroundComment) CommentView {
	var parentID *int64
	if comment.ParentCommentID.Valid {
		value := comment.ParentCommentID.Int64
		parentID = &value
	}
	return CommentView{
		ID:              comment.ID,
		PostID:          comment.PostID,
		Author:          toAuthorView(comment.Author),
		ParentCommentID: parentID,
		Content:         comment.Content,
		LikeCount:       comment.LikeCount,
		ReplyCount:      comment.ReplyCount,
		LikedByMe:       comment.LikedByMe,
		CreatedAt:       comment.CreatedAt.Format(time.RFC3339),
		Replies:         []CommentView{},
	}
}

func toAuthorView(author store.PlaygroundAuthor) AuthorView {
	return AuthorView{
		UserID:     author.UserID,
		Nickname:   author.Nickname,
		SchoolName: schoolNameForEmail(author.Email),
		Department: author.Department,
		Grade:      author.Grade,
	}
}

func schoolNameForEmail(email string) string {
	displayName, ok := school.DisplayNameByEmail(email)
	if !ok {
		return ""
	}
	return displayName
}

func mapPlaygroundStoreError(err error) error {
	if err == nil {
		return nil
	}
	if err == store.ErrNotFound {
		return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "resource not found")
	}
	return httpapi.InternalError(err)
}

func invalidParams() error {
	return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid params")
}
