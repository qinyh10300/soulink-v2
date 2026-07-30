package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type MatchCycle struct {
	ID                   int64
	CycleKey             string
	Title                string
	QuestionnaireKey     string
	QuestionnaireVersion string
	RevealStatus         string
	RevealAt             time.Time
	ParticipantCount     int
	MatchedUserCount     int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type MatchResult struct {
	ID                  int64
	CycleID             int64
	UserID              int64
	ResultStatus        string
	PartnerUserID       sql.NullInt64
	MatchScore          sql.NullFloat64
	MatchReason         string
	MatchReasonJSON     json.RawMessage
	ResultMessage       string
	PartnerNickname     string
	PartnerDepartment   string
	PartnerGrade        string
	PartnerContactsJSON json.RawMessage
	RevealedAt          sql.NullTime
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type MatchResultRecord struct {
	MatchResult
	CycleKey             string
	CycleTitle           string
	QuestionnaireKey     string
	QuestionnaireVersion string
	RevealStatus         string
	CycleRevealAt        time.Time
	PartnerSchoolKey     string
	PartnerEmail         string
}

type UpsertMatchCycleParams struct {
	CycleKey             string
	Title                string
	QuestionnaireKey     string
	QuestionnaireVersion string
	RevealStatus         string
	RevealAt             time.Time
	ParticipantCount     int
	MatchedUserCount     int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type UpsertMatchResultParams struct {
	CycleID             int64
	UserID              int64
	ResultStatus        string
	PartnerUserID       *int64
	MatchScore          *float64
	MatchReason         string
	MatchReasonJSON     json.RawMessage
	ResultMessage       string
	PartnerNickname     string
	PartnerDepartment   string
	PartnerGrade        string
	PartnerContactsJSON json.RawMessage
	RevealedAt          *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// UpsertMatchCycle 保存或覆盖一个匹配周期。
func (s *Store) UpsertMatchCycle(ctx context.Context, tx *sql.Tx, params UpsertMatchCycleParams) (MatchCycle, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO match_cycles (
			cycle_key, title, questionnaire_key, questionnaire_version, reveal_status, reveal_at,
			participant_count, matched_user_count, created_at, updated_at
		) VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6, $7, $8, $9, $10)
		ON CONFLICT (cycle_key)
		DO UPDATE SET
			title = EXCLUDED.title,
			questionnaire_key = EXCLUDED.questionnaire_key,
			questionnaire_version = EXCLUDED.questionnaire_version,
			reveal_status = EXCLUDED.reveal_status,
			reveal_at = EXCLUDED.reveal_at,
			participant_count = EXCLUDED.participant_count,
			matched_user_count = EXCLUDED.matched_user_count,
			updated_at = EXCLUDED.updated_at
		RETURNING id, cycle_key, title, questionnaire_key, questionnaire_version, reveal_status, reveal_at,
		          participant_count, matched_user_count, created_at, updated_at
	`,
		params.CycleKey,
		params.Title,
		params.QuestionnaireKey,
		params.QuestionnaireVersion,
		params.RevealStatus,
		params.RevealAt,
		params.ParticipantCount,
		params.MatchedUserCount,
		params.CreatedAt,
		params.UpdatedAt,
	)

	cycle, err := scanMatchCycle(row)
	if err != nil {
		return MatchCycle{}, fmt.Errorf("upsert match cycle: %w", err)
	}
	return cycle, nil
}

// GetLatestMatchCycle 查询最新一个匹配周期，不区分是否已揭晓。
func (s *Store) GetLatestMatchCycle(ctx context.Context, tx *sql.Tx) (MatchCycle, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, cycle_key, title, questionnaire_key, questionnaire_version, reveal_status, reveal_at,
		       participant_count, matched_user_count, created_at, updated_at
		FROM match_cycles
		ORDER BY reveal_at DESC, id DESC
		LIMIT 1
	`)
	return scanMatchCycle(row)
}

// GetMatchCycleByID 根据主键查询匹配周期。
func (s *Store) GetMatchCycleByID(ctx context.Context, tx *sql.Tx, cycleID int64) (MatchCycle, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, cycle_key, title, questionnaire_key, questionnaire_version, reveal_status, reveal_at,
		       participant_count, matched_user_count, created_at, updated_at
		FROM match_cycles
		WHERE id = $1
		LIMIT 1
	`, cycleID)
	return scanMatchCycle(row)
}

// UpdateMatchCycleRevealStatus 更新匹配周期的揭晓状态。
func (s *Store) UpdateMatchCycleRevealStatus(ctx context.Context, tx *sql.Tx, cycleID int64, revealStatus string, now time.Time) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE match_cycles
		SET reveal_status = $2,
		    updated_at = $3
		WHERE id = $1
	`, cycleID, revealStatus, now); err != nil {
		return fmt.Errorf("update match cycle reveal status: %w", err)
	}
	return nil
}

// GetLatestVisibleMatchCycle 查询最近一个已揭晓、可向用户展示的匹配周期。
func (s *Store) GetLatestVisibleMatchCycle(ctx context.Context, tx *sql.Tx) (MatchCycle, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, cycle_key, title, questionnaire_key, questionnaire_version, reveal_status, reveal_at,
		       participant_count, matched_user_count, created_at, updated_at
		FROM match_cycles
		WHERE reveal_status IN ('revealed', 'archived')
		ORDER BY reveal_at DESC, id DESC
		LIMIT 1
	`)
	return scanMatchCycle(row)
}

// UpsertUserMatchResult 保存或覆盖用户在某个周期内的匹配结果。
func (s *Store) UpsertUserMatchResult(ctx context.Context, tx *sql.Tx, params UpsertMatchResultParams) (MatchResult, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO user_match_results (
			cycle_id, user_id, result_status, partner_user_id, match_score,
			match_reason, match_reason_json, result_message, partner_nickname, partner_department,
			partner_grade, partner_contacts_json, revealed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11, $12::jsonb, $13, $14, $15)
		ON CONFLICT (cycle_id, user_id)
		DO UPDATE SET
			result_status = EXCLUDED.result_status,
			partner_user_id = EXCLUDED.partner_user_id,
			match_score = EXCLUDED.match_score,
			match_reason = EXCLUDED.match_reason,
			match_reason_json = EXCLUDED.match_reason_json,
			result_message = EXCLUDED.result_message,
			partner_nickname = EXCLUDED.partner_nickname,
			partner_department = EXCLUDED.partner_department,
			partner_grade = EXCLUDED.partner_grade,
			partner_contacts_json = EXCLUDED.partner_contacts_json,
			revealed_at = EXCLUDED.revealed_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id, cycle_id, user_id, result_status, partner_user_id, match_score,
		          match_reason, match_reason_json, result_message, partner_nickname, partner_department,
		          partner_grade, partner_contacts_json, revealed_at, created_at, updated_at
	`,
		params.CycleID,
		params.UserID,
		params.ResultStatus,
		nullableInt64(params.PartnerUserID),
		nullableFloat64(params.MatchScore),
		params.MatchReason,
		nullableRawJSON(params.MatchReasonJSON),
		params.ResultMessage,
		params.PartnerNickname,
		params.PartnerDepartment,
		params.PartnerGrade,
		string(defaultJSON(params.PartnerContactsJSON)),
		nullableMatchRevealTime(params.RevealedAt),
		params.CreatedAt,
		params.UpdatedAt,
	)

	result, err := scanMatchResult(row)
	if err != nil {
		return MatchResult{}, fmt.Errorf("upsert user match result: %w", err)
	}
	return result, nil
}

// UpdateMatchResultRevealAt 更新指定用户在某个周期下的结果公布时间。
func (s *Store) UpdateMatchResultRevealAt(ctx context.Context, tx *sql.Tx, cycleID, userID int64, revealAt *time.Time, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE user_match_results
		SET revealed_at = $3,
		    updated_at = $4
		WHERE cycle_id = $1 AND user_id = $2
	`, cycleID, userID, nullableMatchRevealTime(revealAt), now)
	if err != nil {
		return fmt.Errorf("update match result reveal at: %w", err)
	}
	return requireRowsAffected(result)
}

// ListMatchResultsByCycle 查询指定周期下的全部用户匹配结果。
func (s *Store) ListMatchResultsByCycle(ctx context.Context, tx *sql.Tx, cycleID int64) ([]MatchResult, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT id, cycle_id, user_id, result_status, partner_user_id, match_score,
		       match_reason, match_reason_json, result_message, partner_nickname, partner_department,
		       partner_grade, partner_contacts_json, revealed_at, created_at, updated_at
		FROM user_match_results
		WHERE cycle_id = $1
		ORDER BY id ASC
	`, cycleID)
	if err != nil {
		return nil, fmt.Errorf("list match results by cycle: %w", err)
	}
	defer rows.Close()

	results := make([]MatchResult, 0)
	for rows.Next() {
		result, err := scanMatchResultCommon(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate match results by cycle: %w", err)
	}
	return results, nil
}

// DeleteMatchResultsByCycleExceptUsers 删除一个匹配周期中不在保留名单内的结果，并返回被删除的用户 ID。
func (s *Store) DeleteMatchResultsByCycleExceptUsers(ctx context.Context, tx *sql.Tx, cycleID int64, keepUserIDs []int64) ([]int64, error) {
	query := `
		DELETE FROM user_match_results
		WHERE cycle_id = $1
	`
	args := []any{cycleID}
	if len(keepUserIDs) > 0 {
		placeholders := make([]string, 0, len(keepUserIDs))
		for _, userID := range keepUserIDs {
			args = append(args, userID)
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		query += fmt.Sprintf(" AND user_id NOT IN (%s)", strings.Join(placeholders, ", "))
	}
	query += " RETURNING user_id"

	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("delete match results by cycle except users: %w", err)
	}
	defer rows.Close()

	deletedUserIDs := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan deleted match result user id: %w", err)
		}
		deletedUserIDs = append(deletedUserIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate deleted match result user ids: %w", err)
	}
	return deletedUserIDs, nil
}

// GetUserMatchResultByCycle 查询用户在指定周期内的匹配结果。
func (s *Store) GetUserMatchResultByCycle(ctx context.Context, tx *sql.Tx, userID, cycleID int64) (MatchResultRecord, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT r.id, r.cycle_id, r.user_id, r.result_status, r.partner_user_id, r.match_score,
		       r.match_reason, r.match_reason_json, r.result_message, r.partner_nickname, r.partner_department,
		       r.partner_grade, r.partner_contacts_json, r.revealed_at, r.created_at, r.updated_at,
		       c.cycle_key, c.title, c.questionnaire_key, c.questionnaire_version, c.reveal_status, c.reveal_at,
		       COALESCE(p.school_key, ''), COALESCE(p.email, '')
		FROM user_match_results r
		JOIN match_cycles c ON c.id = r.cycle_id
		LEFT JOIN users p ON p.id = r.partner_user_id
		WHERE r.user_id = $1 AND r.cycle_id = $2
		LIMIT 1
	`, userID, cycleID)
	return scanMatchResultRecord(row)
}

// GetUserMatchResultByID 查询用户的一条历史匹配结果。
func (s *Store) GetUserMatchResultByID(ctx context.Context, tx *sql.Tx, userID, matchID int64) (MatchResultRecord, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT r.id, r.cycle_id, r.user_id, r.result_status, r.partner_user_id, r.match_score,
		       r.match_reason, r.match_reason_json, r.result_message, r.partner_nickname, r.partner_department,
		       r.partner_grade, r.partner_contacts_json, r.revealed_at, r.created_at, r.updated_at,
		       c.cycle_key, c.title, c.questionnaire_key, c.questionnaire_version, c.reveal_status, c.reveal_at,
		       COALESCE(p.school_key, ''), COALESCE(p.email, '')
		FROM user_match_results r
		JOIN match_cycles c ON c.id = r.cycle_id
		LEFT JOIN users p ON p.id = r.partner_user_id
		WHERE r.user_id = $1 AND r.id = $2
		LIMIT 1
	`, userID, matchID)
	return scanMatchResultRecord(row)
}

// ListUserMatchResults 查询用户的历史匹配结果列表。
func (s *Store) ListUserMatchResults(ctx context.Context, tx *sql.Tx, userID int64, limit int) ([]MatchResultRecord, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT r.id, r.cycle_id, r.user_id, r.result_status, r.partner_user_id, r.match_score,
		       r.match_reason, r.match_reason_json, r.result_message, r.partner_nickname, r.partner_department,
		       r.partner_grade, r.partner_contacts_json, r.revealed_at, r.created_at, r.updated_at,
		       c.cycle_key, c.title, c.questionnaire_key, c.questionnaire_version, c.reveal_status, c.reveal_at,
		       COALESCE(p.school_key, ''), COALESCE(p.email, '')
		FROM user_match_results r
		JOIN match_cycles c ON c.id = r.cycle_id
		LEFT JOIN users p ON p.id = r.partner_user_id
		WHERE r.user_id = $1
		ORDER BY c.reveal_at DESC, r.id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list user match results: %w", err)
	}
	defer rows.Close()

	results := make([]MatchResultRecord, 0)
	for rows.Next() {
		result, err := scanMatchResultRecordRows(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user match results: %w", err)
	}
	return results, nil
}

// DeleteUserMatchResultsByUserReference 删除与指定用户相关的全部匹配结果，并同步刷新受影响周期的统计。
func (s *Store) DeleteUserMatchResultsByUserReference(ctx context.Context, tx *sql.Tx, userID int64, now time.Time) error {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		DELETE FROM user_match_results
		WHERE user_id = $1 OR partner_user_id = $1
		RETURNING cycle_id
	`, userID)
	if err != nil {
		return fmt.Errorf("delete user match results by user reference: %w", err)
	}
	defer rows.Close()

	affectedCycles := make(map[int64]struct{})
	for rows.Next() {
		var cycleID int64
		if err := rows.Scan(&cycleID); err != nil {
			return fmt.Errorf("scan deleted user match result cycle id: %w", err)
		}
		affectedCycles[cycleID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate deleted user match result cycle ids: %w", err)
	}

	for cycleID := range affectedCycles {
		if _, err := executor(s.db, tx).ExecContext(ctx, `
			UPDATE match_cycles
			SET participant_count = COALESCE((
				SELECT COUNT(*)
				FROM user_match_results
				WHERE cycle_id = $1
			), 0),
				matched_user_count = COALESCE((
					SELECT COUNT(*)
					FROM user_match_results
					WHERE cycle_id = $1 AND result_status = 'matched'
				), 0),
				updated_at = $2
			WHERE id = $1
		`, cycleID, now); err != nil {
			return fmt.Errorf("refresh match cycle stats: %w", err)
		}
	}
	return nil
}

func scanMatchCycle(row *sql.Row) (MatchCycle, error) {
	var cycle MatchCycle
	var questionnaireKey sql.NullString
	var questionnaireVersion sql.NullString
	if err := row.Scan(
		&cycle.ID,
		&cycle.CycleKey,
		&cycle.Title,
		&questionnaireKey,
		&questionnaireVersion,
		&cycle.RevealStatus,
		&cycle.RevealAt,
		&cycle.ParticipantCount,
		&cycle.MatchedUserCount,
		&cycle.CreatedAt,
		&cycle.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return MatchCycle{}, ErrNotFound
		}
		return MatchCycle{}, fmt.Errorf("scan match cycle: %w", err)
	}
	cycle.QuestionnaireKey = nullableString(questionnaireKey)
	cycle.QuestionnaireVersion = nullableString(questionnaireVersion)
	return cycle, nil
}

func scanMatchResult(row *sql.Row) (MatchResult, error) {
	result, err := scanMatchResultCommon(row)
	if err != nil {
		return MatchResult{}, err
	}
	return result, nil
}

func scanMatchResultRecord(row *sql.Row) (MatchResultRecord, error) {
	result, err := scanMatchResultRecordCommon(row)
	if err != nil {
		return MatchResultRecord{}, err
	}
	return result, nil
}

func scanMatchResultRecordRows(rows *sql.Rows) (MatchResultRecord, error) {
	result, err := scanMatchResultRecordCommon(rows)
	if err != nil {
		return MatchResultRecord{}, err
	}
	return result, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMatchResultCommon(scanner rowScanner) (MatchResult, error) {
	var result MatchResult
	var partnerContacts []byte
	var matchReasonJSON []byte
	var partnerNickname sql.NullString
	var partnerDepartment sql.NullString
	var partnerGrade sql.NullString
	if err := scanner.Scan(
		&result.ID,
		&result.CycleID,
		&result.UserID,
		&result.ResultStatus,
		&result.PartnerUserID,
		&result.MatchScore,
		&result.MatchReason,
		&matchReasonJSON,
		&result.ResultMessage,
		&partnerNickname,
		&partnerDepartment,
		&partnerGrade,
		&partnerContacts,
		&result.RevealedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return MatchResult{}, ErrNotFound
		}
		return MatchResult{}, fmt.Errorf("scan match result: %w", err)
	}
	result.PartnerNickname = nullableString(partnerNickname)
	result.PartnerDepartment = nullableString(partnerDepartment)
	result.PartnerGrade = nullableString(partnerGrade)
	result.MatchReasonJSON = append(json.RawMessage(nil), matchReasonJSON...)
	result.PartnerContactsJSON = append(json.RawMessage(nil), partnerContacts...)
	return result, nil
}

func scanMatchResultRecordCommon(scanner rowScanner) (MatchResultRecord, error) {
	var result MatchResultRecord
	var partnerContacts []byte
	var matchReasonJSON []byte
	var partnerNickname sql.NullString
	var partnerDepartment sql.NullString
	var partnerGrade sql.NullString
	var questionnaireKey sql.NullString
	var questionnaireVersion sql.NullString
	var partnerSchoolKey sql.NullString
	var partnerEmail sql.NullString
	if err := scanner.Scan(
		&result.ID,
		&result.CycleID,
		&result.UserID,
		&result.ResultStatus,
		&result.PartnerUserID,
		&result.MatchScore,
		&result.MatchReason,
		&matchReasonJSON,
		&result.ResultMessage,
		&partnerNickname,
		&partnerDepartment,
		&partnerGrade,
		&partnerContacts,
		&result.RevealedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CycleKey,
		&result.CycleTitle,
		&questionnaireKey,
		&questionnaireVersion,
		&result.RevealStatus,
		&result.CycleRevealAt,
		&partnerSchoolKey,
		&partnerEmail,
	); err != nil {
		if err == sql.ErrNoRows {
			return MatchResultRecord{}, ErrNotFound
		}
		return MatchResultRecord{}, fmt.Errorf("scan match result record: %w", err)
	}
	result.PartnerNickname = nullableString(partnerNickname)
	result.PartnerDepartment = nullableString(partnerDepartment)
	result.PartnerGrade = nullableString(partnerGrade)
	result.MatchReasonJSON = append(json.RawMessage(nil), matchReasonJSON...)
	result.PartnerContactsJSON = append(json.RawMessage(nil), partnerContacts...)
	result.QuestionnaireKey = nullableString(questionnaireKey)
	result.QuestionnaireVersion = nullableString(questionnaireVersion)
	result.PartnerSchoolKey = nullableString(partnerSchoolKey)
	result.PartnerEmail = nullableString(partnerEmail)
	return result, nil
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableFloat64(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableMatchRevealTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableRawJSON(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}

func defaultJSON(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage(`[]`)
	}
	return value
}
