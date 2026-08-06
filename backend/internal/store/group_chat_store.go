package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ChatGroup struct {
	ID        int64
	Name      string
	OwnerID   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChatGroupSummary struct {
	ChatGroup
	MemberCount       int
	LastMessage       string
	LastMessageAt     sql.NullTime
	LastMessageSender string
	UnreadCount       int
}

type ChatGroupMember struct {
	GroupID    int64
	UserID     int64
	Role       string
	Nickname   string
	Department string
	Grade      string
	JoinedAt   time.Time
}

type ChatGroupMessage struct {
	ID             int64
	GroupID        int64
	SenderID       int64
	SenderNickname string
	Content        string
	ContentType    string
	CreatedAt      time.Time
}

func (s *Store) CreateChatGroup(ctx context.Context, tx *sql.Tx, ownerID int64, name string, now time.Time) (ChatGroup, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO chat_groups (name, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		RETURNING id, name, owner_id, created_at, updated_at
	`, name, ownerID, now)
	return scanChatGroup(row)
}

func (s *Store) GetChatGroup(ctx context.Context, tx *sql.Tx, groupID int64) (ChatGroup, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, name, owner_id, created_at, updated_at
		FROM chat_groups
		WHERE id = $1 AND deleted_at IS NULL
	`, groupID)
	return scanChatGroup(row)
}

func (s *Store) UpdateChatGroupName(ctx context.Context, tx *sql.Tx, groupID int64, name string, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE chat_groups
		SET name = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, groupID, name, now)
	if err != nil {
		return fmt.Errorf("update chat group name: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) UpsertChatGroupMember(ctx context.Context, tx *sql.Tx, groupID, userID int64, role string, now time.Time) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO chat_group_members (group_id, user_id, role, last_read_at, joined_at, left_at)
		VALUES ($1, $2, $3, $4, $4, NULL)
		ON CONFLICT (group_id, user_id) DO UPDATE
		SET role = EXCLUDED.role,
		    joined_at = EXCLUDED.joined_at,
		    left_at = NULL
	`, groupID, userID, role, now)
	if err != nil {
		return fmt.Errorf("upsert chat group member: %w", err)
	}
	return nil
}

func (s *Store) RemoveChatGroupMember(ctx context.Context, tx *sql.Tx, groupID, userID int64, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE chat_group_members
		SET left_at = $3
		WHERE group_id = $1 AND user_id = $2 AND left_at IS NULL
	`, groupID, userID, now)
	if err != nil {
		return fmt.Errorf("remove chat group member: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) GetChatGroupMember(ctx context.Context, tx *sql.Tx, groupID, userID int64) (ChatGroupMember, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT gm.group_id, gm.user_id, gm.role,
		       u.nickname, u.department, u.grade, gm.joined_at
		FROM chat_group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = $1 AND gm.user_id = $2
		  AND gm.left_at IS NULL AND u.deleted_at IS NULL
	`, groupID, userID)
	return scanChatGroupMember(row)
}

func (s *Store) ListChatGroupMembers(ctx context.Context, tx *sql.Tx, groupID int64) ([]ChatGroupMember, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT gm.group_id, gm.user_id, gm.role,
		       u.nickname, u.department, u.grade, gm.joined_at
		FROM chat_group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = $1 AND gm.left_at IS NULL AND u.deleted_at IS NULL
		ORDER BY CASE WHEN gm.role = 'owner' THEN 0 ELSE 1 END, gm.joined_at, gm.user_id
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list chat group members: %w", err)
	}
	defer rows.Close()

	members := make([]ChatGroupMember, 0)
	for rows.Next() {
		var member ChatGroupMember
		if err := rows.Scan(
			&member.GroupID,
			&member.UserID,
			&member.Role,
			&member.Nickname,
			&member.Department,
			&member.Grade,
			&member.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat group member row: %w", err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat group members: %w", err)
	}
	return members, nil
}

func (s *Store) ListChatGroups(ctx context.Context, tx *sql.Tx, userID int64) ([]ChatGroupSummary, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT g.id, g.name, g.owner_id, g.created_at, g.updated_at,
		       members.cnt,
		       COALESCE(lm.content, ''), lm.created_at, COALESCE(lm.sender_nickname, ''),
		       COALESCE(unread.cnt, 0)
		FROM chat_group_members mine
		JOIN chat_groups g ON g.id = mine.group_id AND g.deleted_at IS NULL
		JOIN LATERAL (
			SELECT COUNT(*) AS cnt
			FROM chat_group_members gm
			WHERE gm.group_id = g.id AND gm.left_at IS NULL
		) members ON TRUE
		LEFT JOIN LATERAL (
			SELECT m.content, m.created_at, u.nickname AS sender_nickname
			FROM chat_group_messages m
			JOIN users u ON u.id = m.sender_id
			WHERE m.group_id = g.id
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT 1
		) lm ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS cnt
			FROM chat_group_messages m
			WHERE m.group_id = g.id
			  AND m.sender_id != $1
			  AND m.created_at > COALESCE(mine.last_read_at, mine.joined_at)
		) unread ON TRUE
		WHERE mine.user_id = $1 AND mine.left_at IS NULL
		ORDER BY COALESCE(lm.created_at, g.created_at) DESC, g.id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list chat groups: %w", err)
	}
	defer rows.Close()

	groups := make([]ChatGroupSummary, 0)
	for rows.Next() {
		var group ChatGroupSummary
		if err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.OwnerID,
			&group.CreatedAt,
			&group.UpdatedAt,
			&group.MemberCount,
			&group.LastMessage,
			&group.LastMessageAt,
			&group.LastMessageSender,
			&group.UnreadCount,
		); err != nil {
			return nil, fmt.Errorf("scan chat group summary: %w", err)
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat groups: %w", err)
	}
	return groups, nil
}

func (s *Store) CreateChatGroupMessage(ctx context.Context, tx *sql.Tx, groupID, senderID int64, content, contentType string, now time.Time) (ChatGroupMessage, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		WITH inserted AS (
			INSERT INTO chat_group_messages (group_id, sender_id, content, content_type, created_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, group_id, sender_id, content, content_type, created_at
		)
		SELECT i.id, i.group_id, i.sender_id, u.nickname,
		       i.content, i.content_type, i.created_at
		FROM inserted i
		JOIN users u ON u.id = i.sender_id
	`, groupID, senderID, content, contentType, now)
	return scanChatGroupMessage(row)
}

func (s *Store) ListChatGroupMessages(ctx context.Context, tx *sql.Tx, groupID, beforeID int64, limit int) ([]ChatGroupMessage, error) {
	query := `
		SELECT m.id, m.group_id, m.sender_id, u.nickname,
		       m.content, m.content_type, m.created_at
		FROM chat_group_messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.group_id = $1
	`
	args := []any{groupID}
	if beforeID > 0 {
		args = append(args, beforeID)
		query += fmt.Sprintf(" AND m.id < $%d", len(args))
	}
	args = append(args, limit)
	query += fmt.Sprintf(" ORDER BY m.created_at DESC, m.id DESC LIMIT $%d", len(args))

	rows, err := executor(s.db, tx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list chat group messages: %w", err)
	}
	defer rows.Close()

	messages := make([]ChatGroupMessage, 0)
	for rows.Next() {
		var message ChatGroupMessage
		if err := rows.Scan(
			&message.ID,
			&message.GroupID,
			&message.SenderID,
			&message.SenderNickname,
			&message.Content,
			&message.ContentType,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat group message row: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat group messages: %w", err)
	}
	return messages, nil
}

func (s *Store) MarkChatGroupRead(ctx context.Context, tx *sql.Tx, groupID, userID int64, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE chat_group_members
		SET last_read_at = $3
		WHERE group_id = $1 AND user_id = $2 AND left_at IS NULL
	`, groupID, userID, now)
	if err != nil {
		return fmt.Errorf("mark chat group read: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) GetChatGroupUnreadCount(ctx context.Context, tx *sql.Tx, userID int64) (int, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM chat_group_members gm
		JOIN chat_groups g ON g.id = gm.group_id AND g.deleted_at IS NULL
		JOIN chat_group_messages m ON m.group_id = gm.group_id
		WHERE gm.user_id = $1 AND gm.left_at IS NULL
		  AND m.sender_id != $1
		  AND m.created_at > COALESCE(gm.last_read_at, gm.joined_at)
	`, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("get chat group unread count: %w", err)
	}
	return count, nil
}

func scanChatGroup(row *sql.Row) (ChatGroup, error) {
	var group ChatGroup
	if err := row.Scan(&group.ID, &group.Name, &group.OwnerID, &group.CreatedAt, &group.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return ChatGroup{}, ErrNotFound
		}
		return ChatGroup{}, fmt.Errorf("scan chat group: %w", err)
	}
	return group, nil
}

func scanChatGroupMember(row *sql.Row) (ChatGroupMember, error) {
	var member ChatGroupMember
	if err := row.Scan(
		&member.GroupID,
		&member.UserID,
		&member.Role,
		&member.Nickname,
		&member.Department,
		&member.Grade,
		&member.JoinedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return ChatGroupMember{}, ErrNotFound
		}
		return ChatGroupMember{}, fmt.Errorf("scan chat group member: %w", err)
	}
	return member, nil
}

func scanChatGroupMessage(row *sql.Row) (ChatGroupMessage, error) {
	var message ChatGroupMessage
	if err := row.Scan(
		&message.ID,
		&message.GroupID,
		&message.SenderID,
		&message.SenderNickname,
		&message.Content,
		&message.ContentType,
		&message.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return ChatGroupMessage{}, ErrNotFound
		}
		return ChatGroupMessage{}, fmt.Errorf("scan chat group message: %w", err)
	}
	return message, nil
}
