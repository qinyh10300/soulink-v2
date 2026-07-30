package feedback

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type Service struct {
	store  FeedbackStore
	logger *slog.Logger
}

type FeedbackStore interface {
	CreateUserFeedback(ctx context.Context, tx *sql.Tx, params store.CreateUserFeedbackParams) (store.UserFeedback, error)
}

type SubmitInput struct {
	UserID       int64
	UserEmail    string
	Category     string
	Rating       int
	Details      string
	ContactEmail string
}

type SubmitResult struct {
	TicketID   string `json:"ticketId"`
	ReceivedAt string `json:"receivedAt"`
}

func NewService(dataStore FeedbackStore, logger *slog.Logger) *Service {
	return &Service{store: dataStore, logger: logger}
}

func (s *Service) Submit(ctx context.Context, input SubmitInput) (SubmitResult, error) {
	if s == nil || s.store == nil || s.logger == nil {
		return SubmitResult{}, httpapi.FeatureNotReady("feedback")
	}

	input.Category = strings.TrimSpace(strings.ToLower(input.Category))
	input.Details = strings.TrimSpace(input.Details)
	input.ContactEmail = strings.TrimSpace(strings.ToLower(input.ContactEmail))

	switch input.Category {
	case "match", "bug", "feature", "other":
	default:
		return SubmitResult{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	if input.Rating < 1 || input.Rating > 5 {
		return SubmitResult{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	if len(input.Details) > 2000 {
		return SubmitResult{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	if input.ContactEmail != "" {
		if _, err := mail.ParseAddress(input.ContactEmail); err != nil {
			return SubmitResult{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
	}

	now := time.Now().UTC()
	result := SubmitResult{
		TicketID:   fmt.Sprintf("fb_%d", now.UnixNano()),
		ReceivedAt: now.Format(time.RFC3339),
	}

	if _, err := s.store.CreateUserFeedback(ctx, nil, store.CreateUserFeedbackParams{
		TicketID:     result.TicketID,
		UserID:       input.UserID,
		UserEmail:    input.UserEmail,
		Category:     input.Category,
		Rating:       input.Rating,
		Details:      input.Details,
		ContactEmail: input.ContactEmail,
		SubmittedAt:  now,
	}); err != nil {
		return SubmitResult{}, httpapi.InternalError(err)
	}

	s.logger.Info(
		"feedback submitted",
		"ticket_id", result.TicketID,
		"user_id", input.UserID,
		"user_email", input.UserEmail,
		"category", input.Category,
		"rating", input.Rating,
		"contact_email", input.ContactEmail,
		"details", input.Details,
	)

	return result, nil
}
