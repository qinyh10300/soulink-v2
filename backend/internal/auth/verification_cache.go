package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	platformcache "soulink-backend/internal/platform/cache"
	"soulink-backend/internal/store"
)

var ErrVerificationCodeCacheNotFound = errors.New("auth: verification code cache not found")

type CachedVerificationCode struct {
	CodeHash   string     `json:"codeHash"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	SentAt     time.Time  `json:"sentAt"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty"`
	ConsumedAt *time.Time `json:"consumedAt,omitempty"`
}

type VerificationCodeCache interface {
	Load(ctx context.Context, email string, purpose store.CodePurpose) (CachedVerificationCode, error)
	Store(ctx context.Context, email string, purpose store.CodePurpose, record CachedVerificationCode, ttl time.Duration) error
	Delete(ctx context.Context, email string, purpose store.CodePurpose) error
}

type RedisVerificationCodeCache struct {
	cache *platformcache.Redis
}

func NewRedisVerificationCodeCache(cache *platformcache.Redis) *RedisVerificationCodeCache {
	return &RedisVerificationCodeCache{cache: cache}
}

func (c *RedisVerificationCodeCache) Load(ctx context.Context, email string, purpose store.CodePurpose) (CachedVerificationCode, error) {
	payload, err := c.cache.Get(ctx, verificationCodeCacheKey(email, purpose))
	if err != nil {
		if errors.Is(err, platformcache.ErrNotFound) {
			return CachedVerificationCode{}, ErrVerificationCodeCacheNotFound
		}
		return CachedVerificationCode{}, err
	}
	var record CachedVerificationCode
	if err := json.Unmarshal([]byte(payload), &record); err != nil {
		return CachedVerificationCode{}, fmt.Errorf("unmarshal cached verification code: %w", err)
	}
	return record, nil
}

func (c *RedisVerificationCodeCache) Store(ctx context.Context, email string, purpose store.CodePurpose, record CachedVerificationCode, ttl time.Duration) error {
	if ttl <= 0 {
		return c.Delete(ctx, email, purpose)
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal cached verification code: %w", err)
	}
	return c.cache.Set(ctx, verificationCodeCacheKey(email, purpose), string(payload), ttl)
}

func (c *RedisVerificationCodeCache) Delete(ctx context.Context, email string, purpose store.CodePurpose) error {
	return c.cache.Delete(ctx, verificationCodeCacheKey(email, purpose))
}

func verificationCodeCacheKey(email string, purpose store.CodePurpose) string {
	return fmt.Sprintf("auth:verification_code:%s:%s", purpose, email)
}
