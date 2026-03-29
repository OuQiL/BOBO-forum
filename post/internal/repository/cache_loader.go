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
	CacheKeyPrefix = "post:cache:"
	CacheTTL       = 30 * time.Second
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

	post, err := loadFn(ctx)
	if err != nil {
		return nil, err
	}

	if post != nil && l.localCache.ShouldCache(post.ViewCount) {
		l.localCache.Set(post)
		l.setRedisCache(ctx, key, post)
	}

	return post, nil
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
