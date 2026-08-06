package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type CalendarEvent struct {
	ID                   int64
	OrganizerID          int64
	OrganizerNickname    string
	OrganizerDepartment  string
	EventType            string
	Title                string
	Description          string
	Location             string
	StartsAt             time.Time
	EndsAt               time.Time
	Visibility           string
	Status               string
	ViewerIsParticipant  bool
	ViewerAttendeeStatus sql.NullString
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type CalendarAttendee struct {
	EventID     int64
	UserID      int64
	Nickname    string
	Department  string
	Grade       string
	Status      string
	RespondedAt sql.NullTime
	CreatedAt   time.Time
}

type CreateCalendarEventParams struct {
	OrganizerID int64
	EventType   string
	Title       string
	Description string
	Location    string
	StartsAt    time.Time
	EndsAt      time.Time
	Visibility  string
	Now         time.Time
}

type UpdateCalendarEventParams struct {
	EventID     int64
	EventType   string
	Title       string
	Description string
	Location    string
	StartsAt    time.Time
	EndsAt      time.Time
	Visibility  string
	Now         time.Time
}

func (s *Store) CreateCalendarEvent(ctx context.Context, tx *sql.Tx, params CreateCalendarEventParams) (CalendarEvent, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO calendar_events (
			organizer_id, event_type, title, description, location,
			starts_at, ends_at, visibility, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active', $9, $9)
		RETURNING id
	`, params.OrganizerID, params.EventType, params.Title, params.Description, params.Location,
		params.StartsAt, params.EndsAt, params.Visibility, params.Now)
	var eventID int64
	if err := row.Scan(&eventID); err != nil {
		return CalendarEvent{}, fmt.Errorf("create calendar event: %w", err)
	}
	return s.GetCalendarEvent(ctx, tx, params.OrganizerID, eventID)
}

func (s *Store) GetCalendarEvent(ctx context.Context, tx *sql.Tx, viewerID, eventID int64) (CalendarEvent, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, calendarEventSelect+`
		WHERE e.id = $2 AND e.status = 'active' AND organizer.deleted_at IS NULL
	`, viewerID, eventID)
	return scanCalendarEvent(row)
}

func (s *Store) FindCalendarEvent(ctx context.Context, tx *sql.Tx, organizerID int64, title string) (CalendarEvent, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, calendarEventSelect+`
		WHERE e.organizer_id = $1 AND e.title = $2 AND e.status = 'active'
		  AND organizer.deleted_at IS NULL
		ORDER BY e.id
		LIMIT 1
	`, organizerID, title)
	return scanCalendarEvent(row)
}

func (s *Store) ListCalendarEvents(ctx context.Context, tx *sql.Tx, userID int64, from, to time.Time) ([]CalendarEvent, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, calendarEventSelect+`
		WHERE e.status = 'active'
		  AND e.starts_at < $3 AND e.ends_at > $2
		  AND organizer.deleted_at IS NULL
		  AND (
		    e.organizer_id = $1
		    OR EXISTS (
		      SELECT 1 FROM calendar_event_attendees mine
		      WHERE mine.event_id = e.id AND mine.user_id = $1 AND mine.status != 'declined'
		    )
		  )
		ORDER BY e.starts_at, e.id
	`, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list calendar events: %w", err)
	}
	defer rows.Close()
	return scanCalendarEventRows(rows)
}

func (s *Store) ListUserCalendarAvailability(ctx context.Context, tx *sql.Tx, viewerID, targetUserID int64, from, to time.Time) ([]CalendarEvent, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, calendarEventSelect+`
		WHERE e.status = 'active'
		  AND e.starts_at < $4 AND e.ends_at > $3
		  AND organizer.deleted_at IS NULL
		  AND (
		    e.organizer_id = $2
		    OR EXISTS (
		      SELECT 1 FROM calendar_event_attendees target_attendee
		      WHERE target_attendee.event_id = e.id
		        AND target_attendee.user_id = $2
		        AND target_attendee.status = 'accepted'
		    )
		  )
		ORDER BY e.starts_at, e.id
	`, viewerID, targetUserID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list user calendar availability: %w", err)
	}
	defer rows.Close()
	return scanCalendarEventRows(rows)
}

func (s *Store) UpdateCalendarEvent(ctx context.Context, tx *sql.Tx, viewerID int64, params UpdateCalendarEventParams) (CalendarEvent, error) {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE calendar_events
		SET event_type = $2, title = $3, description = $4, location = $5,
		    starts_at = $6, ends_at = $7, visibility = $8, updated_at = $9
		WHERE id = $1 AND status = 'active'
	`, params.EventID, params.EventType, params.Title, params.Description, params.Location,
		params.StartsAt, params.EndsAt, params.Visibility, params.Now)
	if err != nil {
		return CalendarEvent{}, fmt.Errorf("update calendar event: %w", err)
	}
	if err := requireRowsAffected(result); err != nil {
		return CalendarEvent{}, err
	}
	return s.GetCalendarEvent(ctx, tx, viewerID, params.EventID)
}

func (s *Store) CancelCalendarEvent(ctx context.Context, tx *sql.Tx, eventID int64, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE calendar_events
		SET status = 'cancelled', updated_at = $2
		WHERE id = $1 AND status = 'active'
	`, eventID, now)
	if err != nil {
		return fmt.Errorf("cancel calendar event: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) AddCalendarAttendee(ctx context.Context, tx *sql.Tx, eventID, userID int64, now time.Time) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO calendar_event_attendees (event_id, user_id, status, created_at)
		VALUES ($1, $2, 'pending', $3)
		ON CONFLICT (event_id, user_id) DO UPDATE
		SET status = CASE
		      WHEN calendar_event_attendees.status = 'declined' THEN 'pending'
		      ELSE calendar_event_attendees.status
		    END,
		    responded_at = CASE
		      WHEN calendar_event_attendees.status = 'declined' THEN NULL
		      ELSE calendar_event_attendees.responded_at
		    END
	`, eventID, userID, now)
	if err != nil {
		return fmt.Errorf("add calendar attendee: %w", err)
	}
	return nil
}

func (s *Store) ListCalendarAttendees(ctx context.Context, tx *sql.Tx, eventID int64) ([]CalendarAttendee, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT a.event_id, a.user_id, u.nickname, u.department, u.grade,
		       a.status, a.responded_at, a.created_at
		FROM calendar_event_attendees a
		JOIN users u ON u.id = a.user_id
		WHERE a.event_id = $1 AND u.deleted_at IS NULL
		ORDER BY a.created_at, a.user_id
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("list calendar attendees: %w", err)
	}
	defer rows.Close()

	attendees := make([]CalendarAttendee, 0)
	for rows.Next() {
		var attendee CalendarAttendee
		if err := rows.Scan(
			&attendee.EventID,
			&attendee.UserID,
			&attendee.Nickname,
			&attendee.Department,
			&attendee.Grade,
			&attendee.Status,
			&attendee.RespondedAt,
			&attendee.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan calendar attendee row: %w", err)
		}
		attendees = append(attendees, attendee)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate calendar attendees: %w", err)
	}
	return attendees, nil
}

func (s *Store) UpdateCalendarAttendeeStatus(ctx context.Context, tx *sql.Tx, eventID, userID int64, status string, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE calendar_event_attendees
		SET status = $3, responded_at = $4
		WHERE event_id = $1 AND user_id = $2
	`, eventID, userID, status, now)
	if err != nil {
		return fmt.Errorf("update calendar attendee status: %w", err)
	}
	return requireRowsAffected(result)
}

const calendarEventSelect = `
	SELECT e.id, e.organizer_id, organizer.nickname, organizer.department,
	       e.event_type, e.title, e.description, e.location,
	       e.starts_at, e.ends_at, e.visibility, e.status,
	       (e.organizer_id = $1 OR EXISTS (
	          SELECT 1 FROM calendar_event_attendees viewer_attendee
	          WHERE viewer_attendee.event_id = e.id
	            AND viewer_attendee.user_id = $1
	            AND viewer_attendee.status != 'declined'
	       )) AS viewer_is_participant,
	       (
	          SELECT viewer_attendee.status
	          FROM calendar_event_attendees viewer_attendee
	          WHERE viewer_attendee.event_id = e.id AND viewer_attendee.user_id = $1
	       ) AS viewer_attendee_status,
	       e.created_at, e.updated_at
	FROM calendar_events e
	JOIN users organizer ON organizer.id = e.organizer_id
`

func scanCalendarEvent(row *sql.Row) (CalendarEvent, error) {
	var event CalendarEvent
	if err := row.Scan(
		&event.ID,
		&event.OrganizerID,
		&event.OrganizerNickname,
		&event.OrganizerDepartment,
		&event.EventType,
		&event.Title,
		&event.Description,
		&event.Location,
		&event.StartsAt,
		&event.EndsAt,
		&event.Visibility,
		&event.Status,
		&event.ViewerIsParticipant,
		&event.ViewerAttendeeStatus,
		&event.CreatedAt,
		&event.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return CalendarEvent{}, ErrNotFound
		}
		return CalendarEvent{}, fmt.Errorf("scan calendar event: %w", err)
	}
	return event, nil
}

func scanCalendarEventRows(rows *sql.Rows) ([]CalendarEvent, error) {
	events := make([]CalendarEvent, 0)
	for rows.Next() {
		var event CalendarEvent
		if err := rows.Scan(
			&event.ID,
			&event.OrganizerID,
			&event.OrganizerNickname,
			&event.OrganizerDepartment,
			&event.EventType,
			&event.Title,
			&event.Description,
			&event.Location,
			&event.StartsAt,
			&event.EndsAt,
			&event.Visibility,
			&event.Status,
			&event.ViewerIsParticipant,
			&event.ViewerAttendeeStatus,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan calendar event row: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate calendar events: %w", err)
	}
	return events, nil
}
