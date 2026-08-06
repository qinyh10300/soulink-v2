package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type UserTrustProfile struct {
	UserID             int64
	EmailVerified      bool
	ProfileComplete    bool
	PhotoStatus        string
	PhotoRequestID     sql.NullInt64
	PhotoReviewNotes   string
	PhotoAppealReason  string
	RecentlyActive     bool
	LastActiveAt       sql.NullTime
	PhotoRequestAt     sql.NullTime
	PhotoRequestUpdate sql.NullTime
}

type PhotoVerificationRequest struct {
	ID           int64
	UserID       int64
	UserNickname string
	UserEmail    string
	EvidenceMIME string
	Status       string
	ReviewNotes  string
	ReviewedBy   sql.NullInt64
	ReviewedAt   sql.NullTime
	AppealReason string
	AppealedAt   sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SafetyReviewEvent struct {
	ID             int64
	EntityType     string
	EntityID       int64
	ActorUserID    int64
	ActorNickname  string
	Action         string
	PreviousStatus string
	NextStatus     string
	Note           string
	CreatedAt      time.Time
}

type TrustedContact struct {
	ID           int64
	UserID       int64
	Label        string
	ContactType  string
	ContactValue string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type MeetupSafetyShare struct {
	ID               int64
	MeetupID         int64
	SharerID         int64
	TrustedContactID int64
	Summary          json.RawMessage
	DeliveryStatus   string
	DeliveredAt      sql.NullTime
	CreatedAt        time.Time
}

func (s *Store) GetUserTrustProfile(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	now time.Time,
) (UserTrustProfile, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT user_account.id,
		       user_account.email_verified_at IS NOT NULL,
		       (
		         NULLIF(BTRIM(user_account.nickname), '') IS NOT NULL
		         AND NULLIF(BTRIM(user_account.department), '') IS NOT NULL
		         AND NULLIF(BTRIM(user_account.grade), '') IS NOT NULL
		         AND (
		           user_account.has_completed_questionnaire
		           OR EXISTS (
		             SELECT 1 FROM matching_profiles profile
		             WHERE profile.user_id = user_account.id
		               AND NULLIF(BTRIM(profile.bio), '') IS NOT NULL
		           )
		         )
		       ),
		       COALESCE(photo.status, 'not_submitted'), photo.id,
		       COALESCE(photo.review_notes, ''), COALESCE(photo.appeal_reason, ''),
		       COALESCE(activity.last_active_at >= $2::timestamptz - INTERVAL '14 days', FALSE),
		       activity.last_active_at, photo.created_at, photo.updated_at
		FROM users user_account
		LEFT JOIN LATERAL (
		  SELECT request.id, request.status, request.review_notes, request.appeal_reason,
		         request.created_at, request.updated_at
		  FROM photo_verification_requests request
		  WHERE request.user_id = user_account.id
		  ORDER BY request.updated_at DESC, request.id DESC
		  LIMIT 1
		) photo ON TRUE
		LEFT JOIN LATERAL (
		  SELECT MAX(session.last_seen_at) AS last_active_at
		  FROM auth_sessions session
		  WHERE session.user_id = user_account.id
		) activity ON TRUE
		WHERE user_account.id = $1 AND user_account.deleted_at IS NULL
	`, userID, now)

	var profile UserTrustProfile
	if err := row.Scan(
		&profile.UserID, &profile.EmailVerified, &profile.ProfileComplete,
		&profile.PhotoStatus, &profile.PhotoRequestID, &profile.PhotoReviewNotes,
		&profile.PhotoAppealReason, &profile.RecentlyActive, &profile.LastActiveAt,
		&profile.PhotoRequestAt, &profile.PhotoRequestUpdate,
	); err != nil {
		if err == sql.ErrNoRows {
			return UserTrustProfile{}, ErrNotFound
		}
		return UserTrustProfile{}, fmt.Errorf("get user trust profile: %w", err)
	}
	return profile, nil
}

func (s *Store) CreatePhotoVerificationRequest(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	mimeType string,
	evidence []byte,
	now time.Time,
) (PhotoVerificationRequest, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO photo_verification_requests (
		  user_id, evidence_mime, evidence_data, status, created_at, updated_at
		) VALUES ($1, $2, $3, 'pending', $4, $4)
		RETURNING id, user_id, '', '', evidence_mime, status, review_notes,
		          reviewed_by, reviewed_at, appeal_reason, appealed_at, created_at, updated_at
	`, userID, mimeType, evidence, now)
	return scanPhotoVerificationRequest(row)
}

func (s *Store) GetPhotoVerificationRequest(
	ctx context.Context,
	tx *sql.Tx,
	requestID int64,
) (PhotoVerificationRequest, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, photoVerificationSelect+`
		WHERE request.id = $1
	`, requestID)
	return scanPhotoVerificationRequest(row)
}

func (s *Store) GetPhotoVerificationEvidence(
	ctx context.Context,
	tx *sql.Tx,
	requestID int64,
) (int64, string, []byte, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT user_id, evidence_mime, evidence_data
		FROM photo_verification_requests
		WHERE id = $1
	`, requestID)
	var userID int64
	var mimeType string
	var evidence []byte
	if err := row.Scan(&userID, &mimeType, &evidence); err != nil {
		if err == sql.ErrNoRows {
			return 0, "", nil, ErrNotFound
		}
		return 0, "", nil, fmt.Errorf("get photo verification evidence: %w", err)
	}
	return userID, mimeType, evidence, nil
}

func (s *Store) ListPhotoVerificationRequests(
	ctx context.Context,
	tx *sql.Tx,
	status string,
	limit int,
) ([]PhotoVerificationRequest, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := queryContext(executor(s.db, tx), ctx, photoVerificationSelect+`
		WHERE ($1 = '' OR request.status = $1)
		ORDER BY CASE request.status WHEN 'pending' THEN 0 ELSE 1 END,
		         request.updated_at DESC, request.id DESC
		LIMIT $2
	`, status, limit)
	if err != nil {
		return nil, fmt.Errorf("list photo verification requests: %w", err)
	}
	defer rows.Close()
	result := make([]PhotoVerificationRequest, 0)
	for rows.Next() {
		request, err := scanPhotoVerificationRequest(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, request)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate photo verification requests: %w", err)
	}
	return result, nil
}

func (s *Store) ReviewPhotoVerificationRequest(
	ctx context.Context,
	tx *sql.Tx,
	requestID, adminID int64,
	status, notes string,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE photo_verification_requests
		SET status = $3, review_notes = $4, reviewed_by = $2, reviewed_at = $5,
		    updated_at = $5
		WHERE id = $1 AND status = 'pending'
	`, requestID, adminID, status, notes, now)
	if err != nil {
		return fmt.Errorf("review photo verification request: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) AppealPhotoVerificationRequest(
	ctx context.Context,
	tx *sql.Tx,
	requestID, userID int64,
	reason string,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE photo_verification_requests
		SET status = 'pending', appeal_reason = $3, appealed_at = $4,
		    reviewed_by = NULL, reviewed_at = NULL, updated_at = $4
		WHERE id = $1 AND user_id = $2 AND status = 'rejected'
	`, requestID, userID, reason, now)
	if err != nil {
		return fmt.Errorf("appeal photo verification request: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) CreateSafetyReviewEvent(
	ctx context.Context,
	tx *sql.Tx,
	event SafetyReviewEvent,
) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO safety_review_events (
		  entity_type, entity_id, actor_user_id, action,
		  previous_status, next_status, note, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, event.EntityType, event.EntityID, event.ActorUserID, event.Action,
		event.PreviousStatus, event.NextStatus, event.Note, event.CreatedAt)
	if err != nil {
		return fmt.Errorf("create safety review event: %w", err)
	}
	return nil
}

func (s *Store) ListSafetyReviewEvents(
	ctx context.Context,
	tx *sql.Tx,
	entityType string,
	entityID int64,
) ([]SafetyReviewEvent, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT event.id, event.entity_type, event.entity_id, event.actor_user_id,
		       actor.nickname, event.action, event.previous_status,
		       event.next_status, event.note, event.created_at
		FROM safety_review_events event
		JOIN users actor ON actor.id = event.actor_user_id
		WHERE event.entity_type = $1 AND event.entity_id = $2
		ORDER BY event.created_at, event.id
	`, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("list safety review events: %w", err)
	}
	defer rows.Close()
	result := make([]SafetyReviewEvent, 0)
	for rows.Next() {
		var event SafetyReviewEvent
		if err := rows.Scan(
			&event.ID, &event.EntityType, &event.EntityID, &event.ActorUserID,
			&event.ActorNickname, &event.Action, &event.PreviousStatus,
			&event.NextStatus, &event.Note, &event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan safety review event: %w", err)
		}
		result = append(result, event)
	}
	return result, rows.Err()
}

func (s *Store) CreateTrustedContact(
	ctx context.Context,
	tx *sql.Tx,
	contact TrustedContact,
	now time.Time,
) (TrustedContact, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO trusted_contacts (
		  user_id, label, contact_type, contact_value, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, user_id, label, contact_type, contact_value, created_at, updated_at
	`, contact.UserID, contact.Label, contact.ContactType, contact.ContactValue, now)
	return scanTrustedContact(row)
}

func (s *Store) ListTrustedContacts(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
) ([]TrustedContact, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT id, user_id, label, contact_type, contact_value, created_at, updated_at
		FROM trusted_contacts
		WHERE user_id = $1
		ORDER BY updated_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list trusted contacts: %w", err)
	}
	defer rows.Close()
	result := make([]TrustedContact, 0)
	for rows.Next() {
		contact, err := scanTrustedContact(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, contact)
	}
	return result, rows.Err()
}

func (s *Store) GetTrustedContact(
	ctx context.Context,
	tx *sql.Tx,
	userID, contactID int64,
) (TrustedContact, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, user_id, label, contact_type, contact_value, created_at, updated_at
		FROM trusted_contacts
		WHERE id = $2 AND user_id = $1
	`, userID, contactID)
	return scanTrustedContact(row)
}

func (s *Store) DeleteTrustedContact(
	ctx context.Context,
	tx *sql.Tx,
	userID, contactID int64,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM trusted_contacts WHERE id = $2 AND user_id = $1
	`, userID, contactID)
	if err != nil {
		return fmt.Errorf("delete trusted contact: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) CreateMeetupSafetyShare(
	ctx context.Context,
	tx *sql.Tx,
	share MeetupSafetyShare,
	now time.Time,
) (MeetupSafetyShare, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO meetup_safety_shares (
		  meetup_id, sharer_id, trusted_contact_id, summary,
		  delivery_status, delivered_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, meetup_id, sharer_id, trusted_contact_id, summary,
		          delivery_status, delivered_at, created_at
	`, share.MeetupID, share.SharerID, share.TrustedContactID, share.Summary,
		share.DeliveryStatus, nullableTime(share.DeliveredAt), now)
	return scanMeetupSafetyShare(row)
}

func (s *Store) UpdateMeetupSafetyShareDelivery(
	ctx context.Context,
	tx *sql.Tx,
	shareID int64,
	status string,
	deliveredAt sql.NullTime,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE meetup_safety_shares
		SET delivery_status = $2, delivered_at = $3
		WHERE id = $1
	`, shareID, status, nullableTime(deliveredAt))
	if err != nil {
		return fmt.Errorf("update meetup safety share delivery: %w", err)
	}
	return requireRowsAffected(result)
}

const photoVerificationSelect = `
	SELECT request.id, request.user_id, user_account.nickname, user_account.email,
	       request.evidence_mime, request.status, request.review_notes,
	       request.reviewed_by, request.reviewed_at, request.appeal_reason,
	       request.appealed_at, request.created_at, request.updated_at
	FROM photo_verification_requests request
	JOIN users user_account ON user_account.id = request.user_id
`

func scanPhotoVerificationRequest(row rowScanner) (PhotoVerificationRequest, error) {
	var request PhotoVerificationRequest
	if err := row.Scan(
		&request.ID, &request.UserID, &request.UserNickname, &request.UserEmail,
		&request.EvidenceMIME, &request.Status, &request.ReviewNotes,
		&request.ReviewedBy, &request.ReviewedAt, &request.AppealReason,
		&request.AppealedAt, &request.CreatedAt, &request.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return PhotoVerificationRequest{}, ErrNotFound
		}
		return PhotoVerificationRequest{}, fmt.Errorf("scan photo verification request: %w", err)
	}
	return request, nil
}

func scanTrustedContact(row rowScanner) (TrustedContact, error) {
	var contact TrustedContact
	if err := row.Scan(
		&contact.ID, &contact.UserID, &contact.Label, &contact.ContactType,
		&contact.ContactValue, &contact.CreatedAt, &contact.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return TrustedContact{}, ErrNotFound
		}
		return TrustedContact{}, fmt.Errorf("scan trusted contact: %w", err)
	}
	return contact, nil
}

func scanMeetupSafetyShare(row rowScanner) (MeetupSafetyShare, error) {
	var share MeetupSafetyShare
	if err := row.Scan(
		&share.ID, &share.MeetupID, &share.SharerID, &share.TrustedContactID,
		&share.Summary, &share.DeliveryStatus, &share.DeliveredAt, &share.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return MeetupSafetyShare{}, ErrNotFound
		}
		return MeetupSafetyShare{}, fmt.Errorf("scan meetup safety share: %w", err)
	}
	return share, nil
}
