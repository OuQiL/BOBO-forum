package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"post/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	ErrLockAcquireFailed = errors.New("failed to acquire lock")
	ErrLockTimeout       = errors.New("lock wait timeout")
)

const (
	LockKeyPrefix   = "post:lock:"
	LockTTL         = 400 * time.Millisecond
	LockWaitTime    = 20 * time.Millisecond
	LockWaitTimeout = 5 * time.Second
	CacheKeyPrefix  = "post:cache:"
	CacheTTL        = 30 * time.Second
)

type CacheLoader struct {
	redis      *redis.Redis
	localCache *HotPostLocalCache
}

func NewCacheLoader(rds *redis.Redis, localCache *HotPostLocalCache) *CacheLoader {
	return &CacheLoader{
		redis:      rds,
		localCache: localCache,
	}
}

func (l *CacheLoader) GetOrLoad(ctx context.Context, key int64, loadFn func(ctx context.Context) (*model.Post, error)) (*model.Post, error) {
	if post, found := l.localCache.Get(key); found {
		return post, nil
	}

	if post, found := l.getRedisCache(ctx, key); found {
		if l.localCache.ShouldCache(post.ViewCount) {
			l.localCache.Set(post)
		}
		return post, nil
	}

	return l.getWithLock(ctx, key, loadFn)
}

func (l *CacheLoader) getWithLock(ctx context.Context, key int64, loadFn func(ctx context.Context) (*model.Post, error)) (*model.Post, error) {
	lockKey := fmt.Sprintf("%s%d", LockKeyPrefix, key)
	lockValue := fmt.Sprintf("%d:%d", time.Now().UnixNano(), key)

	acquired, err := l.acquireLock(ctx, lockKey, lockValue)
	if err != nil {
		logx.Errorf("Failed to acquire lock for post %d: %v, loading directly", key, err)
		return loadFn(ctx)
	}

	if acquired {
		defer l.releaseLock(ctx, lockKey, lockValue)

		if post, found := l.localCache.Get(key); found {
			return post, nil
		}
		if post, found := l.getRedisCache(ctx, key); found {
			if l.localCache.ShouldCache(post.ViewCount) {
				l.localCache.Set(post)
			}
			return post, nil
		}

		post, err := loadFn(ctx)
		if err != nil {
			return nil, err
		}

		if post != nil {
			if cachedPost, found := l.localCache.Get(key); found {
				return cachedPost, nil
			}
			if cachedPost, found := l.getRedisCache(ctx, key); found {
				if l.localCache.ShouldCache(cachedPost.ViewCount) {
					l.localCache.Set(cachedPost)
				}
				return cachedPost, nil
			}

			if l.localCache.ShouldCache(post.ViewCount) {
				l.localCache.Set(post)
				l.setRedisCache(ctx, key, post)
			}
		}

		return post, nil
	}

	return l.waitForCache(ctx, key, loadFn)
}

func (l *CacheLoader) acquireLock(ctx context.Context, key, value string) (bool, error) {
	script := `
local ok = redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
if ok then
	return 1
else
	return 0
end`
	result, err := l.redis.Eval(script, []string{key}, value, int(LockTTL.Milliseconds()))
	if err != nil {
		return false, err
	}
	return result == "1", nil
}

func (l *CacheLoader) releaseLock(ctx context.Context, key, value string) {
	script := `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`
	_, err := l.redis.Eval(script, []string{key}, value)
	if err != nil {
		logx.Errorf("Failed to release lock %s: %v", key, err)
	}
}

func (l *CacheLoader) waitForCache(ctx context.Context, key int64, loadFn func(ctx context.Context) (*model.Post, error)) (*model.Post, error) {
	timeout := time.After(LockWaitTimeout)
	ticker := time.NewTicker(LockWaitTime)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			logx.Errorf("Lock wait timeout for post %d, loading directly", key)
			return loadFn(ctx)
		case <-ticker.C:
			if post, found := l.localCache.Get(key); found {
				return post, nil
			}
			if post, found := l.getRedisCache(ctx, key); found {
				if l.localCache.ShouldCache(post.ViewCount) {
					l.localCache.Set(post)
				}
				return post, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (l *CacheLoader) getRedisCache(ctx context.Context, key int64) (*model.Post, bool) {
	cacheKey := fmt.Sprintf("%s%d", CacheKeyPrefix, key)
	val, err := l.redis.Get(cacheKey)
	if err != nil || val == "" {
		return nil, false
	}

	var post model.Post
	if err := json.Unmarshal([]byte(val), &post); err != nil {
		logx.Errorf("Failed to unmarshal post from redis cache: %v", err)
		return nil, false
	}

	return &post, true
}

func (l *CacheLoader) setRedisCache(ctx context.Context, key int64, post *model.Post) {
	cacheKey := fmt.Sprintf("%s%d", CacheKeyPrefix, key)
	data, err := json.Marshal(post)
	if err != nil {
		logx.Errorf("Failed to marshal post for redis cache: %v", err)
		return
	}
	if err := l.redis.Setex(cacheKey, string(data), int(CacheTTL.Seconds())); err != nil {
		logx.Errorf("Failed to set redis cache: %v", err)
	}
}
