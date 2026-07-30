package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SafetyReport struct {
	ID                   int64
	ReporterID           int64
	ReporterNickname     string
	ReporterEmail        string
	ReportedUserID       int64
	ReportedUserNickname string
	ReportedUserEmail    string
	IntentID             sql.NullInt64
	RoomID               sql.NullInt64
	ContextTitle         string
	Category             string
	Details              string
	Status               string
	CreatedAt            time.Time
	ResolvedAt           sql.NullTime
	ResolvedBy           sql.NullInt64
}

type CreateSafetyReportParams struct {
	ReporterID     int64
	ReportedUserID int64
	IntentID       sql.NullInt64
	RoomID         sql.NullInt64
	Category       string
	Details        string
	Now            time.Time
}

func (s *Store) CreateSafetyReport(
	ctx context.Context,
	tx *sql.Tx,
	params CreateSafetyReportParams,
) (int64, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO safety_reports (
			reporter_id, reported_user_id, intent_id, room_id,
			category, details, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'open', $7)
		RETURNING id
	`, params.ReporterID, params.ReportedUserID, nullableSQLInt64(params.IntentID),
		nullableSQLInt64(params.RoomID), params.Category, params.Details, params.Now)

	var reportID int64
	if err := row.Scan(&reportID); err != nil {
		return 0, fmt.Errorf("create safety report: %w", err)
	}
	return reportID, nil
}

func (s *Store) GetSafetyReport(
	ctx context.Context,
	tx *sql.Tx,
	reportID int64,
) (SafetyReport, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, safetyReportSelect+`
		WHERE r.id = $1
	`, reportID)
	return scanSafetyReport(row)
}

func (s *Store) ListSafetyReports(
	ctx context.Context,
	tx *sql.Tx,
	status string,
	limit int,
) ([]SafetyReport, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := queryContext(executor(s.db, tx), ctx, safetyReportSelect+`
		WHERE ($1 = '' OR r.status = $1)
		ORDER BY
			CASE r.status WHEN 'open' THEN 0 WHEN 'reviewing' THEN 1 ELSE 2 END,
			r.created_at DESC,
			r.id DESC
		LIMIT $2
	`, status, limit)
	if err != nil {
		return nil, fmt.Errorf("list safety reports: %w", err)
	}
	defer rows.Close()

	reports := make([]SafetyReport, 0)
	for rows.Next() {
		report, err := scanSafetyReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate safety reports: %w", err)
	}
	return reports, nil
}

func (s *Store) UpdateSafetyReportStatus(
	ctx context.Context,
	tx *sql.Tx,
	reportID int64,
	status string,
	resolvedBy int64,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE safety_reports
		SET status = $2,
		    resolved_at = CASE
				WHEN $2 IN ('resolved', 'dismissed') THEN $4
				ELSE NULL
			END,
		    resolved_by = CASE
				WHEN $2 IN ('resolved', 'dismissed') THEN $3
				ELSE NULL
			END
		WHERE id = $1
	`, reportID, status, resolvedBy, now)
	if err != nil {
		return fmt.Errorf("update safety report status: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) UpsertUserBlock(
	ctx context.Context,
	tx *sql.Tx,
	blockerID int64,
	blockedID int64,
	reason string,
	now time.Time,
) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		INSERT INTO user_blocks (blocker_id, blocked_id, reason, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (blocker_id, blocked_id) DO UPDATE SET
			reason = EXCLUDED.reason
	`, blockerID, blockedID, reason, now)
	if err != nil {
		return fmt.Errorf("upsert user block: %w", err)
	}
	return nil
}

func (s *Store) UsersBlocked(
	ctx context.Context,
	tx *sql.Tx,
	firstUserID int64,
	secondUserID int64,
) (bool, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_blocks
			WHERE (blocker_id = $1 AND blocked_id = $2)
			   OR (blocker_id = $2 AND blocked_id = $1)
		)
	`, firstUserID, secondUserID)

	var blocked bool
	if err := row.Scan(&blocked); err != nil {
		return false, fmt.Errorf("check user block: %w", err)
	}
	return blocked, nil
}

func (s *Store) CloseMatchRoomsBetweenUsers(
	ctx context.Context,
	tx *sql.Tx,
	firstUserID int64,
	secondUserID int64,
	now time.Time,
) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE match_rooms
		SET status = 'closed',
		    owner_contact_shared = FALSE,
		    candidate_contact_shared = FALSE,
		    updated_at = $3
		WHERE status = 'open'
		  AND (
			(owner_id = $1 AND candidate_id = $2)
			OR (owner_id = $2 AND candidate_id = $1)
		  )
	`, firstUserID, secondUserID, now)
	if err != nil {
		return fmt.Errorf("close blocked match rooms: %w", err)
	}
	return nil
}

const safetyReportSelect = `
	SELECT r.id,
	       r.reporter_id, reporter.nickname, reporter.email,
	       r.reported_user_id, reported.nickname, reported.email,
	       r.intent_id, r.room_id,
	       COALESCE(i.title, room_intent.title, ''),
	       r.category, r.details, r.status, r.created_at, r.resolved_at, r.resolved_by
	FROM safety_reports r
	JOIN users reporter ON reporter.id = r.reporter_id
	JOIN users reported ON reported.id = r.reported_user_id
	LEFT JOIN connection_intents i ON i.id = r.intent_id
	LEFT JOIN match_rooms room ON room.id = r.room_id
	LEFT JOIN connection_intents room_intent ON room_intent.id = room.intent_id
`

func scanSafetyReport(row rowScanner) (SafetyReport, error) {
	var report SafetyReport
	if err := row.Scan(
		&report.ID,
		&report.ReporterID,
		&report.ReporterNickname,
		&report.ReporterEmail,
		&report.ReportedUserID,
		&report.ReportedUserNickname,
		&report.ReportedUserEmail,
		&report.IntentID,
		&report.RoomID,
		&report.ContextTitle,
		&report.Category,
		&report.Details,
		&report.Status,
		&report.CreatedAt,
		&report.ResolvedAt,
		&report.ResolvedBy,
	); err != nil {
		if err == sql.ErrNoRows {
			return SafetyReport{}, ErrNotFound
		}
		return SafetyReport{}, fmt.Errorf("scan safety report: %w", err)
	}
	return report, nil
}

func nullableSQLInt64(value sql.NullInt64) any {
	if !value.Valid {
		return nil
	}
	return value.Int64
}
