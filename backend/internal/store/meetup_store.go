package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type AvailabilityWindow struct {
	ID          int64
	UserID      int64
	WindowType  string
	Timezone    string
	StartsAt    sql.NullTime
	EndsAt      sql.NullTime
	Weekday     sql.NullInt64
	StartMinute sql.NullInt64
	EndMinute   sql.NullInt64
	ValidFrom   sql.NullTime
	ValidUntil  sql.NullTime
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateAvailabilityWindowParams struct {
	UserID      int64
	WindowType  string
	Timezone    string
	StartsAt    sql.NullTime
	EndsAt      sql.NullTime
	Weekday     sql.NullInt64
	StartMinute sql.NullInt64
	EndMinute   sql.NullInt64
	ValidFrom   sql.NullTime
	ValidUntil  sql.NullTime
	Now         time.Time
}

type BusyInterval struct {
	UserID   int64
	StartsAt time.Time
	EndsAt   time.Time
}

type Meetup struct {
	ID              int64
	ContextType     string
	ContextID       int64
	OrganizerID     int64
	Title           string
	Category        string
	StartsAt        time.Time
	EndsAt          time.Time
	Location        string
	Status          string
	Version         int
	CalendarEventID sql.NullInt64
	CancelledBy     sql.NullInt64
	CancelledReason string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type MeetupParticipant struct {
	MeetupID    int64
	UserID      int64
	Nickname    string
	Department  string
	Grade       string
	Status      string
	RespondedAt sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MeetupFeedback struct {
	MeetupID       int64
	UserID         int64
	Attendance     string
	Rating         sql.NullInt64
	WouldMeetAgain sql.NullBool
	SafetyConcern  bool
	PrivateNotes   string
	SubmittedAt    time.Time
	UpdatedAt      time.Time
	RevokedAt      sql.NullTime
}

type CreateMeetupParams struct {
	ContextType string
	ContextID   int64
	OrganizerID int64
	Title       string
	Category    string
	StartsAt    time.Time
	EndsAt      time.Time
	Location    string
	Now         time.Time
}

type UpdateMeetupParams struct {
	MeetupID        int64
	ExpectedVersion int
	StartsAt        time.Time
	EndsAt          time.Time
	Location        string
	Status          string
	CalendarEventID sql.NullInt64
	CancelledBy     sql.NullInt64
	CancelledReason string
	Now             time.Time
}

type UpsertMeetupFeedbackParams struct {
	MeetupID       int64
	UserID         int64
	Attendance     string
	Rating         sql.NullInt64
	WouldMeetAgain sql.NullBool
	SafetyConcern  bool
	PrivateNotes   string
	Now            time.Time
}

func (s *Store) DeleteAvailabilityWindows(ctx context.Context, tx *sql.Tx, userID int64) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM availability_windows WHERE user_id = $1
	`, userID); err != nil {
		return fmt.Errorf("delete availability windows: %w", err)
	}
	return nil
}

func (s *Store) CreateAvailabilityWindow(
	ctx context.Context,
	tx *sql.Tx,
	params CreateAvailabilityWindowParams,
) (AvailabilityWindow, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO availability_windows (
			user_id, window_type, timezone, starts_at, ends_at, weekday,
			start_minute, end_minute, valid_from, valid_until, enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE, $11, $11)
		RETURNING id, user_id, window_type, timezone, starts_at, ends_at, weekday,
		          start_minute, end_minute, valid_from, valid_until, enabled, created_at, updated_at
	`, params.UserID, params.WindowType, params.Timezone, nullableTime(params.StartsAt),
		nullableTime(params.EndsAt), nullableInt64Value(params.Weekday),
		nullableInt64Value(params.StartMinute), nullableInt64Value(params.EndMinute),
		nullableDate(params.ValidFrom), nullableDate(params.ValidUntil), params.Now)
	return scanAvailabilityWindow(row)
}

func (s *Store) ListAvailabilityWindows(
	ctx context.Context,
	tx *sql.Tx,
	userIDs []int64,
) ([]AvailabilityWindow, error) {
	if len(userIDs) == 0 {
		return []AvailabilityWindow{}, nil
	}
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT id, user_id, window_type, timezone, starts_at, ends_at, weekday,
		       start_minute, end_minute, valid_from, valid_until, enabled, created_at, updated_at
		FROM availability_windows
		WHERE user_id = ANY($1::bigint[]) AND enabled = TRUE
		ORDER BY user_id, window_type, starts_at, weekday, start_minute, id
	`, userIDs)
	if err != nil {
		return nil, fmt.Errorf("list availability windows: %w", err)
	}
	defer rows.Close()

	windows := make([]AvailabilityWindow, 0)
	for rows.Next() {
		window, err := scanAvailabilityWindow(rows)
		if err != nil {
			return nil, err
		}
		windows = append(windows, window)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate availability windows: %w", err)
	}
	return windows, nil
}

func (s *Store) ListBusyIntervals(
	ctx context.Context,
	tx *sql.Tx,
	userIDs []int64,
	from time.Time,
	to time.Time,
	excludeCalendarEventID int64,
) ([]BusyInterval, error) {
	if len(userIDs) == 0 {
		return []BusyInterval{}, nil
	}
	rows, err := queryContext(executor(s.db, tx), ctx, `
		WITH requested_users AS (
			SELECT unnest($1::bigint[]) AS user_id
		)
		SELECT DISTINCT requested.user_id, event.starts_at, event.ends_at
		FROM requested_users requested
		JOIN calendar_events event ON event.status = 'active'
		  AND event.starts_at < $3 AND event.ends_at > $2
		  AND ($4 = 0 OR event.id <> $4)
		  AND (
		    event.organizer_id = requested.user_id
		    OR EXISTS (
		      SELECT 1 FROM calendar_event_attendees attendee
		      WHERE attendee.event_id = event.id
		        AND attendee.user_id = requested.user_id
		        AND attendee.status = 'accepted'
		    )
		  )
		ORDER BY requested.user_id, event.starts_at, event.ends_at
	`, userIDs, from, to, excludeCalendarEventID)
	if err != nil {
		return nil, fmt.Errorf("list busy intervals: %w", err)
	}
	defer rows.Close()

	intervals := make([]BusyInterval, 0)
	for rows.Next() {
		var interval BusyInterval
		if err := rows.Scan(&interval.UserID, &interval.StartsAt, &interval.EndsAt); err != nil {
			return nil, fmt.Errorf("scan busy interval: %w", err)
		}
		intervals = append(intervals, interval)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate busy intervals: %w", err)
	}
	return intervals, nil
}

func (s *Store) CreateMeetup(ctx context.Context, tx *sql.Tx, params CreateMeetupParams) (Meetup, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO meetups (
			context_type, context_id, organizer_id, title, category,
			starts_at, ends_at, location, status, version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'proposed', 1, $9, $9)
		RETURNING id, context_type, context_id, organizer_id, title, category,
		          starts_at, ends_at, location, status, version, calendar_event_id,
		          cancelled_by, cancelled_reason, created_at, updated_at
	`, params.ContextType, params.ContextID, params.OrganizerID, params.Title,
		params.Category, params.StartsAt, params.EndsAt, params.Location, params.Now)
	return scanMeetup(row)
}

func (s *Store) AddMeetupParticipant(
	ctx context.Context,
	tx *sql.Tx,
	meetupID int64,
	userID int64,
	status string,
	now time.Time,
) error {
	respondedAt := sql.NullTime{}
	if status != "pending" {
		respondedAt = sql.NullTime{Time: now, Valid: true}
	}
	_, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO meetup_participants (
			meetup_id, user_id, status, responded_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (meetup_id, user_id) DO UPDATE
		SET status = EXCLUDED.status,
		    responded_at = EXCLUDED.responded_at,
		    updated_at = EXCLUDED.updated_at
	`, meetupID, userID, status, nullableTime(respondedAt), now)
	if err != nil {
		return fmt.Errorf("add meetup participant: %w", err)
	}
	return nil
}

func (s *Store) GetMeetup(ctx context.Context, tx *sql.Tx, viewerID, meetupID int64) (Meetup, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, meetupSelect+`
		WHERE meetup.id = $2
		  AND EXISTS (
		    SELECT 1 FROM meetup_participants viewer
		    WHERE viewer.meetup_id = meetup.id AND viewer.user_id = $1
		  )
	`, viewerID, meetupID)
	return scanMeetup(row)
}

func (s *Store) GetMeetupForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	viewerID int64,
	meetupID int64,
) (Meetup, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, meetupSelect+`
		WHERE meetup.id = $2
		  AND EXISTS (
		    SELECT 1 FROM meetup_participants viewer
		    WHERE viewer.meetup_id = meetup.id AND viewer.user_id = $1
		  )
		FOR UPDATE
	`, viewerID, meetupID)
	return scanMeetup(row)
}

func (s *Store) ListMeetupsForUser(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	from time.Time,
	to time.Time,
	limit int,
) ([]Meetup, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, meetupSelect+`
		JOIN meetup_participants mine ON mine.meetup_id = meetup.id AND mine.user_id = $1
		WHERE meetup.starts_at < $3 AND meetup.ends_at > $2
		ORDER BY meetup.starts_at, meetup.id
		LIMIT $4
	`, userID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("list meetups for user: %w", err)
	}
	defer rows.Close()
	return scanMeetupRows(rows)
}

func (s *Store) ListMeetupsForContext(
	ctx context.Context,
	tx *sql.Tx,
	viewerID int64,
	contextType string,
	contextID int64,
	limit int,
) ([]Meetup, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, meetupSelect+`
		WHERE meetup.context_type = $2 AND meetup.context_id = $3
		  AND EXISTS (
		    SELECT 1 FROM meetup_participants viewer
		    WHERE viewer.meetup_id = meetup.id AND viewer.user_id = $1
		  )
		ORDER BY meetup.updated_at DESC, meetup.id DESC
		LIMIT $4
	`, viewerID, contextType, contextID, limit)
	if err != nil {
		return nil, fmt.Errorf("list meetups for context: %w", err)
	}
	defer rows.Close()
	return scanMeetupRows(rows)
}

func (s *Store) ListMeetupParticipants(
	ctx context.Context,
	tx *sql.Tx,
	meetupID int64,
) ([]MeetupParticipant, error) {
	rows, err := queryContext(executor(s.db, tx), ctx, `
		SELECT participant.meetup_id, participant.user_id,
		       user_account.nickname, user_account.department, user_account.grade,
		       participant.status, participant.responded_at,
		       participant.created_at, participant.updated_at
		FROM meetup_participants participant
		JOIN users user_account ON user_account.id = participant.user_id
		WHERE participant.meetup_id = $1 AND user_account.deleted_at IS NULL
		ORDER BY participant.created_at, participant.user_id
	`, meetupID)
	if err != nil {
		return nil, fmt.Errorf("list meetup participants: %w", err)
	}
	defer rows.Close()

	participants := make([]MeetupParticipant, 0)
	for rows.Next() {
		var participant MeetupParticipant
		if err := rows.Scan(
			&participant.MeetupID,
			&participant.UserID,
			&participant.Nickname,
			&participant.Department,
			&participant.Grade,
			&participant.Status,
			&participant.RespondedAt,
			&participant.CreatedAt,
			&participant.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan meetup participant: %w", err)
		}
		participants = append(participants, participant)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meetup participants: %w", err)
	}
	return participants, nil
}

func (s *Store) UpdateMeetup(ctx context.Context, tx *sql.Tx, params UpdateMeetupParams) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE meetups
		SET starts_at = $3,
		    ends_at = $4,
		    location = $5,
		    status = $6,
		    calendar_event_id = $7,
		    cancelled_by = $8,
		    cancelled_reason = $9,
		    version = version + 1,
		    updated_at = $10
		WHERE id = $1 AND version = $2
	`, params.MeetupID, params.ExpectedVersion, params.StartsAt, params.EndsAt,
		params.Location, params.Status, nullableInt64Value(params.CalendarEventID),
		nullableInt64Value(params.CancelledBy), params.CancelledReason, params.Now)
	if err != nil {
		return fmt.Errorf("update meetup: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) ResetMeetupParticipants(
	ctx context.Context,
	tx *sql.Tx,
	meetupID int64,
	proposerID int64,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE meetup_participants
		SET status = CASE WHEN user_id = $2 THEN 'accepted' ELSE 'pending' END,
		    responded_at = CASE WHEN user_id = $2 THEN $3 ELSE NULL END,
		    updated_at = $3
		WHERE meetup_id = $1
	`, meetupID, proposerID, now)
	if err != nil {
		return fmt.Errorf("reset meetup participants: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) UpdateMeetupParticipantStatus(
	ctx context.Context,
	tx *sql.Tx,
	meetupID int64,
	userID int64,
	status string,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE meetup_participants
		SET status = $3, responded_at = $4, updated_at = $4
		WHERE meetup_id = $1 AND user_id = $2 AND status = 'pending'
	`, meetupID, userID, status, now)
	if err != nil {
		return fmt.Errorf("update meetup participant status: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) GetMeetupFeedback(
	ctx context.Context,
	tx *sql.Tx,
	meetupID int64,
	userID int64,
) (MeetupFeedback, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT meetup_id, user_id, attendance, rating, would_meet_again,
		       safety_concern, private_notes, submitted_at, updated_at, revoked_at
		FROM meetup_feedback
		WHERE meetup_id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, meetupID, userID)
	return scanMeetupFeedback(row)
}

func (s *Store) UpsertMeetupFeedback(
	ctx context.Context,
	tx *sql.Tx,
	params UpsertMeetupFeedbackParams,
) (MeetupFeedback, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO meetup_feedback (
			meetup_id, user_id, attendance, rating, would_meet_again,
			safety_concern, private_notes, submitted_at, updated_at, revoked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, NULL)
		ON CONFLICT (meetup_id, user_id) DO UPDATE
		SET attendance = EXCLUDED.attendance,
		    rating = EXCLUDED.rating,
		    would_meet_again = EXCLUDED.would_meet_again,
		    safety_concern = EXCLUDED.safety_concern,
		    private_notes = EXCLUDED.private_notes,
		    updated_at = EXCLUDED.updated_at,
		    revoked_at = NULL
		RETURNING meetup_id, user_id, attendance, rating, would_meet_again,
		          safety_concern, private_notes, submitted_at, updated_at, revoked_at
	`, params.MeetupID, params.UserID, params.Attendance, nullableInt64Value(params.Rating),
		nullableBool(params.WouldMeetAgain), params.SafetyConcern, params.PrivateNotes, params.Now)
	return scanMeetupFeedback(row)
}

func (s *Store) RevokeMeetupFeedback(
	ctx context.Context,
	tx *sql.Tx,
	meetupID int64,
	userID int64,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE meetup_feedback
		SET revoked_at = $3, updated_at = $3
		WHERE meetup_id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, meetupID, userID, now)
	if err != nil {
		return fmt.Errorf("revoke meetup feedback: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) ReplaceMeetupLearningSignals(
	ctx context.Context,
	tx *sql.Tx,
	meetupID int64,
	userID int64,
	otherUserIDs []int64,
	signalType string,
	weight int,
	now time.Time,
) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE recommendation_learning_signals
		SET revoked_at = $3, updated_at = $3
		WHERE source_meetup_id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, meetupID, userID, now); err != nil {
		return fmt.Errorf("revoke previous meetup learning signals: %w", err)
	}
	for _, otherUserID := range otherUserIDs {
		if otherUserID == userID {
			continue
		}
		if _, err := executor(s.db, tx).ExecContext(ctx, `
			INSERT INTO recommendation_learning_signals (
				user_id, other_user_id, source_meetup_id, signal_type,
				weight, created_at, updated_at, revoked_at
			) VALUES ($1, $2, $3, $4, $5, $6, $6, NULL)
			ON CONFLICT (user_id, other_user_id, source_meetup_id, signal_type) DO UPDATE
			SET weight = EXCLUDED.weight, updated_at = EXCLUDED.updated_at, revoked_at = NULL
		`, userID, otherUserID, meetupID, signalType, weight, now); err != nil {
			return fmt.Errorf("upsert meetup learning signal: %w", err)
		}
	}
	return nil
}

func (s *Store) RevokeMeetupLearningSignals(
	ctx context.Context,
	tx *sql.Tx,
	meetupID int64,
	userID int64,
	now time.Time,
) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE recommendation_learning_signals
		SET revoked_at = $3, updated_at = $3
		WHERE source_meetup_id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, meetupID, userID, now); err != nil {
		return fmt.Errorf("revoke meetup learning signals: %w", err)
	}
	return nil
}

const meetupSelect = `
	SELECT meetup.id, meetup.context_type, meetup.context_id, meetup.organizer_id,
	       meetup.title, meetup.category, meetup.starts_at, meetup.ends_at,
	       meetup.location, meetup.status, meetup.version, meetup.calendar_event_id,
	       meetup.cancelled_by, meetup.cancelled_reason, meetup.created_at, meetup.updated_at
	FROM meetups meetup
`

func scanAvailabilityWindow(row rowScanner) (AvailabilityWindow, error) {
	var window AvailabilityWindow
	if err := row.Scan(
		&window.ID,
		&window.UserID,
		&window.WindowType,
		&window.Timezone,
		&window.StartsAt,
		&window.EndsAt,
		&window.Weekday,
		&window.StartMinute,
		&window.EndMinute,
		&window.ValidFrom,
		&window.ValidUntil,
		&window.Enabled,
		&window.CreatedAt,
		&window.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return AvailabilityWindow{}, ErrNotFound
		}
		return AvailabilityWindow{}, fmt.Errorf("scan availability window: %w", err)
	}
	return window, nil
}

func scanMeetup(row rowScanner) (Meetup, error) {
	var meetup Meetup
	if err := row.Scan(
		&meetup.ID,
		&meetup.ContextType,
		&meetup.ContextID,
		&meetup.OrganizerID,
		&meetup.Title,
		&meetup.Category,
		&meetup.StartsAt,
		&meetup.EndsAt,
		&meetup.Location,
		&meetup.Status,
		&meetup.Version,
		&meetup.CalendarEventID,
		&meetup.CancelledBy,
		&meetup.CancelledReason,
		&meetup.CreatedAt,
		&meetup.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return Meetup{}, ErrNotFound
		}
		return Meetup{}, fmt.Errorf("scan meetup: %w", err)
	}
	return meetup, nil
}

func scanMeetupRows(rows *sql.Rows) ([]Meetup, error) {
	meetups := make([]Meetup, 0)
	for rows.Next() {
		meetup, err := scanMeetup(rows)
		if err != nil {
			return nil, err
		}
		meetups = append(meetups, meetup)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meetups: %w", err)
	}
	return meetups, nil
}

func scanMeetupFeedback(row rowScanner) (MeetupFeedback, error) {
	var feedback MeetupFeedback
	if err := row.Scan(
		&feedback.MeetupID,
		&feedback.UserID,
		&feedback.Attendance,
		&feedback.Rating,
		&feedback.WouldMeetAgain,
		&feedback.SafetyConcern,
		&feedback.PrivateNotes,
		&feedback.SubmittedAt,
		&feedback.UpdatedAt,
		&feedback.RevokedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return MeetupFeedback{}, ErrNotFound
		}
		return MeetupFeedback{}, fmt.Errorf("scan meetup feedback: %w", err)
	}
	return feedback, nil
}

func nullableInt64Value(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}

func nullableBool(value sql.NullBool) any {
	if value.Valid {
		return value.Bool
	}
	return nil
}

func nullableDate(value sql.NullTime) any {
	if value.Valid {
		return value.Time.Format("2006-01-02")
	}
	return nil
}
