package store

import (
	"context"
	"database/sql"
	"encoding/json"
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
	ConversationID       sql.NullInt64
	GroupID              sql.NullInt64
	ContextTitle         string
	Category             string
	Details              string
	ContextSnapshot      json.RawMessage
	RiskSignals          json.RawMessage
	Status               string
	AssignedTo           sql.NullInt64
	AssignedToNickname   string
	ReviewNotes          string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ResolvedAt           sql.NullTime
	ResolvedBy           sql.NullInt64
}

type CreateSafetyReportParams struct {
	ReporterID      int64
	ReportedUserID  int64
	IntentID        sql.NullInt64
	RoomID          sql.NullInt64
	ConversationID  sql.NullInt64
	GroupID         sql.NullInt64
	Category        string
	Details         string
	ContextSnapshot json.RawMessage
	RiskSignals     json.RawMessage
	Now             time.Time
}

type SafetyReportAppeal struct {
	ID                int64
	ReportID          int64
	AppellantID       int64
	AppellantNickname string
	Reason            string
	Status            string
	ReviewNotes       string
	ReviewedBy        sql.NullInt64
	ReviewedAt        sql.NullTime
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UpdateSafetyReportParams struct {
	ReportID    int64
	Status      string
	ReviewNotes string
	AssignToMe  bool
	AdminID     int64
	Now         time.Time
}

func (s *Store) CreateSafetyReport(
	ctx context.Context,
	tx *sql.Tx,
	params CreateSafetyReportParams,
) (int64, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO safety_reports (
			reporter_id, reported_user_id, intent_id, room_id, conversation_id, group_id,
			category, details, context_snapshot, risk_signals, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'open', $11, $11)
		RETURNING id
	`, params.ReporterID, params.ReportedUserID, nullableSQLInt64(params.IntentID),
		nullableSQLInt64(params.RoomID), nullableSQLInt64(params.ConversationID),
		nullableSQLInt64(params.GroupID), params.Category, params.Details,
		params.ContextSnapshot, params.RiskSignals, params.Now)

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

func (s *Store) ListSafetyReportsForReportedUser(
	ctx context.Context,
	tx *sql.Tx,
	userID int64,
	limit int,
) ([]SafetyReport, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := queryContext(executor(s.db, tx), ctx, safetyReportSelect+`
		WHERE r.reported_user_id = $1
		  AND r.status IN ('resolved', 'dismissed')
		ORDER BY r.updated_at DESC, r.id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list safety cases: %w", err)
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
	return reports, rows.Err()
}

func (s *Store) UpdateSafetyReport(
	ctx context.Context,
	tx *sql.Tx,
	params UpdateSafetyReportParams,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE safety_reports
		SET status = $2::varchar,
		    review_notes = $3::text,
		    assigned_to = CASE
				WHEN $4::boolean THEN $5::bigint
				WHEN assigned_to IS NULL AND $2::varchar <> 'open' THEN $5::bigint
				ELSE assigned_to
			END,
		    resolved_at = CASE
					WHEN $2::varchar IN ('resolved', 'dismissed') THEN $6::timestamptz
					ELSE NULL::timestamptz
				END,
		    resolved_by = CASE
					WHEN $2::varchar IN ('resolved', 'dismissed') THEN $5::bigint
					ELSE NULL::bigint
				END,
		    updated_at = $6::timestamptz
		WHERE id = $1
	`, params.ReportID, params.Status, params.ReviewNotes, params.AssignToMe, params.AdminID, params.Now)
	if err != nil {
		return fmt.Errorf("update safety report: %w", err)
	}
	return requireRowsAffected(result)
}

func (s *Store) CreateSafetyReportAppeal(
	ctx context.Context,
	tx *sql.Tx,
	reportID, appellantID int64,
	reason string,
	now time.Time,
) (SafetyReportAppeal, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		WITH inserted AS (
		  INSERT INTO safety_report_appeals (
		    report_id, appellant_id, reason, status, created_at, updated_at
		  ) VALUES ($1, $2, $3, 'pending', $4, $4)
		  RETURNING id, report_id, appellant_id, reason, status, review_notes,
		            reviewed_by, reviewed_at, created_at, updated_at
		)
		SELECT inserted.id, inserted.report_id, inserted.appellant_id, appellant.nickname,
		       inserted.reason, inserted.status, inserted.review_notes,
		       inserted.reviewed_by, inserted.reviewed_at,
		       inserted.created_at, inserted.updated_at
		FROM inserted
		JOIN users appellant ON appellant.id = inserted.appellant_id
	`, reportID, appellantID, reason, now)
	return scanSafetyReportAppeal(row)
}

func (s *Store) GetSafetyReportAppeal(
	ctx context.Context,
	tx *sql.Tx,
	appealID int64,
) (SafetyReportAppeal, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, safetyReportAppealSelect+`
		WHERE appeal.id = $1
	`, appealID)
	return scanSafetyReportAppeal(row)
}

func (s *Store) GetLatestSafetyReportAppeal(
	ctx context.Context,
	tx *sql.Tx,
	reportID int64,
) (SafetyReportAppeal, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, safetyReportAppealSelect+`
		WHERE appeal.report_id = $1
		ORDER BY appeal.created_at DESC, appeal.id DESC
		LIMIT 1
	`, reportID)
	return scanSafetyReportAppeal(row)
}

func (s *Store) ReviewSafetyReportAppeal(
	ctx context.Context,
	tx *sql.Tx,
	appealID, adminID int64,
	status, notes string,
	now time.Time,
) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE safety_report_appeals
		SET status = $3, review_notes = $4, reviewed_by = $2,
		    reviewed_at = $5, updated_at = $5
		WHERE id = $1 AND status = 'pending'
	`, appealID, adminID, status, notes, now)
	if err != nil {
		return fmt.Errorf("review safety report appeal: %w", err)
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

func (s *Store) UsersHaveActiveSafetyReport(
	ctx context.Context,
	tx *sql.Tx,
	leftUserID int64,
	rightUserID int64,
) (bool, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM safety_reports report
			WHERE report.status <> 'dismissed'
			  AND (
			    (report.reporter_id = $1 AND report.reported_user_id = $2)
			    OR (report.reporter_id = $2 AND report.reported_user_id = $1)
			  )
		)
	`, leftUserID, rightUserID)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("check active safety reports: %w", err)
	}
	return exists, nil
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
	       r.intent_id, r.room_id, r.conversation_id, r.group_id,
	       COALESCE(i.title, room_intent.title, chat_group.name,
	                CASE WHEN r.conversation_id IS NOT NULL THEN '一对一会话' END, ''),
	       r.category, r.details, r.context_snapshot, r.risk_signals, r.status,
		       r.assigned_to, COALESCE(assignee.nickname, ''), r.review_notes,
		       r.created_at, r.updated_at, r.resolved_at, r.resolved_by
	FROM safety_reports r
	JOIN users reporter ON reporter.id = r.reporter_id
	JOIN users reported ON reported.id = r.reported_user_id
	LEFT JOIN connection_intents i ON i.id = r.intent_id
	LEFT JOIN match_rooms room ON room.id = r.room_id
	LEFT JOIN connection_intents room_intent ON room_intent.id = room.intent_id
	LEFT JOIN chat_groups chat_group ON chat_group.id = r.group_id
	LEFT JOIN users assignee ON assignee.id = r.assigned_to
`

const safetyReportAppealSelect = `
	SELECT appeal.id, appeal.report_id, appeal.appellant_id, appellant.nickname,
	       appeal.reason, appeal.status, appeal.review_notes, appeal.reviewed_by,
	       appeal.reviewed_at, appeal.created_at, appeal.updated_at
	FROM safety_report_appeals appeal
	JOIN users appellant ON appellant.id = appeal.appellant_id
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
		&report.ConversationID,
		&report.GroupID,
		&report.ContextTitle,
		&report.Category,
		&report.Details,
		&report.ContextSnapshot,
		&report.RiskSignals,
		&report.Status,
		&report.AssignedTo,
		&report.AssignedToNickname,
		&report.ReviewNotes,
		&report.CreatedAt,
		&report.UpdatedAt,
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

func scanSafetyReportAppeal(row rowScanner) (SafetyReportAppeal, error) {
	var appeal SafetyReportAppeal
	if err := row.Scan(
		&appeal.ID, &appeal.ReportID, &appeal.AppellantID,
		&appeal.AppellantNickname, &appeal.Reason, &appeal.Status,
		&appeal.ReviewNotes, &appeal.ReviewedBy, &appeal.ReviewedAt,
		&appeal.CreatedAt, &appeal.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return SafetyReportAppeal{}, ErrNotFound
		}
		return SafetyReportAppeal{}, fmt.Errorf("scan safety report appeal: %w", err)
	}
	return appeal, nil
}

func nullableSQLInt64(value sql.NullInt64) any {
	if !value.Valid {
		return nil
	}
	return value.Int64
}
