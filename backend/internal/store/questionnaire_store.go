package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type QuestionnaireAnswer struct {
	ID                   int64
	UserID               int64
	QuestionnaireKey     string
	QuestionnaireVersion string
	QuestionID           string
	ValueJSON            json.RawMessage
	IsImportant          bool
	AnsweredAt           time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type QuestionnaireState struct {
	UserID                    int64
	QuestionnaireKey          string
	QuestionnaireTitle        string
	QuestionnaireVersion      string
	TotalCount                int
	RequiredCount             int
	AnsweredCount             int
	RequiredAnsweredCount     int
	ProgressPercent           int
	HasCompletedQuestionnaire bool
	ParticipatingInMatch      bool
	LastAction                string
	LastSavedAt               time.Time
	ResultRevealAt            sql.NullTime
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type QuestionnairePortrait struct {
	UserID               int64
	QuestionnaireKey     string
	QuestionnaireVersion string
	Status               string
	AnswerSignature      string
	SoulSignature        string
	PersonalProfile      string
	IdealResonance       string
	TraitsJSON           json.RawMessage
	ErrorMessage         string
	GeneratedAt          sql.NullTime
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type UpsertQuestionnaireAnswerParams struct {
	UserID               int64
	QuestionnaireKey     string
	QuestionnaireVersion string
	QuestionID           string
	ValueJSON            json.RawMessage
	IsImportant          bool
	AnsweredAt           time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type UpsertQuestionnaireStateParams struct {
	UserID                    int64
	QuestionnaireKey          string
	QuestionnaireTitle        string
	QuestionnaireVersion      string
	TotalCount                int
	RequiredCount             int
	AnsweredCount             int
	RequiredAnsweredCount     int
	ProgressPercent           int
	HasCompletedQuestionnaire bool
	ParticipatingInMatch      bool
	LastAction                string
	LastSavedAt               time.Time
	ResultRevealAt            *time.Time
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type UpsertQuestionnairePortraitParams struct {
	UserID               int64
	QuestionnaireKey     string
	QuestionnaireVersion string
	Status               string
	AnswerSignature      string
	SoulSignature        string
	PersonalProfile      string
	IdealResonance       string
	TraitsJSON           json.RawMessage
	ErrorMessage         string
	GeneratedAt          *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// ListQuestionnaireAnswers 查询用户在指定问卷版本下的所有已保存答案。
func (s *Store) ListQuestionnaireAnswers(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string) ([]QuestionnaireAnswer, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
        SELECT id, user_id, questionnaire_key, questionnaire_version, question_id,
               value_json, is_important, answered_at, created_at, updated_at
        FROM user_questionnaire_answers
        WHERE user_id = $1 AND questionnaire_key = $2 AND questionnaire_version = $3
        ORDER BY answered_at ASC, id ASC
    `, userID, questionnaireKey, questionnaireVersion)
	if err != nil {
		return nil, fmt.Errorf("list questionnaire answers: %w", err)
	}
	defer rows.Close()

	answers := make([]QuestionnaireAnswer, 0)
	for rows.Next() {
		var answer QuestionnaireAnswer
		var raw []byte
		if err := rows.Scan(
			&answer.ID,
			&answer.UserID,
			&answer.QuestionnaireKey,
			&answer.QuestionnaireVersion,
			&answer.QuestionID,
			&raw,
			&answer.IsImportant,
			&answer.AnsweredAt,
			&answer.CreatedAt,
			&answer.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan questionnaire answer: %w", err)
		}
		answer.ValueJSON = append(json.RawMessage(nil), raw...)
		answers = append(answers, answer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate questionnaire answers: %w", err)
	}
	return answers, nil
}

// GetQuestionnaireState 查询用户在指定问卷版本下的保存状态。
func (s *Store) GetQuestionnaireState(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string) (QuestionnaireState, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
        SELECT user_id, questionnaire_key, questionnaire_title, questionnaire_version, total_count,
               required_count, answered_count, required_answered_count, progress_percent,
               has_completed_questionnaire, participating_in_match, last_action, last_saved_at, result_reveal_at,
               created_at, updated_at
        FROM user_questionnaire_states
        WHERE user_id = $1 AND questionnaire_key = $2 AND questionnaire_version = $3
        LIMIT 1
    `, userID, questionnaireKey, questionnaireVersion)

	var state QuestionnaireState
	if err := row.Scan(
		&state.UserID,
		&state.QuestionnaireKey,
		&state.QuestionnaireTitle,
		&state.QuestionnaireVersion,
		&state.TotalCount,
		&state.RequiredCount,
		&state.AnsweredCount,
		&state.RequiredAnsweredCount,
		&state.ProgressPercent,
		&state.HasCompletedQuestionnaire,
		&state.ParticipatingInMatch,
		&state.LastAction,
		&state.LastSavedAt,
		&state.ResultRevealAt,
		&state.CreatedAt,
		&state.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return QuestionnaireState{}, ErrNotFound
		}
		return QuestionnaireState{}, fmt.Errorf("get questionnaire state: %w", err)
	}
	return state, nil
}

// ListQuestionnaireStatesByVersion 查询指定问卷版本下所有用户的保存状态。
func (s *Store) ListQuestionnaireStatesByVersion(ctx context.Context, tx *sql.Tx, questionnaireKey, questionnaireVersion string) ([]QuestionnaireState, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
        SELECT user_id, questionnaire_key, questionnaire_title, questionnaire_version, total_count,
               required_count, answered_count, required_answered_count, progress_percent,
               has_completed_questionnaire, participating_in_match, last_action, last_saved_at, result_reveal_at,
               created_at, updated_at
        FROM user_questionnaire_states
        WHERE questionnaire_key = $1 AND questionnaire_version = $2
        ORDER BY user_id ASC
    `, questionnaireKey, questionnaireVersion)
	if err != nil {
		return nil, fmt.Errorf("list questionnaire states by version: %w", err)
	}
	defer rows.Close()

	states := make([]QuestionnaireState, 0)
	for rows.Next() {
		var state QuestionnaireState
		if err := rows.Scan(
			&state.UserID,
			&state.QuestionnaireKey,
			&state.QuestionnaireTitle,
			&state.QuestionnaireVersion,
			&state.TotalCount,
			&state.RequiredCount,
			&state.AnsweredCount,
			&state.RequiredAnsweredCount,
			&state.ProgressPercent,
			&state.HasCompletedQuestionnaire,
			&state.ParticipatingInMatch,
			&state.LastAction,
			&state.LastSavedAt,
			&state.ResultRevealAt,
			&state.CreatedAt,
			&state.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan questionnaire state: %w", err)
		}
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate questionnaire states: %w", err)
	}
	return states, nil
}

// UpsertQuestionnaireAnswer 保存或覆盖指定问卷版本下的单题答案。
func (s *Store) UpsertQuestionnaireAnswer(ctx context.Context, tx *sql.Tx, params UpsertQuestionnaireAnswerParams) (QuestionnaireAnswer, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
        INSERT INTO user_questionnaire_answers (
            user_id, questionnaire_key, questionnaire_version, question_id,
            value_json, is_important, answered_at, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8, $9)
        ON CONFLICT (user_id, questionnaire_key, questionnaire_version, question_id)
        DO UPDATE SET
            value_json = EXCLUDED.value_json,
            is_important = EXCLUDED.is_important,
            answered_at = EXCLUDED.answered_at,
            updated_at = EXCLUDED.updated_at
        RETURNING id, user_id, questionnaire_key, questionnaire_version, question_id,
                  value_json, is_important, answered_at, created_at, updated_at
    `,
		params.UserID,
		params.QuestionnaireKey,
		params.QuestionnaireVersion,
		params.QuestionID,
		string(params.ValueJSON),
		params.IsImportant,
		params.AnsweredAt,
		params.CreatedAt,
		params.UpdatedAt,
	)

	var answer QuestionnaireAnswer
	var raw []byte
	if err := row.Scan(
		&answer.ID,
		&answer.UserID,
		&answer.QuestionnaireKey,
		&answer.QuestionnaireVersion,
		&answer.QuestionID,
		&raw,
		&answer.IsImportant,
		&answer.AnsweredAt,
		&answer.CreatedAt,
		&answer.UpdatedAt,
	); err != nil {
		return QuestionnaireAnswer{}, fmt.Errorf("upsert questionnaire answer: %w", err)
	}
	answer.ValueJSON = append(json.RawMessage(nil), raw...)
	return answer, nil
}

// UpsertQuestionnaireState 保存用户在指定问卷版本下的整体进度状态。
func (s *Store) UpsertQuestionnaireState(ctx context.Context, tx *sql.Tx, params UpsertQuestionnaireStateParams) (QuestionnaireState, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
        INSERT INTO user_questionnaire_states (
            user_id, questionnaire_key, questionnaire_title, questionnaire_version,
            total_count, required_count, answered_count, required_answered_count,
            progress_percent, has_completed_questionnaire, participating_in_match, last_action,
            last_saved_at, result_reveal_at, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
        ON CONFLICT (user_id, questionnaire_key, questionnaire_version)
        DO UPDATE SET
            questionnaire_title = EXCLUDED.questionnaire_title,
            total_count = EXCLUDED.total_count,
            required_count = EXCLUDED.required_count,
            answered_count = EXCLUDED.answered_count,
            required_answered_count = EXCLUDED.required_answered_count,
            progress_percent = EXCLUDED.progress_percent,
            has_completed_questionnaire = EXCLUDED.has_completed_questionnaire,
            participating_in_match = EXCLUDED.participating_in_match,
            last_action = EXCLUDED.last_action,
            last_saved_at = EXCLUDED.last_saved_at,
            result_reveal_at = EXCLUDED.result_reveal_at,
            updated_at = EXCLUDED.updated_at
        RETURNING user_id, questionnaire_key, questionnaire_title, questionnaire_version, total_count,
                  required_count, answered_count, required_answered_count, progress_percent,
                  has_completed_questionnaire, participating_in_match, last_action, last_saved_at, result_reveal_at,
                  created_at, updated_at
    `,
		params.UserID,
		params.QuestionnaireKey,
		params.QuestionnaireTitle,
		params.QuestionnaireVersion,
		params.TotalCount,
		params.RequiredCount,
		params.AnsweredCount,
		params.RequiredAnsweredCount,
		params.ProgressPercent,
		params.HasCompletedQuestionnaire,
		params.ParticipatingInMatch,
		params.LastAction,
		params.LastSavedAt,
		nullableQuestionnaireRevealTime(params.ResultRevealAt),
		params.CreatedAt,
		params.UpdatedAt,
	)

	var state QuestionnaireState
	if err := row.Scan(
		&state.UserID,
		&state.QuestionnaireKey,
		&state.QuestionnaireTitle,
		&state.QuestionnaireVersion,
		&state.TotalCount,
		&state.RequiredCount,
		&state.AnsweredCount,
		&state.RequiredAnsweredCount,
		&state.ProgressPercent,
		&state.HasCompletedQuestionnaire,
		&state.ParticipatingInMatch,
		&state.LastAction,
		&state.LastSavedAt,
		&state.ResultRevealAt,
		&state.CreatedAt,
		&state.UpdatedAt,
	); err != nil {
		return QuestionnaireState{}, fmt.Errorf("upsert questionnaire state: %w", err)
	}
	return state, nil
}

// CountQuestionnaireRevealAssignments 统计指定问卷版本下已经分配公布时间的用户数量。
func (s *Store) CountQuestionnaireRevealAssignments(ctx context.Context, tx *sql.Tx, questionnaireKey, questionnaireVersion string) (int, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM user_questionnaire_states
        WHERE questionnaire_key = $1
          AND questionnaire_version = $2
          AND has_completed_questionnaire = TRUE
          AND result_reveal_at IS NOT NULL
    `, questionnaireKey, questionnaireVersion)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count questionnaire reveal assignments: %w", err)
	}
	return count, nil
}

// UpdateQuestionnaireStateRevealAt 更新单个用户在指定问卷版本下的公布时间。
func (s *Store) UpdateQuestionnaireStateRevealAt(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string, revealAt *time.Time, now time.Time) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
        UPDATE user_questionnaire_states
        SET result_reveal_at = $4,
            updated_at = $5
        WHERE user_id = $1 AND questionnaire_key = $2 AND questionnaire_version = $3
    `, userID, questionnaireKey, questionnaireVersion, nullableQuestionnaireRevealTime(revealAt), now); err != nil {
		return fmt.Errorf("update questionnaire state reveal at: %w", err)
	}
	return nil
}

func (s *Store) UpdateQuestionnaireMatchParticipation(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string, participating bool, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
        UPDATE user_questionnaire_states
        SET participating_in_match = $4,
            updated_at = $5
        WHERE user_id = $1 AND questionnaire_key = $2 AND questionnaire_version = $3
    `, userID, questionnaireKey, questionnaireVersion, participating, now)
	if err != nil {
		return fmt.Errorf("update questionnaire match participation: %w", err)
	}
	return requireRowsAffected(result)
}

// ResetQuestionnaireMatchParticipationByVersion 将指定问卷版本下所有用户的参与状态重置为未参与。
func (s *Store) ResetQuestionnaireMatchParticipationByVersion(ctx context.Context, tx *sql.Tx, questionnaireKey, questionnaireVersion string, now time.Time) (int64, error) {
	result, err := executor(s.db, tx).ExecContext(ctx, `
        UPDATE user_questionnaire_states
        SET participating_in_match = FALSE,
            updated_at = $3
        WHERE questionnaire_key = $1
          AND questionnaire_version = $2
          AND participating_in_match = TRUE
    `, questionnaireKey, questionnaireVersion, now)
	if err != nil {
		return 0, fmt.Errorf("reset questionnaire match participation by version: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read affected rows for resetting questionnaire match participation by version: %w", err)
	}
	return rowsAffected, nil
}

func (s *Store) GetQuestionnairePortrait(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string) (QuestionnairePortrait, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
        SELECT user_id, questionnaire_key, questionnaire_version, status,
               answer_signature,
               soul_signature, personal_profile, ideal_resonance, traits_json,
               error_message, generated_at, created_at, updated_at
        FROM user_questionnaire_portraits
        WHERE user_id = $1 AND questionnaire_key = $2 AND questionnaire_version = $3
        LIMIT 1
    `, userID, questionnaireKey, questionnaireVersion)

	var portrait QuestionnairePortrait
	var raw []byte
	if err := row.Scan(
		&portrait.UserID,
		&portrait.QuestionnaireKey,
		&portrait.QuestionnaireVersion,
		&portrait.Status,
		&portrait.AnswerSignature,
		&portrait.SoulSignature,
		&portrait.PersonalProfile,
		&portrait.IdealResonance,
		&raw,
		&portrait.ErrorMessage,
		&portrait.GeneratedAt,
		&portrait.CreatedAt,
		&portrait.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return QuestionnairePortrait{}, ErrNotFound
		}
		return QuestionnairePortrait{}, fmt.Errorf("get questionnaire portrait: %w", err)
	}
	if len(raw) > 0 {
		portrait.TraitsJSON = append(json.RawMessage(nil), raw...)
	}
	return portrait, nil
}

func (s *Store) UpsertQuestionnairePortrait(ctx context.Context, tx *sql.Tx, params UpsertQuestionnairePortraitParams) (QuestionnairePortrait, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
        INSERT INTO user_questionnaire_portraits (
            user_id, questionnaire_key, questionnaire_version, status,
            answer_signature,
            soul_signature, personal_profile, ideal_resonance, traits_json,
            error_message, generated_at, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11, $12, $13)
        ON CONFLICT (user_id, questionnaire_key, questionnaire_version)
        DO UPDATE SET
            status = EXCLUDED.status,
            answer_signature = EXCLUDED.answer_signature,
            soul_signature = EXCLUDED.soul_signature,
            personal_profile = EXCLUDED.personal_profile,
            ideal_resonance = EXCLUDED.ideal_resonance,
            traits_json = EXCLUDED.traits_json,
            error_message = EXCLUDED.error_message,
            generated_at = EXCLUDED.generated_at,
            updated_at = EXCLUDED.updated_at
        RETURNING user_id, questionnaire_key, questionnaire_version, status,
                  answer_signature,
                  soul_signature, personal_profile, ideal_resonance, traits_json,
                  error_message, generated_at, created_at, updated_at
    `,
		params.UserID,
		params.QuestionnaireKey,
		params.QuestionnaireVersion,
		params.Status,
		params.AnswerSignature,
		params.SoulSignature,
		params.PersonalProfile,
		params.IdealResonance,
		nullableJSON(params.TraitsJSON),
		params.ErrorMessage,
		nullableQuestionnaireRevealTime(params.GeneratedAt),
		params.CreatedAt,
		params.UpdatedAt,
	)

	var portrait QuestionnairePortrait
	var raw []byte
	if err := row.Scan(
		&portrait.UserID,
		&portrait.QuestionnaireKey,
		&portrait.QuestionnaireVersion,
		&portrait.Status,
		&portrait.AnswerSignature,
		&portrait.SoulSignature,
		&portrait.PersonalProfile,
		&portrait.IdealResonance,
		&raw,
		&portrait.ErrorMessage,
		&portrait.GeneratedAt,
		&portrait.CreatedAt,
		&portrait.UpdatedAt,
	); err != nil {
		return QuestionnairePortrait{}, fmt.Errorf("upsert questionnaire portrait: %w", err)
	}
	if len(raw) > 0 {
		portrait.TraitsJSON = append(json.RawMessage(nil), raw...)
	}
	return portrait, nil
}

// ClearQuestionnaireRevealAssignments 清空指定问卷版本下所有用户的公布时间。
func (s *Store) ClearQuestionnaireRevealAssignments(ctx context.Context, tx *sql.Tx, questionnaireKey, questionnaireVersion string) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
        UPDATE user_questionnaire_states
        SET result_reveal_at = NULL
        WHERE questionnaire_key = $1 AND questionnaire_version = $2 AND result_reveal_at IS NOT NULL
    `, questionnaireKey, questionnaireVersion); err != nil {
		return fmt.Errorf("clear questionnaire reveal assignments: %w", err)
	}
	return nil
}

// SetUserQuestionnaireCompleted 更新用户的问卷完成标记。
func (s *Store) SetUserQuestionnaireCompleted(ctx context.Context, tx *sql.Tx, userID int64, completed bool, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
        UPDATE users
        SET has_completed_questionnaire = $2, updated_at = $3
        WHERE id = $1 AND deleted_at IS NULL
    `, userID, completed, now)
	if err != nil {
		return fmt.Errorf("set user questionnaire completed: %w", err)
	}
	return requireRowsAffected(result)
}

// DeleteQuestionnaireDataByUser 删除用户的问卷状态和答案。
func (s *Store) DeleteQuestionnaireDataByUser(ctx context.Context, tx *sql.Tx, userID int64) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `DELETE FROM user_questionnaire_answers WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete questionnaire answers: %w", err)
	}
	if _, err := executor(s.db, tx).ExecContext(ctx, `DELETE FROM user_questionnaire_states WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete questionnaire states: %w", err)
	}
	if _, err := executor(s.db, tx).ExecContext(ctx, `DELETE FROM user_questionnaire_portraits WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete questionnaire portraits: %w", err)
	}
	return nil
}

// DeleteQuestionnaireVersionDataByUser 删除用户在指定问卷版本下的答案、状态和画像。
func (s *Store) DeleteQuestionnaireVersionDataByUser(ctx context.Context, tx *sql.Tx, userID int64, questionnaireKey, questionnaireVersion string) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
        DELETE FROM user_questionnaire_answers
        WHERE user_id = $1 AND questionnaire_key = $2 AND questionnaire_version = $3
    `, userID, questionnaireKey, questionnaireVersion); err != nil {
		return fmt.Errorf("delete questionnaire answers by version: %w", err)
	}
	if _, err := executor(s.db, tx).ExecContext(ctx, `
        DELETE FROM user_questionnaire_states
        WHERE user_id = $1 AND questionnaire_key = $2 AND questionnaire_version = $3
    `, userID, questionnaireKey, questionnaireVersion); err != nil {
		return fmt.Errorf("delete questionnaire states by version: %w", err)
	}
	if _, err := executor(s.db, tx).ExecContext(ctx, `
        DELETE FROM user_questionnaire_portraits
        WHERE user_id = $1 AND questionnaire_key = $2 AND questionnaire_version = $3
    `, userID, questionnaireKey, questionnaireVersion); err != nil {
		return fmt.Errorf("delete questionnaire portraits by version: %w", err)
	}
	return nil
}

func nullableQuestionnaireRevealTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableJSON(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}
