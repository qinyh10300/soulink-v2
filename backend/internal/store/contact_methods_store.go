package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ContactMethod struct {
	UserID       int64
	ContactType  string
	ContactValue string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UpsertContactMethodParams struct {
	UserID       int64
	ContactType  string
	ContactValue string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ListUserContactMethods 查询用户的全部联系方式。
func (s *Store) ListUserContactMethods(ctx context.Context, tx *sql.Tx, userID int64) ([]ContactMethod, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT user_id, contact_type, contact_value, created_at, updated_at
		FROM user_contact_methods
		WHERE user_id = $1
		ORDER BY contact_type ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user contact methods: %w", err)
	}
	defer rows.Close()

	contacts := make([]ContactMethod, 0)
	for rows.Next() {
		var contact ContactMethod
		if err := rows.Scan(
			&contact.UserID,
			&contact.ContactType,
			&contact.ContactValue,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user contact method: %w", err)
		}
		contacts = append(contacts, contact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user contact methods: %w", err)
	}
	return contacts, nil
}

// UpsertUserContactMethod 保存或覆盖用户指定渠道的联系方式。
func (s *Store) UpsertUserContactMethod(ctx context.Context, tx *sql.Tx, params UpsertContactMethodParams) (ContactMethod, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO user_contact_methods (user_id, contact_type, contact_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, contact_type)
		DO UPDATE SET
			contact_value = EXCLUDED.contact_value,
			updated_at = EXCLUDED.updated_at
		RETURNING user_id, contact_type, contact_value, created_at, updated_at
	`, params.UserID, params.ContactType, params.ContactValue, params.CreatedAt, params.UpdatedAt)

	var contact ContactMethod
	if err := row.Scan(
		&contact.UserID,
		&contact.ContactType,
		&contact.ContactValue,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	); err != nil {
		return ContactMethod{}, fmt.Errorf("upsert user contact method: %w", err)
	}
	return contact, nil
}

// DeleteUserContactMethod 删除用户指定渠道的联系方式。
func (s *Store) DeleteUserContactMethod(ctx context.Context, tx *sql.Tx, userID int64, contactType string) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM user_contact_methods
		WHERE user_id = $1 AND contact_type = $2
	`, userID, contactType)
	if err != nil {
		return fmt.Errorf("delete user contact method: %w", err)
	}
	return requireRowsAffected(result)
}

// DeleteAllUserContactMethods 删除用户的全部联系方式。
func (s *Store) DeleteAllUserContactMethods(ctx context.Context, tx *sql.Tx, userID int64) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM user_contact_methods
		WHERE user_id = $1
	`, userID); err != nil {
		return fmt.Errorf("delete all user contact methods: %w", err)
	}
	return nil
}

// SyncUserPrimaryContact 用联系方式表中的最新记录刷新 users 表里的兼容字段。
func (s *Store) SyncUserPrimaryContact(ctx context.Context, tx *sql.Tx, userID int64, now time.Time) error {
	var contactType sql.NullString
	var contactValue sql.NullString

	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT contact_type, contact_value
		FROM user_contact_methods
		WHERE user_id = $1
		ORDER BY updated_at DESC, contact_type ASC
		LIMIT 1
	`, userID)
	if err := row.Scan(&contactType, &contactValue); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("select primary contact method: %w", err)
	}

	var contactTypeArg any
	var contactValueArg any
	if contactType.Valid {
		contactTypeArg = contactType.String
	}
	if contactValue.Valid {
		contactValueArg = contactValue.String
	}

	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE users
		SET contact_type = $2,
		    contact_value = $3,
		    updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, contactTypeArg, contactValueArg, now)
	if err != nil {
		return fmt.Errorf("sync user primary contact: %w", err)
	}
	return requireRowsAffected(result)
}
