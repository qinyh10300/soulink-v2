package playground

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/analytics"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	maxPostContentLength    = 2000
	maxCommentContentLength = 1000
	maxPostImageCount       = 9
	maxImageURLLength       = 2048
	maxSearchQueryLength    = 80
	defaultPostPageSize     = 12
	maxPostPageSize         = 30
	maxPostCursorLength     = 512
)

var allowedTopics = map[string]struct{}{
	"campus":     {},
	"partner":    {},
	"study":      {},
	"emotion":    {},
	"secondhand": {},
}

type Service struct {
	store    *store.Store
	recorder analytics.EventRecorder
	nowFunc  func() time.Time
}

type CreatePostInput struct {
	Topic     string
	Content   string
	ImageURLs []string
}

type ListPostsInput struct {
	Topic  string
	Query  string
	Sort   string
	Cursor string
	Limit  int
}

type CreateCommentInput struct {
	PostID          int64
	Content         string
	ParentCommentID int64
}

type AuthorView struct {
	Label   string `json:"label"`
	IsMe    bool   `json:"isMe"`
	IsOwner bool   `json:"isOwner"`
}

type CommentView struct {
	ID              int64         `json:"id"`
	PostID          int64         `json:"postId"`
	Author          AuthorView    `json:"author"`
	ParentCommentID *int64        `json:"parentCommentId,omitempty"`
	Content         string        `json:"content"`
	LikeCount       int           `json:"likeCount"`
	ReplyCount      int           `json:"replyCount"`
	Floor           int           `json:"floor"`
	LikedByMe       bool          `json:"likedByMe"`
	CreatedAt       string        `json:"createdAt"`
	Replies         []CommentView `json:"replies"`
}

type PostView struct {
	ID           int64                    `json:"id"`
	Author       AuthorView               `json:"author"`
	Topic        string                   `json:"topic"`
	Content      string                   `json:"content"`
	ImageURLs    []string                 `json:"imageUrls"`
	LikeCount    int                      `json:"likeCount"`
	CommentCount int                      `json:"commentCount"`
	LikedByMe    bool                     `json:"likedByMe"`
	CreatedAt    string                   `json:"createdAt"`
	Comments     []CommentView            `json:"comments"`
	GroupInvite  *PlazaGroupInviteSummary `json:"groupInvite,omitempty"`
}

type PostPage struct {
	Items      []PostView `json:"items"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

type encodedPostCursor struct {
	Version   int    `json:"v"`
	Scope     string `json:"s"`
	Score     int    `json:"o"`
	CreatedAt string `json:"t"`
	ID        int64  `json:"i"`
}

type LikeResult struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"likeCount"`
}

func NewService(dataStore *store.Store) *Service {
	return &Service{store: dataStore, nowFunc: time.Now}
}

func (s *Service) WithEventRecorder(recorder analytics.EventRecorder) *Service {
	s.recorder = recorder
	return s
}

func (s *Service) ListPosts(ctx context.Context, viewerID int64, input ListPostsInput) ([]PostView, error) {
	input.Cursor = ""
	input.Limit = 50
	page, err := s.listPostPage(ctx, viewerID, input, 50)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s *Service) ListPostPage(ctx context.Context, viewerID int64, input ListPostsInput) (PostPage, error) {
	return s.listPostPage(ctx, viewerID, input, maxPostPageSize)
}

func (s *Service) listPostPage(ctx context.Context, viewerID int64, input ListPostsInput, maxLimit int) (PostPage, error) {
	if err := s.ensureReady(); err != nil {
		return PostPage{}, err
	}
	topic := strings.TrimSpace(input.Topic)
	if topic != "" {
		if _, ok := allowedTopics[topic]; !ok {
			return PostPage{}, invalidParams()
		}
	}
	query := strings.TrimSpace(input.Query)
	if utf8.RuneCountInString(query) > maxSearchQueryLength {
		return PostPage{}, invalidParams()
	}
	sortMode := strings.TrimSpace(input.Sort)
	if sortMode == "" {
		sortMode = "latest"
	}
	if sortMode != "latest" && sortMode != "hot" {
		return PostPage{}, invalidParams()
	}
	limit := input.Limit
	if limit == 0 {
		limit = defaultPostPageSize
	}
	if limit < 1 || limit > maxLimit {
		return PostPage{}, invalidParams()
	}

	scope := postCursorScope(topic, query, sortMode)
	var storeCursor *store.PlaygroundPostCursor
	if rawCursor := strings.TrimSpace(input.Cursor); rawCursor != "" {
		decoded, err := decodePostCursor(rawCursor, scope)
		if err != nil {
			return PostPage{}, invalidParams()
		}
		storeCursor = &decoded
	}
	if err := s.expirePlazaGroups(ctx); err != nil {
		return PostPage{}, err
	}

	posts, err := s.store.ListPlaygroundPosts(ctx, nil, viewerID, store.PlaygroundPostFilter{
		Topic:  topic,
		Query:  query,
		Sort:   sortMode,
		Limit:  limit + 1,
		Cursor: storeCursor,
	})
	if err != nil {
		return PostPage{}, httpapi.InternalError(err)
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	views := make([]PostView, 0, len(posts))
	for _, post := range posts {
		comments, err := s.store.ListPlaygroundComments(ctx, nil, viewerID, post.ID)
		if err != nil {
			return PostPage{}, httpapi.InternalError(err)
		}
		view := toPostView(post, buildCommentTree(post.Author.UserID, comments, viewerID), viewerID)
		invite, err := s.store.GetPlazaGroupInviteByPost(ctx, nil, post.ID)
		if err == nil {
			applications, listErr := s.store.ListPlazaGroupApplications(ctx, nil, invite.ID)
			if listErr != nil {
				return PostPage{}, httpapi.InternalError(listErr)
			}
			summary := plazaGroupInviteSummary(invite, applications, comments, viewerID)
			view.GroupInvite = &summary
		} else if err != store.ErrNotFound {
			return PostPage{}, httpapi.InternalError(err)
		}
		views = append(views, view)
	}

	page := PostPage{Items: views}
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		page.NextCursor, err = encodePostCursor(encodedPostCursor{
			Version:   1,
			Scope:     scope,
			Score:     last.LikeCount + last.CommentCount*2,
			CreatedAt: last.CreatedAt.UTC().Format(time.RFC3339Nano),
			ID:        last.ID,
		})
		if err != nil {
			return PostPage{}, httpapi.InternalError(err)
		}
	}
	return page, nil
}

func postCursorScope(topic, query, sortMode string) string {
	sum := sha256.Sum256([]byte(sortMode + "\x00" + topic + "\x00" + query))
	return hex.EncodeToString(sum[:8])
}

func encodePostCursor(cursor encodedPostCursor) (string, error) {
	raw, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("encode playground cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodePostCursor(raw, expectedScope string) (store.PlaygroundPostCursor, error) {
	if len(raw) > maxPostCursorLength {
		return store.PlaygroundPostCursor{}, fmt.Errorf("playground cursor too long")
	}
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return store.PlaygroundPostCursor{}, fmt.Errorf("decode playground cursor: %w", err)
	}
	var cursor encodedPostCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return store.PlaygroundPostCursor{}, fmt.Errorf("parse playground cursor: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, cursor.CreatedAt)
	if err != nil || cursor.Version != 1 || cursor.Scope != expectedScope || cursor.ID <= 0 || cursor.Score < 0 {
		return store.PlaygroundPostCursor{}, fmt.Errorf("invalid playground cursor")
	}
	return store.PlaygroundPostCursor{Score: cursor.Score, CreatedAt: createdAt, ID: cursor.ID}, nil
}

func (s *Service) CreatePost(ctx context.Context, authorID int64, input CreatePostInput) (PostView, error) {
	if err := s.ensureReady(); err != nil {
		return PostView{}, err
	}

	topic, content, imageURLs, err := normalizePostInput(input)
	if err != nil {
		return PostView{}, err
	}
	if err := validatePlazaContent(content); err != nil {
		return PostView{}, err
	}

	var post store.PlaygroundPost
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		postID, err := s.store.CreatePlaygroundPost(ctx, tx, store.CreatePlaygroundPostParams{
			AuthorID:  authorID,
			Topic:     topic,
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

	return toPostView(post, nil, authorID), nil
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
	if err := validatePlazaContent(content); err != nil {
		return CommentView{}, err
	}

	var comment store.PlaygroundComment
	var post store.PlaygroundPost
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		loadedPost, err := s.store.GetPlaygroundPost(ctx, tx, authorID, input.PostID)
		if err != nil {
			return mapPlaygroundStoreError(err)
		}
		post = loadedPost

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

	comments, err := s.store.ListPlaygroundComments(ctx, nil, authorID, input.PostID)
	if err != nil {
		return CommentView{}, httpapi.InternalError(err)
	}
	return toCommentView(comment, newAnonymousContext(post.Author.UserID, comments), authorID), nil
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

func normalizePostInput(input CreatePostInput) (string, string, []string, error) {
	topic := strings.TrimSpace(input.Topic)
	if topic == "" {
		topic = "campus"
	}
	if _, ok := allowedTopics[topic]; !ok {
		return "", "", nil, invalidParams()
	}
	content := strings.TrimSpace(input.Content)
	if utf8.RuneCountInString(content) > maxPostContentLength {
		return "", "", nil, invalidParams()
	}

	imageURLs := make([]string, 0, len(input.ImageURLs))
	seen := make(map[string]struct{}, len(input.ImageURLs))
	for _, raw := range input.ImageURLs {
		imageURL := strings.TrimSpace(raw)
		if imageURL == "" {
			continue
		}
		if utf8.RuneCountInString(imageURL) > maxImageURLLength {
			return "", "", nil, invalidParams()
		}
		parsed, err := url.ParseRequestURI(imageURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return "", "", nil, invalidParams()
		}
		if _, ok := seen[imageURL]; ok {
			continue
		}
		seen[imageURL] = struct{}{}
		imageURLs = append(imageURLs, imageURL)
	}

	if content == "" && len(imageURLs) == 0 {
		return "", "", nil, invalidParams()
	}
	if len(imageURLs) > maxPostImageCount {
		return "", "", nil, invalidParams()
	}

	return topic, content, imageURLs, nil
}

type anonymousContext struct {
	labelsByUser   map[int64]string
	floorByComment map[int64]int
	postAuthorID   int64
}

func newAnonymousContext(postAuthorID int64, comments []store.PlaygroundComment) anonymousContext {
	context := anonymousContext{
		labelsByUser:   map[int64]string{postAuthorID: "洞主"},
		floorByComment: make(map[int64]int, len(comments)),
		postAuthorID:   postAuthorID,
	}
	nextAlias := 0
	for index, comment := range comments {
		context.floorByComment[comment.ID] = index + 1
		if _, ok := context.labelsByUser[comment.Author.UserID]; ok {
			continue
		}
		context.labelsByUser[comment.Author.UserID] = anonymousLabel(nextAlias)
		nextAlias++
	}
	return context
}

func anonymousLabel(index int) string {
	if index < 26 {
		return fmt.Sprintf("同学 %c", rune('A'+index))
	}
	return fmt.Sprintf("同学 %d", index+1)
}

func buildCommentTree(postAuthorID int64, comments []store.PlaygroundComment, viewerID int64) []CommentView {
	anonymous := newAnonymousContext(postAuthorID, comments)
	viewsByID := make(map[int64]*CommentView, len(comments))
	roots := make([]CommentView, 0)

	for _, comment := range comments {
		view := toCommentView(comment, anonymous, viewerID)
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

func toPostView(post store.PlaygroundPost, comments []CommentView, viewerID int64) PostView {
	if comments == nil {
		comments = []CommentView{}
	}
	if post.ImageURLs == nil {
		post.ImageURLs = []string{}
	}
	return PostView{
		ID:           post.ID,
		Author:       AuthorView{Label: "洞主", IsMe: post.Author.UserID == viewerID, IsOwner: true},
		Topic:        post.Topic,
		Content:      post.Content,
		ImageURLs:    post.ImageURLs,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		LikedByMe:    post.LikedByMe,
		CreatedAt:    post.CreatedAt.Format(time.RFC3339),
		Comments:     comments,
	}
}

func toCommentView(comment store.PlaygroundComment, anonymous anonymousContext, viewerID int64) CommentView {
	var parentID *int64
	if comment.ParentCommentID.Valid {
		value := comment.ParentCommentID.Int64
		parentID = &value
	}
	return CommentView{
		ID:     comment.ID,
		PostID: comment.PostID,
		Author: AuthorView{
			Label:   anonymous.labelsByUser[comment.Author.UserID],
			IsMe:    comment.Author.UserID == viewerID,
			IsOwner: comment.Author.UserID == anonymous.postAuthorID,
		},
		ParentCommentID: parentID,
		Content:         comment.Content,
		LikeCount:       comment.LikeCount,
		ReplyCount:      comment.ReplyCount,
		Floor:           anonymous.floorByComment[comment.ID],
		LikedByMe:       comment.LikedByMe,
		CreatedAt:       comment.CreatedAt.Format(time.RFC3339),
		Replies:         []CommentView{},
	}
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
