package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"soulink-backend/internal/config"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/mail"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/school"

	"soulink-backend/internal/store"
)

type Service struct {
	store                        *store.Store
	mailer                       mail.Sender
	tokens                       *TokenManager
	logger                       *slog.Logger
	codeTTL                      time.Duration
	codeCooldown                 time.Duration
	bcryptCost                   int
	codeSecret                   string
	codeCache                    VerificationCodeCache
	registeredUserCountRefresher RegisteredUserCountRefresher
	nowFunc                      func() time.Time
}

type RegisterInput struct {
	Email      string
	Code       string
	Password   string
	Nickname   string
	Department string
	Grade      string
}

type RegisteredUserCountRefresher interface {
	RefreshRegisteredUserCount(ctx context.Context) error
}

func NewService(dataStore *store.Store, mailer mail.Sender, tokens *TokenManager, logger *slog.Logger, cfg config.AuthConfig) *Service {
	return NewServiceWithVerificationCache(dataStore, mailer, tokens, logger, cfg, nil)
}

func NewServiceWithVerificationCache(dataStore *store.Store, mailer mail.Sender, tokens *TokenManager, logger *slog.Logger, cfg config.AuthConfig, codeCache VerificationCodeCache) *Service {
	return &Service{store: dataStore, mailer: mailer, tokens: tokens, logger: logger, codeTTL: cfg.VerificationCodeTTL, codeCooldown: cfg.VerificationCodeCooldown, bcryptCost: cfg.BCryptCost, codeSecret: cfg.JWTSecret, codeCache: codeCache, nowFunc: time.Now}
}

func (s *Service) Login(ctx context.Context, email, password string) (string, store.User, error) {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return "", store.User{}, err
	}
	if err := ValidateRequiredText(password); err != nil {
		return "", store.User{}, err
	}
	user, err := s.store.GetUserByEmail(ctx, nil, email)
	if err != nil {
		if err == store.ErrNotFound {
			return "", store.User{}, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeInvalidCredentials, "invalid email or password")
		}
		return "", store.User{}, httpapi.InternalError(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", store.User{}, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeInvalidCredentials, "invalid email or password")
	}
	now := s.currentTime()
	sessionID := uuid.NewString()
	token, expiresAt, err := s.tokens.Issue(user.ID, sessionID, now)
	if err != nil {
		return "", store.User{}, httpapi.InternalError(err)
	}
	if _, err := s.store.CreateSession(ctx, nil, store.CreateSessionParams{ID: sessionID, UserID: user.ID, ExpiresAt: expiresAt, CreatedAt: now, LastSeenAt: now}); err != nil {
		return "", store.User{}, httpapi.InternalError(err)
	}
	return token, user, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	err := s.store.RevokeSession(ctx, nil, sessionID, s.currentTime())
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) SendRegisterCode(ctx context.Context, email string) (int, error) {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return 0, err
	}
	if _, err := s.store.GetUserByEmail(ctx, nil, email); err == nil {
		return 0, httpapi.NewError(http.StatusConflict, httpapi.CodeEmailRegistered, "email already registered")
	} else if err != nil && err != store.ErrNotFound {
		return 0, httpapi.InternalError(err)
	}
	return s.sendVerificationCode(ctx, email, store.CodePurposeRegister)
}

func (s *Service) VerifyRegisterCode(ctx context.Context, email, code string) error {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return err
	}
	return s.validateCode(ctx, nil, email, store.CodePurposeRegister, code, false)
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (string, store.User, error) {
	input.Email = NormalizeEmail(input.Email)
	input.Nickname = strings.TrimSpace(input.Nickname)
	if err := ValidateEmail(input.Email); err != nil {
		return "", store.User{}, err
	}
	if err := ValidatePassword(input.Password); err != nil {
		return "", store.User{}, err
	}
	if err := ValidateVerificationCode(input.Code); err != nil {
		return "", store.User{}, err
	}
	if err := ValidateRequiredText(input.Nickname); err != nil {
		return "", store.User{}, err
	}
	input.Department = strings.TrimSpace(input.Department)
	input.Grade = strings.TrimSpace(input.Grade)
	if !meta.IsValidMajorDescription(input.Department) {
		return "", store.User{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid major description")
	}
	if !meta.IsValidGrade(input.Grade) {
		return "", store.User{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid grade")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.bcryptCost)
	if err != nil {
		return "", store.User{}, httpapi.InternalError(err)
	}
	var token string
	var user store.User
	now := s.currentTime()
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.validateCode(ctx, tx, input.Email, store.CodePurposeRegister, input.Code, true); err != nil {
			return err
		}
		schoolKey, _ := school.KeyByEmail(input.Email)
		createdUser, err := s.store.CreateUser(ctx, tx, store.CreateUserParams{
			Email:        input.Email,
			SchoolKey:    schoolKey,
			PasswordHash: string(passwordHash),
			Nickname:     strings.TrimSpace(input.Nickname),
			Department:   input.Department,
			Grade:        input.Grade,
			Now:          now,
		})
		if err != nil {
			if isDuplicateEmailErr(err) {
				return httpapi.NewError(http.StatusConflict, httpapi.CodeEmailRegistered, "email already registered")
			}
			if isDuplicateNicknameErr(err) {
				return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "nickname already exists")
			}
			return httpapi.InternalError(err)
		}
		sessionID := uuid.NewString()
		issuedToken, expiresAt, err := s.tokens.Issue(createdUser.ID, sessionID, now)
		if err != nil {
			return httpapi.InternalError(err)
		}
		if _, err := s.store.CreateSession(ctx, tx, store.CreateSessionParams{ID: sessionID, UserID: createdUser.ID, ExpiresAt: expiresAt, CreatedAt: now, LastSeenAt: now}); err != nil {
			return httpapi.InternalError(err)
		}
		token = issuedToken
		user = createdUser
		return nil
	})
	if err != nil {
		return "", store.User{}, err
	}
	s.refreshRegisteredUserCount(ctx)
	return token, user, nil
}

func (s *Service) SendForgotPasswordCode(ctx context.Context, email string) (int, error) {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return 0, err
	}
	if _, err := s.store.GetUserByEmail(ctx, nil, email); err != nil {
		if err == store.ErrNotFound {
			return 0, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "user not found")
		}
		return 0, httpapi.InternalError(err)
	}
	return s.sendVerificationCode(ctx, email, store.CodePurposePasswordForgot)
}

func (s *Service) ResetForgotPassword(ctx context.Context, email, code, newPassword string) error {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return err
	}
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}
	user, err := s.store.GetUserByEmail(ctx, nil, email)
	if err != nil {
		if err == store.ErrNotFound {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "user not found")
		}
		return httpapi.InternalError(err)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.bcryptCost)
	if err != nil {
		return httpapi.InternalError(err)
	}
	return s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.validateCode(ctx, tx, email, store.CodePurposePasswordForgot, code, true); err != nil {
			return err
		}
		now := s.currentTime()
		if err := s.store.UpdateUserPassword(ctx, tx, user.ID, string(passwordHash), now); err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.RevokeAllSessionsByUser(ctx, tx, user.ID, now); err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	})
}

func (s *Service) SendChangePasswordCode(ctx context.Context, email string) (int, error) {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return 0, err
	}
	return s.sendVerificationCode(ctx, email, store.CodePurposePasswordChange)
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, email, code, newPassword string) error {
	email = NormalizeEmail(email)
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.bcryptCost)
	if err != nil {
		return httpapi.InternalError(err)
	}
	return s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.validateCode(ctx, tx, email, store.CodePurposePasswordChange, code, true); err != nil {
			return err
		}
		now := s.currentTime()
		if err := s.store.UpdateUserPassword(ctx, tx, userID, string(passwordHash), now); err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.RevokeAllSessionsByUser(ctx, tx, userID, now); err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	})
}

func (s *Service) SendDeleteAccountCode(ctx context.Context, email string) (int, error) {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return 0, err
	}
	return s.sendVerificationCode(ctx, email, store.CodePurposeAccountDelete)
}

func (s *Service) DeleteAccount(ctx context.Context, userID int64, email, code string) error {
	email = NormalizeEmail(email)
	if err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.validateCode(ctx, tx, email, store.CodePurposeAccountDelete, code, true); err != nil {
			return err
		}
		now := s.currentTime()
		if err := s.store.DeleteQuestionnaireDataByUser(ctx, tx, userID); err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.DeleteAllUserContactMethods(ctx, tx, userID); err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.DeleteUserMatchResultsByUserReference(ctx, tx, userID, now); err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.DeleteSessionsByUser(ctx, tx, userID); err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.DeleteVerificationCodesByEmail(ctx, tx, email); err != nil {
			return httpapi.InternalError(err)
		}
		if err := s.store.DeleteUser(ctx, tx, userID); err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	}); err != nil {
		return err
	}

	s.clearVerificationCodeCache(ctx, email)
	s.refreshRegisteredUserCount(ctx)
	return nil
}

func (s *Service) sendVerificationCode(ctx context.Context, email string, purpose store.CodePurpose) (int, error) {
	if err := s.ensureCodeCooldown(ctx, email, purpose); err != nil {
		return 0, err
	}
	code, err := GenerateVerificationCode()
	if err != nil {
		return 0, httpapi.InternalError(err)
	}
	now := s.currentTime()
	codeHash := HashVerificationCode(string(purpose), email, code, s.codeSecret)
	message := verificationMail(purpose, code, s.codeTTL)
	if s.codeCache != nil {
		record := CachedVerificationCode{CodeHash: codeHash, ExpiresAt: now.Add(s.codeTTL), SentAt: now}
		if err := s.codeCache.Store(ctx, email, purpose, record, s.codeTTL); err != nil {
			return 0, httpapi.InternalError(err)
		}
		if err := s.mailer.Send(ctx, mail.Message{To: email, Subject: message.subject, Text: message.text}); err != nil {
			_ = s.codeCache.Delete(ctx, email, purpose)
			return 0, httpapi.InternalError(err)
		}
	} else {
		record, err := s.store.CreateVerificationCode(ctx, nil, store.CreateVerificationCodeParams{Email: email, Purpose: purpose, CodeHash: codeHash, ExpiresAt: now.Add(s.codeTTL), SentAt: now})
		if err != nil {
			return 0, httpapi.InternalError(err)
		}
		if err := s.mailer.Send(ctx, mail.Message{To: email, Subject: message.subject, Text: message.text}); err != nil {
			_ = s.store.DeleteVerificationCode(ctx, nil, record.ID)
			return 0, httpapi.InternalError(err)
		}
	}
	if s.logger != nil {
		s.logger.Info("verification code issued", "email", email, "purpose", purpose)
	}
	return int(s.codeTTL.Seconds()), nil
}

func (s *Service) VerificationCodeCooldownSeconds() int {
	seconds := int(math.Ceil(s.codeCooldown.Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}

func (s *Service) ensureCodeCooldown(ctx context.Context, email string, purpose store.CodePurpose) error {
	if s.codeCache != nil {
		record, err := s.codeCache.Load(ctx, email, purpose)
		if err != nil {
			if errors.Is(err, ErrVerificationCodeCacheNotFound) {
				return nil
			}
			return httpapi.InternalError(err)
		}
		if s.currentTime().Before(record.SentAt.Add(s.codeCooldown)) {
			return httpapi.NewError(http.StatusTooManyRequests, httpapi.CodeVerificationCodeTooFrequent, "verification code sent too frequently")
		}
		return nil
	}
	latest, err := s.store.GetLatestVerificationCode(ctx, nil, email, purpose)
	if err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return httpapi.InternalError(err)
	}
	if s.currentTime().Before(latest.SentAt.Add(s.codeCooldown)) {
		return httpapi.NewError(http.StatusTooManyRequests, httpapi.CodeVerificationCodeTooFrequent, "verification code sent too frequently")
	}
	return nil
}

func (s *Service) validateCode(ctx context.Context, tx *sql.Tx, email string, purpose store.CodePurpose, code string, consume bool) error {
	if err := ValidateVerificationCode(code); err != nil {
		return err
	}
	if s.codeCache != nil {
		return s.validateCachedCode(ctx, email, purpose, code, consume)
	}
	record, err := s.store.GetLatestVerificationCode(ctx, tx, email, purpose)
	if err != nil {
		if err == store.ErrNotFound {
			return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidVerificationCode, "invalid verification code")
		}
		return httpapi.InternalError(err)
	}
	if record.ConsumedAt.Valid {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidVerificationCode, "invalid verification code")
	}
	if s.currentTime().After(record.ExpiresAt) {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeVerificationCodeExpired, "verification code expired")
	}
	if record.CodeHash != HashVerificationCode(string(purpose), email, code, s.codeSecret) {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidVerificationCode, "invalid verification code")
	}
	now := s.currentTime()
	if consume {
		if err := s.store.ConsumeVerificationCode(ctx, tx, record.ID, now); err != nil {
			return httpapi.InternalError(err)
		}
		return nil
	}
	if err := s.store.MarkVerificationCodeVerified(ctx, tx, record.ID, now); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) validateCachedCode(ctx context.Context, email string, purpose store.CodePurpose, code string, consume bool) error {
	record, err := s.codeCache.Load(ctx, email, purpose)
	if err != nil {
		if errors.Is(err, ErrVerificationCodeCacheNotFound) {
			return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidVerificationCode, "invalid verification code")
		}
		return httpapi.InternalError(err)
	}
	now := s.currentTime()
	if record.ConsumedAt != nil {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidVerificationCode, "invalid verification code")
	}
	if !record.ExpiresAt.After(now) {
		_ = s.codeCache.Delete(ctx, email, purpose)
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeVerificationCodeExpired, "verification code expired")
	}
	if record.CodeHash != HashVerificationCode(string(purpose), email, code, s.codeSecret) {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidVerificationCode, "invalid verification code")
	}
	if consume {
		record.ConsumedAt = &now
	} else if record.VerifiedAt == nil {
		record.VerifiedAt = &now
	}
	remainingTTL := record.ExpiresAt.Sub(now)
	if remainingTTL <= 0 {
		_ = s.codeCache.Delete(ctx, email, purpose)
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeVerificationCodeExpired, "verification code expired")
	}
	if err := s.codeCache.Store(ctx, email, purpose, record, remainingTTL); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func isDuplicateEmailErr(err error) bool {
	return strings.Contains(err.Error(), "ux_users_email_active")
}

func isDuplicateNicknameErr(err error) bool {
	return strings.Contains(err.Error(), "ux_users_nickname_active")
}

type verificationEmail struct{ subject, text string }

func verificationMail(purpose store.CodePurpose, code string, ttl time.Duration) verificationEmail {
	minutes := int(ttl.Minutes())
	if minutes <= 0 {
		minutes = 5
	}
	switch purpose {
	case store.CodePurposeRegister:
		return verificationEmail{subject: "Soulink registration code", text: fmt.Sprintf("Your Soulink registration code is %s. It is valid for %d minutes.", code, minutes)}
	case store.CodePurposePasswordForgot:
		return verificationEmail{subject: "Soulink password reset code", text: fmt.Sprintf("Your Soulink password reset code is %s. It is valid for %d minutes.", code, minutes)}
	case store.CodePurposePasswordChange:
		return verificationEmail{subject: "Soulink password change code", text: fmt.Sprintf("Your Soulink password change code is %s. It is valid for %d minutes.", code, minutes)}
	case store.CodePurposeAccountDelete:
		return verificationEmail{subject: "Soulink account deletion code", text: fmt.Sprintf("Your Soulink account deletion code is %s. It is valid for %d minutes.", code, minutes)}
	default:
		return verificationEmail{subject: "Soulink verification code", text: fmt.Sprintf("Your Soulink verification code is %s. It is valid for %d minutes.", code, minutes)}
	}
}

func (s *Service) WithRegisteredUserCountRefresher(refresher RegisteredUserCountRefresher) *Service {
	s.registeredUserCountRefresher = refresher
	return s
}

func (s *Service) refreshRegisteredUserCount(ctx context.Context) {
	if s == nil || s.registeredUserCountRefresher == nil {
		return
	}
	if err := s.registeredUserCountRefresher.RefreshRegisteredUserCount(ctx); err != nil && s.logger != nil {
		s.logger.Warn("refresh registered user count failed", "error", err)
	}
}
func (s *Service) currentTime() time.Time {
	if s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}

func (s *Service) clearVerificationCodeCache(ctx context.Context, email string) {
	if s.codeCache == nil {
		return
	}

	purposes := []store.CodePurpose{
		store.CodePurposeRegister,
		store.CodePurposePasswordForgot,
		store.CodePurposePasswordChange,
		store.CodePurposeAccountDelete,
	}
	for _, purpose := range purposes {
		if err := s.codeCache.Delete(ctx, email, purpose); err != nil && !errors.Is(err, ErrVerificationCodeCacheNotFound) && s.logger != nil {
			s.logger.Warn("clear verification code cache failed", "email", email, "purpose", purpose, "error", err)
		}
	}
}
