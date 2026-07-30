package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("store: not found")

type CodePurpose string

const (
	CodePurposeRegister       CodePurpose = "register"
	CodePurposePasswordForgot CodePurpose = "password_forgot"
	CodePurposePasswordChange CodePurpose = "password_change"
	CodePurposeAccountDelete  CodePurpose = "account_delete"
)

type Store struct {
	db *sql.DB
}

type User struct {
	ID                        int64
	Email                     string
	SchoolKey                 string
	PasswordHash              string
	Nickname                  string
	Department                string
	Grade                     string
	ContactType               string
	ContactValue              string
	HasCompletedQuestionnaire bool
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	DeletedAt                 sql.NullTime
}

type Session struct {
	ID         string
	UserID     int64
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastSeenAt time.Time
	RevokedAt  sql.NullTime
}

type VerificationCode struct {
	ID         int64
	Email      string
	Purpose    CodePurpose
	CodeHash   string
	ExpiresAt  time.Time
	SentAt     time.Time
	VerifiedAt sql.NullTime
	ConsumedAt sql.NullTime
}

type CreateUserParams struct {
	Email        string
	SchoolKey    string
	PasswordHash string
	Nickname     string
	Department   string
	Grade        string
	Now          time.Time
}

type UpdateProfileParams struct {
	UserID       int64
	Nickname     string
	Department   string
	Grade        string
	ContactType  string
	ContactValue string
	Now          time.Time
}

type CreateVerificationCodeParams struct {
	Email     string
	Purpose   CodePurpose
	CodeHash  string
	ExpiresAt time.Time
	SentAt    time.Time
}

type CreateSessionParams struct {
	ID         string
	UserID     int64
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastSeenAt time.Time
}

type queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// New 鍒涘缓鏁版嵁搴撳瓨鍌ㄥ眰瀹炰緥銆?
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// WithTx 鍦ㄥ悓涓€涓簨鍔′腑鎵ц涓€缁勬暟鎹簱鎿嶄綔銆?
func (s *Store) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// GetUserByEmail 鎸夐偖绠辨煡璇㈡湭鍒犻櫎鐢ㄦ埛銆?
func (s *Store) GetUserByEmail(ctx context.Context, tx *sql.Tx, email string) (User, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,
		       has_completed_questionnaire, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
		LIMIT 1
	`, email)

	user, err := scanUser(row)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// GetUserByID 鎸変富閿煡璇㈡湭鍒犻櫎鐢ㄦ埛銆?
func (s *Store) GetUserByNickname(ctx context.Context, tx *sql.Tx, nickname string) (User, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,
		       has_completed_questionnaire, created_at, updated_at, deleted_at
		FROM users
		WHERE LOWER(nickname) = LOWER($1) AND deleted_at IS NULL
		LIMIT 1
	`, nickname)

	user, err := scanUser(row)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, tx *sql.Tx, userID int64) (User, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,
		       has_completed_questionnaire, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
	`, userID)

	user, err := scanUser(row)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// CreateUser 鍒涘缓鏂扮敤鎴疯褰曘€?
func (s *Store) CreateUser(ctx context.Context, tx *sql.Tx, params CreateUserParams) (User, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO users (
			email, school_key, password_hash, nickname, department, grade, has_completed_questionnaire, created_at, updated_at
		) VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, FALSE, $7, $7)
		RETURNING id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,
		          has_completed_questionnaire, created_at, updated_at, deleted_at
	`, params.Email, params.SchoolKey, params.PasswordHash, params.Nickname, params.Department, params.Grade, params.Now)

	user, err := scanUser(row)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// UpdateUserProfile 鏇存柊鐢ㄦ埛璧勬枡瀛楁銆?
func (s *Store) UpdateUserProfile(ctx context.Context, tx *sql.Tx, params UpdateProfileParams) (User, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		UPDATE users
		SET nickname = $2,
		    department = $3,
		    grade = $4,
		    contact_type = NULLIF($5, ''),
		    contact_value = NULLIF($6, ''),
		    updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,
		          has_completed_questionnaire, created_at, updated_at, deleted_at
	`, params.UserID, params.Nickname, params.Department, params.Grade, params.ContactType, params.ContactValue, params.Now)

	user, err := scanUser(row)
	if err != nil {
		return User{}, fmt.Errorf("update user profile: %w", err)
	}
	return user, nil
}

// UpdateUserPassword 鏇存柊鐢ㄦ埛瀵嗙爜鎽樿銆?
func (s *Store) UpdateUserPassword(ctx context.Context, tx *sql.Tx, userID int64, passwordHash string, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE users
		SET password_hash = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, passwordHash, now)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return requireRowsAffected(result)
}

// SoftDeleteUser 瀵圭敤鎴锋墽琛岄€昏緫鍒犻櫎銆?
func (s *Store) SoftDeleteUser(ctx context.Context, tx *sql.Tx, userID int64, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE users
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, now)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	return requireRowsAffected(result)
}

// DeleteUser 鐗╃悊鍒犻櫎鐢ㄦ埛璁板綍銆?
func (s *Store) DeleteUser(ctx context.Context, tx *sql.Tx, userID int64) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM users
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return requireRowsAffected(result)
}

// CreateVerificationCode 鎸佷箙鍖栨柊鐨勯獙璇佺爜璁板綍銆?
// CountActiveUsers returns the number of users whose account is still active.
func (s *Store) CountActiveUsers(ctx context.Context, tx *sql.Tx) (int, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM users
		WHERE deleted_at IS NULL
	`)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count active users: %w", err)
	}
	return count, nil
}

func (s *Store) CountActiveUsersByEmailSuffixes(ctx context.Context, tx *sql.Tx, suffixes []string) (int, error) {
	conditions := make([]string, 0, len(suffixes))
	args := make([]any, 0, len(suffixes))
	for _, suffix := range suffixes {
		normalized := strings.ToLower(strings.TrimSpace(suffix))
		if normalized == "" {
			continue
		}
		args = append(args, "%"+normalized)
		conditions = append(conditions, fmt.Sprintf("LOWER(email) LIKE $%d", len(args)))
	}
	if len(conditions) == 0 {
		return 0, nil
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM users
		WHERE deleted_at IS NULL
		  AND (%s)
	`, strings.Join(conditions, " OR "))

	row := executor(s.db, tx).QueryRowContext(ctx, query, args...)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count active users by email suffix: %w", err)
	}
	return count, nil
}
func (s *Store) CreateVerificationCode(ctx context.Context, tx *sql.Tx, params CreateVerificationCodeParams) (VerificationCode, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO verification_codes (email, purpose, code_hash, expires_at, sent_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, purpose, code_hash, expires_at, sent_at, verified_at, consumed_at
	`, params.Email, params.Purpose, params.CodeHash, params.ExpiresAt, params.SentAt)

	code, err := scanVerificationCode(row)
	if err != nil {
		return VerificationCode{}, fmt.Errorf("create verification code: %w", err)
	}
	return code, nil
}

// DeleteVerificationCode 鍒犻櫎楠岃瘉鐮佽褰曪紝甯哥敤浜庡彂閫佸け璐ュ洖婊氥€?
func (s *Store) DeleteVerificationCode(ctx context.Context, tx *sql.Tx, codeID int64) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `DELETE FROM verification_codes WHERE id = $1`, codeID)
	if err != nil {
		return fmt.Errorf("delete verification code: %w", err)
	}
	return nil
}

// DeleteVerificationCodesByEmail 鍒犻櫎鎸囧畾閭鐨勫叏閮ㄩ獙璇佺爜璁板綍銆?
func (s *Store) DeleteVerificationCodesByEmail(ctx context.Context, tx *sql.Tx, email string) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM verification_codes
		WHERE email = $1
	`, email); err != nil {
		return fmt.Errorf("delete verification codes by email: %w", err)
	}
	return nil
}

// GetLatestVerificationCode 鏌ヨ鎸囧畾閭鍜岀敤閫旂殑鏈€鏂伴獙璇佺爜銆?
func (s *Store) GetLatestVerificationCode(ctx context.Context, tx *sql.Tx, email string, purpose CodePurpose) (VerificationCode, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		SELECT id, email, purpose, code_hash, expires_at, sent_at, verified_at, consumed_at
		FROM verification_codes
		WHERE email = $1 AND purpose = $2
		ORDER BY sent_at DESC
		LIMIT 1
	`, email, purpose)

	code, err := scanVerificationCode(row)
	if err != nil {
		return VerificationCode{}, err
	}
	return code, nil
}

// MarkVerificationCodeVerified 鏍囪楠岃瘉鐮佸凡缁忓畬鎴愰鏍￠獙銆?
func (s *Store) MarkVerificationCodeVerified(ctx context.Context, tx *sql.Tx, codeID int64, now time.Time) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE verification_codes
		SET verified_at = COALESCE(verified_at, $2)
		WHERE id = $1
	`, codeID, now)
	if err != nil {
		return fmt.Errorf("mark verification code verified: %w", err)
	}
	return nil
}

// ConsumeVerificationCode 鏍囪楠岃瘉鐮佸凡琚寮忔秷璐广€?
func (s *Store) ConsumeVerificationCode(ctx context.Context, tx *sql.Tx, codeID int64, now time.Time) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE verification_codes
		SET consumed_at = $2
		WHERE id = $1
	`, codeID, now)
	if err != nil {
		return fmt.Errorf("consume verification code: %w", err)
	}
	return nil
}

// CreateSession 鍒涘缓鏂扮殑鐧诲綍浼氳瘽璁板綍銆?
func (s *Store) CreateSession(ctx context.Context, tx *sql.Tx, params CreateSessionParams) (Session, error) {
	row := executor(s.db, tx).QueryRowContext(ctx, `
		INSERT INTO auth_sessions (id, user_id, expires_at, created_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, expires_at, created_at, last_seen_at, revoked_at
	`, params.ID, params.UserID, params.ExpiresAt, params.CreatedAt, params.LastSeenAt)

	session, err := scanSession(row)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

// GetActiveSession 鏌ヨ浠嶆湁鏁堢殑鐧诲綍浼氳瘽鍙婂叾鐢ㄦ埛淇℃伅銆?
func (s *Store) GetActiveSession(ctx context.Context, sessionID string) (Session, User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT s.id, s.user_id, s.expires_at, s.created_at, s.last_seen_at, s.revoked_at,
		       u.id, u.email, u.school_key, u.password_hash, u.nickname, u.department, u.grade, u.contact_type, u.contact_value,
		       u.has_completed_questionnaire, u.created_at, u.updated_at, u.deleted_at
		FROM auth_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = $1
		  AND s.revoked_at IS NULL
		  AND s.expires_at > NOW()
		  AND u.deleted_at IS NULL
		LIMIT 1
	`, sessionID)

	var session Session
	var user User
	var schoolKey sql.NullString
	var contactType sql.NullString
	var contactValue sql.NullString
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastSeenAt,
		&session.RevokedAt,
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
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, User{}, ErrNotFound
		}
		return Session{}, User{}, fmt.Errorf("get active session: %w", err)
	}

	user.SchoolKey = nullableString(schoolKey)
	user.ContactType = nullableString(contactType)
	user.ContactValue = nullableString(contactValue)
	return session, user, nil
}

// RevokeSession 鎾ら攢鍗曚釜鐧诲綍浼氳瘽銆?
func (s *Store) RevokeSession(ctx context.Context, tx *sql.Tx, sessionID string, now time.Time) error {
	result, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, $2)
		WHERE id = $1
	`, sessionID, now)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return requireRowsAffected(result)
}

// RevokeAllSessionsByUser 鎾ら攢鎸囧畾鐢ㄦ埛鐨勫叏閮ㄤ細璇濄€?
func (s *Store) RevokeAllSessionsByUser(ctx context.Context, tx *sql.Tx, userID int64, now time.Time) error {
	_, err := executor(s.db, tx).ExecContext(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, $2)
		WHERE user_id = $1
	`, userID, now)
	if err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

// DeleteSessionsByUser 鍒犻櫎鎸囧畾鐢ㄦ埛鐨勫叏閮ㄤ細璇濊褰曘€?
func (s *Store) DeleteSessionsByUser(ctx context.Context, tx *sql.Tx, userID int64) error {
	if _, err := executor(s.db, tx).ExecContext(ctx, `
		DELETE FROM auth_sessions
		WHERE user_id = $1
	`, userID); err != nil {
		return fmt.Errorf("delete user sessions: %w", err)
	}
	return nil
}

func executor(db *sql.DB, tx *sql.Tx) queryer {
	if tx != nil {
		return tx
	}
	return db
}

func scanUser(row *sql.Row) (User, error) {
	var user User
	var schoolKey sql.NullString
	var contactType sql.NullString
	var contactValue sql.NullString

	err := row.Scan(
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
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}

	user.SchoolKey = nullableString(schoolKey)
	user.ContactType = nullableString(contactType)
	user.ContactValue = nullableString(contactValue)
	return user, nil
}

func scanVerificationCode(row *sql.Row) (VerificationCode, error) {
	var code VerificationCode
	err := row.Scan(
		&code.ID,
		&code.Email,
		&code.Purpose,
		&code.CodeHash,
		&code.ExpiresAt,
		&code.SentAt,
		&code.VerifiedAt,
		&code.ConsumedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return VerificationCode{}, ErrNotFound
		}
		return VerificationCode{}, fmt.Errorf("scan verification code: %w", err)
	}
	return code, nil
}

func scanSession(row *sql.Row) (Session, error) {
	var session Session
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastSeenAt,
		&session.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("scan session: %w", err)
	}
	return session, nil
}

func requireRowsAffected(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func nullableString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}
