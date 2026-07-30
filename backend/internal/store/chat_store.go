package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Conversation struct {
	ID                         int64
	User1ID                    int64
	User2ID                    int64
	User1LastReadAt            sql.NullTime
	User2LastReadAt            sql.NullTime
	FriendshipDisabledAt       sql.NullTime
	FriendshipDisabledByUserID sql.NullInt64
	CreatedAt                  time.Time
}

type ConversationWithPeer struct {
	Conversation
	PeerNickname       string
	PeerEmail          string
	PeerDepartment     string
	PeerGrade          string
	LastMessageContent string
	LastMessageAt      sql.NullTime
	UnreadCount        int
}

type Message struct {
	ID             int64
	ConversationID int64
	SenderID       int64
	Content        string
	ContentType    string
	CreatedAt      time.Time
}

// CreateConversation creates a conversation between two users (IDs are auto-ordered).
func (s *Store) CreateConversation(ctx context.Context, tx *sql.Tx, userA, userB int64, now time.Time) (Conversation, error) {
	u1, u2 := userA, userB
	if u1 > u2 {
		u1, u2 = u2, u1
	}
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO conversations (user1_id, user2_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user1_id, user2_id) DO UPDATE
		SET friendship_disabled_at = NULL,
		    friendship_disabled_by_user_id = NULL
		RETURNING id, user1_id, user2_id, user1_last_read_at, user2_last_read_at,
		          friendship_disabled_at, friendship_disabled_by_user_id, created_at
	`, u1, u2, now)
	return scanConversation(row)
}

func (s *Store) GetConversation(ctx context.Context, tx *sql.Tx, id int64) (Conversation, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, user1_id, user2_id, user1_last_read_at, user2_last_read_at,
		       friendship_disabled_at, friendship_disabled_by_user_id, created_at
		FROM conversations WHERE id = $1
	`, id)
	return scanConversation(row)
}

func (s *Store) GetConversationBetween(ctx context.Context, tx *sql.Tx, userA, userB int64) (Conversation, error) {
	u1, u2 := userA, userB
	if u1 > u2 {
		u1, u2 = u2, u1
	}
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, user1_id, user2_id, user1_last_read_at, user2_last_read_at,
		       friendship_disabled_at, friendship_disabled_by_user_id, created_at
		FROM conversations WHERE user1_id = $1 AND user2_id = $2
	`, u1, u2)
	return scanConversation(row)
}

func (s *Store) DisableConversationBetween(ctx context.Context, tx *sql.Tx, userA, userB, disabledByUserID int64, now time.Time) (Conversation, error) {
	u1, u2 := userA, userB
	if u1 > u2 {
		u1, u2 = u2, u1
	}
	row := executor(s.db, tx).QueryRowContext(ctx, `
		UPDATE conversations
		SET friendship_disabled_at = $4,
		    friendship_disabled_by_user_id = $3
		WHERE user1_id = $1 AND user2_id = $2
		RETURNING id, user1_id, user2_id, user1_last_read_at, user2_last_read_at,
		          friendship_disabled_at, friendship_disabled_by_user_id, created_at
	`, u1, u2, disabledByUserID, now)
	return scanConversation(row)
}

// ListConversations returns all conversations for a user with peer info, last message, and unread count.
func (s *Store) ListConversations(ctx context.Context, tx *sql.Tx, userID int64) ([]ConversationWithPeer, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT c.id, c.user1_id, c.user2_id, c.user1_last_read_at, c.user2_last_read_at,
		       c.friendship_disabled_at, c.friendship_disabled_by_user_id, c.created_at,
		       u.nickname, u.email, u.department, u.grade,
		       COALESCE(lm.content, ''),
		       lm.created_at,
		       COALESCE(unread.cnt, 0)
		FROM conversations c
		JOIN users u ON u.id = CASE WHEN c.user1_id = $1 THEN c.user2_id ELSE c.user1_id END
		LEFT JOIN LATERAL (
			SELECT content, created_at FROM messages
			WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1
		) lm ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS cnt FROM messages
			WHERE conversation_id = c.id
			  AND sender_id != $1
			  AND created_at > COALESCE(
				CASE WHEN c.user1_id = $1 THEN c.user1_last_read_at ELSE c.user2_last_read_at END,
				'1970-01-01'::timestamptz
			  )
		) unread ON TRUE
		WHERE (c.user1_id = $1 OR c.user2_id = $1) AND u.deleted_at IS NULL
		ORDER BY COALESCE(lm.created_at, c.created_at) DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	results := make([]ConversationWithPeer, 0)
	for rows.Next() {
		var cp ConversationWithPeer
		if err := rows.Scan(
			&cp.ID, &cp.User1ID, &cp.User2ID, &cp.User1LastReadAt, &cp.User2LastReadAt,
			&cp.FriendshipDisabledAt, &cp.FriendshipDisabledByUserID, &cp.CreatedAt,
			&cp.PeerNickname, &cp.PeerEmail, &cp.PeerDepartment, &cp.PeerGrade,
			&cp.LastMessageContent, &cp.LastMessageAt, &cp.UnreadCount,
		); err != nil {
			return nil, fmt.Errorf("scan conversation with peer: %w", err)
		}
		results = append(results, cp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate conversations: %w", err)
	}
	return results, nil
}

// MarkConversationRead updates the last-read timestamp for the given user in the conversation.
func (s *Store) MarkConversationRead(ctx context.Context, tx *sql.Tx, convID, userID int64, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE conversations
		SET user1_last_read_at = CASE WHEN user1_id = $2 THEN $3 ELSE user1_last_read_at END,
		    user2_last_read_at = CASE WHEN user2_id = $2 THEN $3 ELSE user2_last_read_at END
		WHERE id = $1 AND (user1_id = $2 OR user2_id = $2)
	`, convID, userID, now)
	if err != nil {
		return fmt.Errorf("mark conversation read: %w", err)
	}
	return requireRowsAffected(result)
}

// CreateMessage inserts a new message into the conversation.
func (s *Store) CreateMessage(ctx context.Context, tx *sql.Tx, convID, senderID int64, content, contentType string, now time.Time) (Message, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO messages (conversation_id, sender_id, content, content_type, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, conversation_id, sender_id, content, content_type, created_at
	`, convID, senderID, content, contentType, now)
	return scanMessage(row)
}

// ListMessages returns messages in a conversation, paginated by cursor.
func (s *Store) ListMessages(ctx context.Context, tx *sql.Tx, convID int64, beforeID int64, limit int) ([]Message, error) {
	var query string
	var args []any
	if beforeID > 0 {
		query = `
			SELECT id, conversation_id, sender_id, content, content_type, created_at
			FROM messages WHERE conversation_id = $1 AND id < $2
			ORDER BY created_at DESC, id DESC LIMIT $3
		`
		args = []any{convID, beforeID, limit}
	} else {
		query = `
			SELECT id, conversation_id, sender_id, content, content_type, created_at
			FROM messages WHERE conversation_id = $1
			ORDER BY created_at DESC, id DESC LIMIT $2
		`
		args = []any{convID, limit}
	}

	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.ContentType, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}
	return messages, nil
}

// GetTotalUnreadCount returns the total number of unread messages across all conversations for a user.
func (s *Store) GetTotalUnreadCount(ctx context.Context, tx *sql.Tx, userID int64) (int, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT COALESCE(SUM(cnt), 0) FROM (
			SELECT COUNT(*) AS cnt FROM messages m
			JOIN conversations c ON c.id = m.conversation_id
			WHERE (c.user1_id = $1 OR c.user2_id = $1)
			  AND m.sender_id != $1
			  AND m.created_at > COALESCE(
				CASE WHEN c.user1_id = $1 THEN c.user1_last_read_at ELSE c.user2_last_read_at END,
				'1970-01-01'::timestamptz
			  )
		) sub
	`, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("get total unread count: %w", err)
	}
	return count, nil
}

func scanConversation(row *sql.Row) (Conversation, error) {
	var c Conversation
	err := row.Scan(
		&c.ID,
		&c.User1ID,
		&c.User2ID,
		&c.User1LastReadAt,
		&c.User2LastReadAt,
		&c.FriendshipDisabledAt,
		&c.FriendshipDisabledByUserID,
		&c.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, ErrNotFound
		}
		return Conversation{}, fmt.Errorf("scan conversation: %w", err)
	}
	return c, nil
}

func scanMessage(row *sql.Row) (Message, error) {
	var m Message
	err := row.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.ContentType, &m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Message{}, ErrNotFound
		}
		return Message{}, fmt.Errorf("scan message: %w", err)
	}
	return m, nil
}
