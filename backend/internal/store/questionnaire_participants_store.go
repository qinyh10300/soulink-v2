package store

import (
	"context"
	"database/sql"
	"fmt"
)

type QuestionnaireUserFilter struct {
	RequireCompletedQuestionnaire bool
	RequireParticipatingInMatch   bool
}

// ListCompletedQuestionnaireUsersByVersion 查询完成指定问卷版本的全部有效用户。
func (s *Store) ListCompletedQuestionnaireUsersByVersion(ctx context.Context, tx *sql.Tx, questionnaireVersion string) ([]User, error) {
	return s.ListQuestionnaireUsers(ctx, tx, "", questionnaireVersion, QuestionnaireUserFilter{
		RequireCompletedQuestionnaire: true,
		RequireParticipatingInMatch:   true,
	})
}

// ListCompletedQuestionnaireUsers 查询完成指定问卷标识和版本的全部有效用户；当问卷标识为空时，仅按版本筛选。
func (s *Store) ListCompletedQuestionnaireUsers(ctx context.Context, tx *sql.Tx, questionnaireKey, questionnaireVersion string) ([]User, error) {
	return s.ListQuestionnaireUsers(ctx, tx, questionnaireKey, questionnaireVersion, QuestionnaireUserFilter{
		RequireCompletedQuestionnaire: true,
		RequireParticipatingInMatch:   true,
	})
}

// ListQuestionnaireUsers 查询指定问卷下的有效用户，并允许按是否完成问卷、是否选择参与进行筛选。
func (s *Store) ListQuestionnaireUsers(ctx context.Context, tx *sql.Tx, questionnaireKey, questionnaireVersion string, filter QuestionnaireUserFilter) ([]User, error) {
	rows, err := executor(s.db, tx).(interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}).QueryContext(ctx, `
		SELECT u.id, u.email, u.school_key, u.password_hash, u.nickname, u.department, u.grade, u.contact_type, u.contact_value,
		       u.has_completed_questionnaire, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		JOIN user_questionnaire_states s ON s.user_id = u.id
		WHERE s.questionnaire_version = $1
		  AND ($2 = '' OR s.questionnaire_key = $2)
		  AND ($3 = FALSE OR s.has_completed_questionnaire = TRUE)
		  AND ($4 = FALSE OR s.participating_in_match = TRUE)
		  AND u.deleted_at IS NULL
		ORDER BY u.id ASC
	`, questionnaireVersion, questionnaireKey, filter.RequireCompletedQuestionnaire, filter.RequireParticipatingInMatch)
	if err != nil {
		return nil, fmt.Errorf("list questionnaire users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		user, err := scanUserFromRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate questionnaire users: %w", err)
	}
	return users, nil
}

type userRowScanner interface {
	Scan(dest ...any) error
}

func scanUserFromRows(scanner userRowScanner) (User, error) {
	var user User
	var schoolKey sql.NullString
	var contactType sql.NullString
	var contactValue sql.NullString

	if err := scanner.Scan(
		&user.ID,
		&user.Email,
		&schoolKey,
		&user.PasswordHash,
		&user.Nickname,
		&user.Department,
		&user.Grade,
		&contactType,
		&contactValue,
		&user.HasCompletedQuestionnaire,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	); err != nil {
		return User{}, fmt.Errorf("scan user rows: %w", err)
	}

	user.SchoolKey = nullableString(schoolKey)
	user.ContactType = nullableString(contactType)
	user.ContactValue = nullableString(contactValue)
	return user, nil
}
