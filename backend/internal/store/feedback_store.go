package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type UserFeedback struct {
	ID           int64
	TicketID     string
	UserID       int64
	UserEmail    string
	Category     string
	Rating       int
	Details      string
	ContactEmail string
	SubmittedAt  time.Time
}

type CreateUserFeedbackParams struct {
	TicketID     string
	UserID       int64
	UserEmail    string
	Category     string
	Rating       int
	Details      string
	ContactEmail string
	SubmittedAt  time.Time
}

func (s *Store) CreateUserFeedback(ctx context.Context, tx *sql.Tx, params CreateUserFeedbackParams) (UserFeedback, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO user_feedback (
			ticket_id, user_id, user_email, category, rating, details, contact_email, submitted_at
		) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8)
		RETURNING id, ticket_id, user_id, user_email, category, rating, details, contact_email, submitted_at
	`, params.TicketID, params.UserID, params.UserEmail, params.Category, params.Rating, params.Details, params.ContactEmail, params.SubmittedAt)

	var feedback UserFeedback
	var contactEmail sql.NullString
	if err := row.Scan(
		&feedback.ID,
		&feedback.TicketID,
		&feedback.UserID,
		&feedback.UserEmail,
		&feedback.Category,
		&feedback.Rating,
		&feedback.Details,
		&contactEmail,
		&feedback.SubmittedAt,
	); err != nil {
		return UserFeedback{}, fmt.Errorf("create user feedback: %w", err)
	}
	feedback.ContactEmail = nullableString(contactEmail)
	return feedback, nil
}
