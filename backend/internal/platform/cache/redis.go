package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"

	"soulink-backend/internal/config"
)

var ErrNotFound = errors.New("cache: not found")

type Redis struct {
	client *redis.Client
	prefix string
}

func Open(ctx context.Context, cfg config.RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB})
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Redis{client: client, prefix: cfg.Prefix}, nil
}

func (r *Redis) Close() error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Close()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, r.prefixed(key)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("redis get %s: %w", key, err)
	}
	return value, nil
}

func (r *Redis) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, r.prefixed(key), value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set %s: %w", key, err)
	}
	return nil
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, r.prefixed(key)).Err(); err != nil {
		return fmt.Errorf("redis delete %s: %w", key, err)
	}
	return nil
}

func (r *Redis) prefixed(key string) string {
	prefix := strings.TrimSpace(r.prefix)
	if prefix == "" {
		return key
	}
	return prefix + key
}
