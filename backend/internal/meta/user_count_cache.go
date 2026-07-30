package meta

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	platformcache "soulink-backend/internal/platform/cache"
	"soulink-backend/internal/school"
)

const registeredUserCountCacheKeyPrefix = "meta:registered_user_count:"

type RedisRegisteredUserCountCache struct {
	cache *platformcache.Redis
}

func NewRedisRegisteredUserCountCache(cache *platformcache.Redis) *RedisRegisteredUserCountCache {
	return &RedisRegisteredUserCountCache{cache: cache}
}

func (c *RedisRegisteredUserCountCache) Load(ctx context.Context, schoolName school.Name) (int, error) {
	payload, err := c.cache.Get(ctx, registeredUserCountCacheKey(schoolName))
	if err != nil {
		if errors.Is(err, platformcache.ErrNotFound) {
			return 0, ErrRegisteredUserCountCacheNotFound
		}
		return 0, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(payload))
	if err != nil {
		return 0, fmt.Errorf("parse cached registered user count: %w", err)
	}
	return count, nil
}

func (c *RedisRegisteredUserCountCache) Store(ctx context.Context, schoolName school.Name, count int, ttl time.Duration) error {
	return c.cache.Set(ctx, registeredUserCountCacheKey(schoolName), strconv.Itoa(count), ttl)
}

func (c *RedisRegisteredUserCountCache) Delete(ctx context.Context, schoolName school.Name) error {
	return c.cache.Delete(ctx, registeredUserCountCacheKey(schoolName))
}

func registeredUserCountCacheKey(schoolName school.Name) string {
	return registeredUserCountCacheKeyPrefix + string(schoolName)
}
