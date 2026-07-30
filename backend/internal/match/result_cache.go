package match

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	platformcache "soulink-backend/internal/platform/cache"
)

var ErrResultCacheNotFound = errors.New("match: current result cache not found")

type ResultCache interface {
	LoadCurrent(ctx context.Context, cycleID, userID int64) (ResultView, error)
	StoreCurrent(ctx context.Context, cycleID, userID int64, view ResultView, ttl time.Duration) error
	DeleteCurrent(ctx context.Context, cycleID, userID int64) error
}

type RedisResultCache struct {
	cache *platformcache.Redis
}

func NewRedisResultCache(cache *platformcache.Redis) *RedisResultCache {
	return &RedisResultCache{cache: cache}
}

func (c *RedisResultCache) LoadCurrent(ctx context.Context, cycleID, userID int64) (ResultView, error) {
	payload, err := c.cache.Get(ctx, currentResultCacheKey(cycleID, userID))
	if err != nil {
		if errors.Is(err, platformcache.ErrNotFound) {
			return ResultView{}, ErrResultCacheNotFound
		}
		return ResultView{}, err
	}
	var view ResultView
	if err := json.Unmarshal([]byte(payload), &view); err != nil {
		return ResultView{}, fmt.Errorf("unmarshal cached current result: %w", err)
	}
	return view, nil
}

func (c *RedisResultCache) StoreCurrent(ctx context.Context, cycleID, userID int64, view ResultView, ttl time.Duration) error {
	payload, err := json.Marshal(view)
	if err != nil {
		return fmt.Errorf("marshal cached current result: %w", err)
	}
	return c.cache.Set(ctx, currentResultCacheKey(cycleID, userID), string(payload), ttl)
}

func (c *RedisResultCache) DeleteCurrent(ctx context.Context, cycleID, userID int64) error {
	return c.cache.Delete(ctx, currentResultCacheKey(cycleID, userID))
}

func currentResultCacheKey(cycleID, userID int64) string {
	return fmt.Sprintf("match:current:%d:%d", cycleID, userID)
}
