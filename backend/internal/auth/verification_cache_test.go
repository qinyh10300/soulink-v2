package auth

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/mail"
	"soulink-backend/internal/store"
)

type captureSender struct{ messages []mail.Message }

func (s *captureSender) Send(_ context.Context, message mail.Message) error {
	s.messages = append(s.messages, message)
	return nil
}

func (s *captureSender) latestCode(t *testing.T) string {
	t.Helper()
	if len(s.messages) == 0 {
		t.Fatal("no messages sent")
	}
	matches := regexp.MustCompile(`\b(\d{6})\b`).FindStringSubmatch(s.messages[len(s.messages)-1].Text)
	if len(matches) != 2 {
		t.Fatalf("failed to extract verification code from %q", s.messages[len(s.messages)-1].Text)
	}
	return matches[1]
}

type memoryVerificationCodeCache struct {
	records map[string]CachedVerificationCode
}

func newMemoryVerificationCodeCache() *memoryVerificationCodeCache {
	return &memoryVerificationCodeCache{records: make(map[string]CachedVerificationCode)}
}

func (c *memoryVerificationCodeCache) Load(_ context.Context, email string, purpose store.CodePurpose) (CachedVerificationCode, error) {
	record, ok := c.records[email+"|"+string(purpose)]
	if !ok {
		return CachedVerificationCode{}, ErrVerificationCodeCacheNotFound
	}
	return record, nil
}

func (c *memoryVerificationCodeCache) Store(_ context.Context, email string, purpose store.CodePurpose, record CachedVerificationCode, _ time.Duration) error {
	c.records[email+"|"+string(purpose)] = record
	return nil
}

func (c *memoryVerificationCodeCache) Delete(_ context.Context, email string, purpose store.CodePurpose) error {
	delete(c.records, email+"|"+string(purpose))
	return nil
}

func TestSendVerificationCodeUsesCacheAndEnforcesCooldown(t *testing.T) {
	now := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	sender := &captureSender{}
	cache := newMemoryVerificationCodeCache()
	service := &Service{mailer: sender, codeTTL: 5 * time.Minute, codeCooldown: time.Minute, codeSecret: "secret", codeCache: cache, nowFunc: func() time.Time { return now }}
	expireSeconds, err := service.sendVerificationCode(context.Background(), "test@example.com", store.CodePurposeRegister)
	if err != nil {
		t.Fatalf("sendVerificationCode() error: %v", err)
	}
	if expireSeconds != 300 {
		t.Fatalf("expireSeconds = %d, want 300", expireSeconds)
	}
	record, err := cache.Load(context.Background(), "test@example.com", store.CodePurposeRegister)
	if err != nil {
		t.Fatalf("cache.Load() error: %v", err)
	}
	if record.CodeHash == "" {
		t.Fatal("record.CodeHash is empty")
	}
	_, err = service.sendVerificationCode(context.Background(), "test@example.com", store.CodePurposeRegister)
	if err == nil {
		t.Fatal("second sendVerificationCode() expected cooldown error")
	}
	apiErr := &httpapi.AppError{}
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *httpapi.Error", err)
	}
	if apiErr.Code != httpapi.CodeVerificationCodeTooFrequent {
		t.Fatalf("apiErr.Code = %d, want %d", apiErr.Code, httpapi.CodeVerificationCodeTooFrequent)
	}
}

func TestValidateCodeUsesCacheAndMarksVerifiedThenConsumed(t *testing.T) {
	now := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	sender := &captureSender{}
	cache := newMemoryVerificationCodeCache()
	service := &Service{mailer: sender, codeTTL: 5 * time.Minute, codeCooldown: time.Minute, codeSecret: "secret", codeCache: cache, nowFunc: func() time.Time { return now }}
	if _, err := service.sendVerificationCode(context.Background(), "test@example.com", store.CodePurposeRegister); err != nil {
		t.Fatalf("sendVerificationCode() error: %v", err)
	}
	code := sender.latestCode(t)
	now = now.Add(30 * time.Second)
	if err := service.validateCode(context.Background(), nil, "test@example.com", store.CodePurposeRegister, code, false); err != nil {
		t.Fatalf("validateCode verify error: %v", err)
	}
	record, err := cache.Load(context.Background(), "test@example.com", store.CodePurposeRegister)
	if err != nil {
		t.Fatalf("cache.Load() error: %v", err)
	}
	if record.VerifiedAt == nil {
		t.Fatal("record.VerifiedAt is nil, want non-nil")
	}
	if record.ConsumedAt != nil {
		t.Fatal("record.ConsumedAt is non-nil before consume")
	}
	now = now.Add(30 * time.Second)
	if err := service.validateCode(context.Background(), nil, "test@example.com", store.CodePurposeRegister, code, true); err != nil {
		t.Fatalf("validateCode consume error: %v", err)
	}
	record, err = cache.Load(context.Background(), "test@example.com", store.CodePurposeRegister)
	if err != nil {
		t.Fatalf("cache.Load() error: %v", err)
	}
	if record.ConsumedAt == nil {
		t.Fatal("record.ConsumedAt is nil, want non-nil")
	}
	err = service.validateCode(context.Background(), nil, "test@example.com", store.CodePurposeRegister, code, true)
	if err == nil {
		t.Fatal("validateCode() expected error for consumed code")
	}
	apiErr := &httpapi.AppError{}
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *httpapi.Error", err)
	}
	if apiErr.Code != httpapi.CodeInvalidVerificationCode {
		t.Fatalf("apiErr.Code = %d, want %d", apiErr.Code, httpapi.CodeInvalidVerificationCode)
	}
}
