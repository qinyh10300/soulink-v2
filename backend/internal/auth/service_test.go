package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"

	"soulink-backend/internal/config"
	"soulink-backend/internal/mail"
	"soulink-backend/internal/store"
)

type noopSender struct{}

func (noopSender) Send(context.Context, mail.Message) error {
	return nil
}

type stubRegisteredUserCountRefresher struct {
	calls int
	err   error
}

func (r *stubRegisteredUserCountRefresher) RefreshRegisteredUserCount(context.Context) error {
	r.calls++
	return r.err
}

func newAuthServiceForTest(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}

	service := NewService(
		store.New(db),
		noopSender{},
		NewTokenManager("secret", time.Hour),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		config.AuthConfig{
			JWTSecret:                "secret",
			JWTTTL:                   time.Hour,
			VerificationCodeTTL:      5 * time.Minute,
			VerificationCodeCooldown: time.Minute,
			BCryptCost:               bcrypt.MinCost,
		},
	)

	return service, mock, func() { _ = db.Close() }
}

func TestServiceLoginSuccess(t *testing.T) {
	service, mock, cleanup := newAuthServiceForTest(t)
	defer cleanup()

	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error: %v", err)
	}

	now := time.Now()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE email = \$1.*LIMIT 1`).
		WithArgs("test@mails.tsinghua.edu.cn").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "school_key", "password_hash", "nickname", "department", "grade", "contact_type", "contact_value", "has_completed_questionnaire", "created_at", "updated_at", "deleted_at"}).
			AddRow(int64(1), "test@mails.tsinghua.edu.cn", "thu", string(hash), "tester", "dept", "grade", nil, nil, false, now, now, nil))
	mock.ExpectQuery(`(?s)INSERT INTO auth_sessions .*RETURNING id, user_id, expires_at, created_at, last_seen_at, revoked_at`).
		WithArgs(sqlmock.AnyArg(), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "expires_at", "created_at", "last_seen_at", "revoked_at"}).
			AddRow("session-1", int64(1), now.Add(time.Hour), now, now, nil))

	token, user, err := service.Login(context.Background(), "test@mails.tsinghua.edu.cn", "12345678")
	if err != nil {
		t.Fatalf("Login() unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("Login() returned empty token")
	}
	if user.Email != "test@mails.tsinghua.edu.cn" {
		t.Fatalf("Login() user email = %q, want %q", user.Email, "test@mails.tsinghua.edu.cn")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceLoginRejectsWrongPassword(t *testing.T) {
	service, mock, cleanup := newAuthServiceForTest(t)
	defer cleanup()

	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error: %v", err)
	}

	now := time.Now()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE email = \$1.*LIMIT 1`).
		WithArgs("test@mails.tsinghua.edu.cn").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "school_key", "password_hash", "nickname", "department", "grade", "contact_type", "contact_value", "has_completed_questionnaire", "created_at", "updated_at", "deleted_at"}).
			AddRow(int64(1), "test@mails.tsinghua.edu.cn", "thu", string(hash), "tester", "dept", "grade", nil, nil, false, now, now, nil))

	if _, _, err := service.Login(context.Background(), "test@mails.tsinghua.edu.cn", "wrong-password"); err == nil {
		t.Fatal("Login() expected error for invalid password")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceSendRegisterCodeRejectsDuplicateEmail(t *testing.T) {
	service, mock, cleanup := newAuthServiceForTest(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery(`(?s)SELECT id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*FROM users.*WHERE email = \$1.*LIMIT 1`).
		WithArgs("test@mails.tsinghua.edu.cn").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "school_key", "password_hash", "nickname", "department", "grade", "contact_type", "contact_value", "has_completed_questionnaire", "created_at", "updated_at", "deleted_at"}).
			AddRow(int64(1), "test@mails.tsinghua.edu.cn", "thu", "hash", "tester", "dept", "grade", nil, nil, false, now, now, nil))

	_, err := service.SendRegisterCode(context.Background(), "test@mails.tsinghua.edu.cn")
	if err == nil {
		t.Fatal("SendRegisterCode() expected duplicate email error")
	}
	if !regexp.MustCompile(`email already registered`).MatchString(err.Error()) {
		t.Fatalf("SendRegisterCode() error = %v, want duplicate email error", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestServiceDeleteAccountDeletesUserData(t *testing.T) {
	service, mock, cleanup := newAuthServiceForTest(t)
	defer cleanup()

	now := time.Now()
	service.nowFunc = func() time.Time { return now }
	refresher := &stubRegisteredUserCountRefresher{}
	service.WithRegisteredUserCountRefresher(refresher)

	email := "test@mails.tsinghua.edu.cn"
	code := "123456"
	codeHash := HashVerificationCode(string(store.CodePurposeAccountDelete), email, code, "secret")

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT id, email, purpose, code_hash, expires_at, sent_at, verified_at, consumed_at.*FROM verification_codes.*WHERE email = \$1 AND purpose = \$2.*LIMIT 1`).
		WithArgs(email, store.CodePurposeAccountDelete).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "purpose", "code_hash", "expires_at", "sent_at", "verified_at", "consumed_at"}).
			AddRow(int64(11), email, string(store.CodePurposeAccountDelete), codeHash, now.Add(5*time.Minute), now.Add(-time.Minute), nil, nil))
	mock.ExpectExec(`(?s)UPDATE verification_codes.*SET consumed_at = \$2.*WHERE id = \$1`).
		WithArgs(int64(11), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM user_questionnaire_answers WHERE user_id = \$1`).
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`DELETE FROM user_questionnaire_states WHERE user_id = \$1`).
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM user_questionnaire_portraits WHERE user_id = \$1`).
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM user_contact_methods WHERE user_id = \$1`).
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`(?s)DELETE FROM user_match_results.*WHERE user_id = \$1 OR partner_user_id = \$1.*RETURNING cycle_id`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"cycle_id"}).AddRow(int64(9)))
	mock.ExpectExec(`(?s)UPDATE match_cycles.*SET participant_count = COALESCE.*WHERE id = \$1`).
		WithArgs(int64(9), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM auth_sessions WHERE user_id = \$1`).
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`DELETE FROM verification_codes WHERE email = \$1`).
		WithArgs(email).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := service.DeleteAccount(context.Background(), 7, email, code); err != nil {
		t.Fatalf("DeleteAccount() unexpected error: %v", err)
	}

	if refresher.calls != 1 {
		t.Fatalf("registered user count refresher calls = %d, want 1", refresher.calls)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
func TestServiceRegisterRefreshesRegisteredUserCount(t *testing.T) {
	service, mock, cleanup := newAuthServiceForTest(t)
	defer cleanup()

	now := time.Now()
	service.nowFunc = func() time.Time { return now }
	service.codeCache = newMemoryVerificationCodeCache()
	refresher := &stubRegisteredUserCountRefresher{}
	service.WithRegisteredUserCountRefresher(refresher)
	nickname := "\u5c0f\u738b"
	department := "\u8ba1\u7b97\u673a\u5b66\u9662"
	grade := "\u672c\u79d1\u4e09\u5e74\u7ea7"

	email := "register@pku.edu.cn"
	code := "123456"
	if err := service.codeCache.Store(context.Background(), email, store.CodePurposeRegister, CachedVerificationCode{
		CodeHash:  HashVerificationCode(string(store.CodePurposeRegister), email, code, "secret"),
		ExpiresAt: now.Add(5 * time.Minute),
		SentAt:    now.Add(-time.Minute),
	}, 5*time.Minute); err != nil {
		t.Fatalf("seed verification cache: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)INSERT INTO users .*RETURNING id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*deleted_at`).
		WithArgs(email, "pku", sqlmock.AnyArg(), nickname, department, grade, now).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "school_key", "password_hash", "nickname", "department", "grade", "contact_type", "contact_value", "has_completed_questionnaire", "created_at", "updated_at", "deleted_at"}).
			AddRow(int64(2), email, "pku", "hash", nickname, department, grade, nil, nil, false, now, now, nil))
	mock.ExpectQuery(`(?s)INSERT INTO auth_sessions .*RETURNING id, user_id, expires_at, created_at, last_seen_at, revoked_at`).
		WithArgs(sqlmock.AnyArg(), int64(2), sqlmock.AnyArg(), now, now).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "expires_at", "created_at", "last_seen_at", "revoked_at"}).
			AddRow("session-2", int64(2), now.Add(time.Hour), now, now, nil))
	mock.ExpectCommit()

	token, user, err := service.Register(context.Background(), RegisterInput{
		Email:      email,
		Code:       code,
		Password:   "12345678",
		Nickname:   nickname,
		Department: department,
		Grade:      grade,
	})
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("Register() returned empty token")
	}
	if user.Email != email {
		t.Fatalf("Register() user email = %q, want %q", user.Email, email)
	}
	if refresher.calls != 1 {
		t.Fatalf("registered user count refresher calls = %d, want 1", refresher.calls)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
func TestServiceRegisterRejectsDuplicateNickname(t *testing.T) {
	service, mock, cleanup := newAuthServiceForTest(t)
	defer cleanup()

	now := time.Now()
	service.nowFunc = func() time.Time { return now }
	service.codeCache = newMemoryVerificationCodeCache()
	department := "\u8ba1\u7b97\u673a\u5b66\u9662"
	grade := "\u672c\u79d1\u4e09\u5e74\u7ea7"

	email := "nickname-conflict@pku.edu.cn"
	code := "123456"
	if err := service.codeCache.Store(context.Background(), email, store.CodePurposeRegister, CachedVerificationCode{
		CodeHash:  HashVerificationCode(string(store.CodePurposeRegister), email, code, "secret"),
		ExpiresAt: now.Add(5 * time.Minute),
		SentAt:    now.Add(-time.Minute),
	}, 5*time.Minute); err != nil {
		t.Fatalf("seed verification cache: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)INSERT INTO users .*RETURNING id, email, school_key, password_hash, nickname, department, grade, contact_type, contact_value,.*deleted_at`).
		WithArgs(email, "pku", sqlmock.AnyArg(), "duplicate-name", department, grade, now).
		WillReturnError(errors.New(`duplicate key value violates unique constraint "ux_users_nickname_active"`))
	mock.ExpectRollback()

	_, _, err := service.Register(context.Background(), RegisterInput{
		Email:      email,
		Code:       code,
		Password:   "12345678",
		Nickname:   "duplicate-name",
		Department: department,
		Grade:      grade,
	})
	if err == nil {
		t.Fatal("Register() expected duplicate nickname error")
	}
	if !regexp.MustCompile(`nickname already exists`).MatchString(err.Error()) {
		t.Fatalf("Register() error = %v, want duplicate nickname error", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
