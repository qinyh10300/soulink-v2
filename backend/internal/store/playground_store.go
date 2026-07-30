package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type PlaygroundAuthor struct {
	UserID     int64
	Nickname   string
	Email      string
	Department string
	Grade      string
}

type PlaygroundPost struct {
	ID           int64
	Author       PlaygroundAuthor
	Content      string
	ImageURLs    []string
	LikeCount    int
	CommentCount int
	LikedByMe    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PlaygroundComment struct {
	ID              int64
	PostID          int64
	Author          PlaygroundAuthor
	ParentCommentID sql.NullInt64
	Content         string
	LikeCount       int
	ReplyCount      int
	LikedByMe       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreatePlaygroundPostParams struct {
	AuthorID  int64
	Content   string
	ImageURLs []string
	Now       time.Time
}

type CreatePlaygroundCommentParams struct {
	PostID          int64
	AuthorID        int64
	ParentCommentID sql.NullInt64
	Content         string
	Now             time.Time
}

func (s *Store) CreatePlaygroundPost(ctx context.Context, tx *sql.Tx, params CreatePlaygroundPostParams) (int64, error) {
	rawImageURLs, err := json.Marshal(params.ImageURLs)
	if err != nil {
		return 0, fmt.Errorf("marshal playground image urls: %w", err)
	}

	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO playground_posts (author_id, content, image_urls, created_at, updated_at)
		VALUES ($1, $2, $3::jsonb, $4, $4)
		RETURNING id
	`, params.AuthorID, params.Content, string(rawImageURLs), params.Now)

	var postID int64
	if err := row.Scan(&postID); err != nil {
		return 0, fmt.Errorf("create playground post: %w", err)
	}
	return postID, nil
}

func (s *Store) GetPlaygroundPost(ctx context.Context, tx *sql.Tx, viewerID, postID int64) (PlaygroundPost, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT p.id,
		       u.id, u.nickname, u.email, u.department, u.grade,
		       p.content, p.image_urls::text, p.like_count, p.comment_count,
		       EXISTS(
		         SELECT 1 FROM playground_post_likes l
		         WHERE l.post_id = p.id AND l.user_id = $1
		       ) AS liked_by_me,
		       p.created_at, p.updated_at
		FROM playground_posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.id = $2 AND p.deleted_at IS NULL AND u.deleted_at IS NULL
		LIMIT 1
	`, viewerID, postID)

	post, err := scanPlaygroundPost(row)
	if err != nil {
		return PlaygroundPost{}, err
	}
	return post, nil
}

func (s *Store) ListPlaygroundPosts(ctx context.Context, tx *sql.Tx, viewerID, authorID int64, limit int) ([]PlaygroundPost, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT p.id,
		       u.id, u.nickname, u.email, u.department, u.grade,
		       p.content, p.image_urls::text, p.like_count, p.comment_count,
		       EXISTS(
		         SELECT 1 FROM playground_post_likes l
		         WHERE l.post_id = p.id AND l.user_id = $1
		       ) AS liked_by_me,
		       p.created_at, p.updated_at
		FROM playground_posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.deleted_at IS NULL AND u.deleted_at IS NULL
	`
	args := []any{viewerID}
	if authorID > 0 {
		args = append(args, authorID)
		query += fmt.Sprintf(" AND p.author_id = $%d", len(args))
	}
	args = append(args, limit)
	query += fmt.Sprintf(" ORDER BY p.created_at DESC, p.id DESC LIMIT $%d", len(args))

	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list playground posts: %w", err)
	}
	defer rows.Close()

	posts := make([]PlaygroundPost, 0)
	for rows.Next() {
		post, err := scanPlaygroundPostRows(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate playground posts: %w", err)
	}
	return posts, nil
}

func (s *Store) CreatePlaygroundPostLike(ctx context.Context, tx *sql.Tx, postID, userID int64, now time.Time) (bool, error) {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO playground_post_likes (post_id, user_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (post_id, user_id) DO NOTHING
	`, postID, userID, now)
	if err != nil {
		return false, fmt.Errorf("create playground post like: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("playground post like rows affected: %w", err)
	}
	return rows > 0, nil
}

func (s *Store) DeletePlaygroundPostLike(ctx context.Context, tx *sql.Tx, postID, userID int64) (bool, error) {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM playground_post_likes
		WHERE post_id = $1 AND user_id = $2
	`, postID, userID)
	if err != nil {
		return false, fmt.Errorf("delete playground post like: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("playground post unlike rows affected: %w", err)
	}
	return rows > 0, nil
}

func (s *Store) AddPlaygroundPostLikeCount(ctx context.Context, tx *sql.Tx, postID int64, delta int) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE playground_posts
		SET like_count = GREATEST(0, like_count + $2)
		WHERE id = $1 AND deleted_at IS NULL
	`, postID, delta)
	if err != nil {
		return fmt.Errorf("update playground post like count: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) AddPlaygroundPostCommentCount(ctx context.Context, tx *sql.Tx, postID int64, delta int) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE playground_posts
		SET comment_count = GREATEST(0, comment_count + $2)
		WHERE id = $1 AND deleted_at IS NULL
	`, postID, delta)
	if err != nil {
		return fmt.Errorf("update playground post comment count: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) CreatePlaygroundComment(ctx context.Context, tx *sql.Tx, params CreatePlaygroundCommentParams) (int64, error) {
	var parentCommentID any
	if params.ParentCommentID.Valid {
		parentCommentID = params.ParentCommentID.Int64
	}

	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO playground_comments (
			post_id, author_id, parent_comment_id, content, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id
	`, params.PostID, params.AuthorID, parentCommentID, params.Content, params.Now)

	var commentID int64
	if err := row.Scan(&commentID); err != nil {
		return 0, fmt.Errorf("create playground comment: %w", err)
	}
	return commentID, nil
}

func (s *Store) GetPlaygroundComment(ctx context.Context, tx *sql.Tx, viewerID, commentID int64) (PlaygroundComment, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT c.id, c.post_id,
		       u.id, u.nickname, u.email, u.department, u.grade,
		       c.parent_comment_id, c.content, c.like_count, c.reply_count,
		       EXISTS(
		         SELECT 1 FROM playground_comment_likes l
		         WHERE l.comment_id = c.id AND l.user_id = $1
		       ) AS liked_by_me,
		       c.created_at, c.updated_at
		FROM playground_comments c
		JOIN users u ON u.id = c.author_id
		JOIN playground_posts p ON p.id = c.post_id
		WHERE c.id = $2
		  AND c.deleted_at IS NULL
		  AND p.deleted_at IS NULL
		  AND u.deleted_at IS NULL
		LIMIT 1
	`, viewerID, commentID)

	comment, err := scanPlaygroundComment(row)
	if err != nil {
		return PlaygroundComment{}, err
	}
	return comment, nil
}

func (s *Store) ListPlaygroundComments(ctx context.Context, tx *sql.Tx, viewerID, postID int64) ([]PlaygroundComment, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT c.id, c.post_id,
		       u.id, u.nickname, u.email, u.department, u.grade,
		       c.parent_comment_id, c.content, c.like_count, c.reply_count,
		       EXISTS(
		         SELECT 1 FROM playground_comment_likes l
		         WHERE l.comment_id = c.id AND l.user_id = $1
		       ) AS liked_by_me,
		       c.created_at, c.updated_at
		FROM playground_comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.post_id = $2
		  AND c.deleted_at IS NULL
		  AND u.deleted_at IS NULL
		ORDER BY c.created_at ASC, c.id ASC
	`, viewerID, postID)
	if err != nil {
		return nil, fmt.Errorf("list playground comments: %w", err)
	}
	defer rows.Close()

	comments := make([]PlaygroundComment, 0)
	for rows.Next() {
		comment, err := scanPlaygroundCommentRows(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate playground comments: %w", err)
	}
	return comments, nil
}

func (s *Store) AddPlaygroundCommentReplyCount(ctx context.Context, tx *sql.Tx, commentID int64, delta int) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE playground_comments
		SET reply_count = GREATEST(0, reply_count + $2)
		WHERE id = $1 AND deleted_at IS NULL
	`, commentID, delta)
	if err != nil {
		return fmt.Errorf("update playground comment reply count: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) CreatePlaygroundCommentLike(ctx context.Context, tx *sql.Tx, commentID, userID int64, now time.Time) (bool, error) {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO playground_comment_likes (comment_id, user_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (comment_id, user_id) DO NOTHING
	`, commentID, userID, now)
	if err != nil {
		return false, fmt.Errorf("create playground comment like: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("playground comment like rows affected: %w", err)
	}
	return rows > 0, nil
}

func (s *Store) DeletePlaygroundCommentLike(ctx context.Context, tx *sql.Tx, commentID, userID int64) (bool, error) {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM playground_comment_likes
		WHERE comment_id = $1 AND user_id = $2
	`, commentID, userID)
	if err != nil {
		return false, fmt.Errorf("delete playground comment like: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("playground comment unlike rows affected: %w", err)
	}
	return rows > 0, nil
}

func (s *Store) AddPlaygroundCommentLikeCount(ctx context.Context, tx *sql.Tx, commentID int64, delta int) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE playground_comments
		SET like_count = GREATEST(0, like_count + $2)
		WHERE id = $1 AND deleted_at IS NULL
	`, commentID, delta)
	if err != nil {
		return fmt.Errorf("update playground comment like count: %w", err)
	}
	return requireRowsAffected(result)
}

func scanPlaygroundPost(row *sql.Row) (PlaygroundPost, error) {
	var post PlaygroundPost
	var rawImageURLs string
	err := row.Scan(
		&post.ID,
		&post.Author.UserID,
		&post.Author.Nickname,
		&post.Author.Email,
		&post.Author.Department,
		&post.Author.Grade,
		&post.Content,
		&rawImageURLs,
		&post.LikeCount,
		&post.CommentCount,
		&post.LikedByMe,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return PlaygroundPost{}, ErrNotFound
		}
		return PlaygroundPost{}, fmt.Errorf("scan playground post: %w", err)
	}
	if err := json.Unmarshal([]byte(rawImageURLs), &post.ImageURLs); err != nil {
		return PlaygroundPost{}, fmt.Errorf("decode playground post image urls: %w", err)
	}
	return post, nil
}

func scanPlaygroundPostRows(rows *sql.Rows) (PlaygroundPost, error) {
	var post PlaygroundPost
	var rawImageURLs string
	if err := rows.Scan(
		&post.ID,
		&post.Author.UserID,
		&post.Author.Nickname,
		&post.Author.Email,
		&post.Author.Department,
		&post.Author.Grade,
		&post.Content,
		&rawImageURLs,
		&post.LikeCount,
		&post.CommentCount,
		&post.LikedByMe,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		return PlaygroundPost{}, fmt.Errorf("scan playground post row: %w", err)
	}
	if err := json.Unmarshal([]byte(rawImageURLs), &post.ImageURLs); err != nil {
		return PlaygroundPost{}, fmt.Errorf("decode playground post row image urls: %w", err)
	}
	return post, nil
}

func scanPlaygroundComment(row *sql.Row) (PlaygroundComment, error) {
	var comment PlaygroundComment
	err := row.Scan(
		&comment.ID,
		&comment.PostID,
		&comment.Author.UserID,
		&comment.Author.Nickname,
		&comment.Author.Email,
		&comment.Author.Department,
		&comment.Author.Grade,
		&comment.ParentCommentID,
		&comment.Content,
		&comment.LikeCount,
		&comment.ReplyCount,
		&comment.LikedByMe,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return PlaygroundComment{}, ErrNotFound
		}
		return PlaygroundComment{}, fmt.Errorf("scan playground comment: %w", err)
	}
	return comment, nil
}

func scanPlaygroundCommentRows(rows *sql.Rows) (PlaygroundComment, error) {
	var comment PlaygroundComment
	if err := rows.Scan(
		&comment.ID,
		&comment.PostID,
		&comment.Author.UserID,
		&comment.Author.Nickname,
		&comment.Author.Email,
		&comment.Author.Department,
		&comment.Author.Grade,
		&comment.ParentCommentID,
		&comment.Content,
		&comment.LikeCount,
		&comment.ReplyCount,
		&comment.LikedByMe,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	); err != nil {
		return PlaygroundComment{}, fmt.Errorf("scan playground comment row: %w", err)
	}
	return comment, nil
}
