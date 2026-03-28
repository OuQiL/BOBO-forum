package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	EmptyValueMarker     = "NULL"
	EmptyValuePrefix     = "post:empty:"
	EmptyValueDefaultTTL = 60
)

type EmptyValueCache struct {
	redis      *redis.Redis
	defaultTTL int
	localCache *syncMapCache
}

type syncMapCache struct {
	data map[string]emptyCacheEntry
	mu   sync.RWMutex
}

type emptyCacheEntry struct {
	value     bool
	expiresAt time.Time
}

func NewSyncMapCache() *syncMapCache {
	return &syncMapCache{
		data: make(map[string]emptyCacheEntry),
	}
}

func (c *syncMapCache) Set(key string, value bool, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = emptyCacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *syncMapCache) Get(key string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.data[key]
	if !ok {
		return false, false
	}
	if time.Now().After(entry.expiresAt) {
		return false, false
	}
	return entry.value, true
}

func (c *syncMapCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func NewEmptyValueCache(redis *redis.Redis, defaultTTL int) *EmptyValueCache {
	if defaultTTL <= 0 {
		defaultTTL = EmptyValueDefaultTTL
	}
	return &EmptyValueCache{
		redis:      redis,
		defaultTTL: defaultTTL,
		localCache: NewSyncMapCache(),
	}
}

func (c *EmptyValueCache) CacheEmpty(ctx context.Context, postID int64, ttl ...int) error {
	cacheTTL := c.defaultTTL
	if len(ttl) > 0 && ttl[0] > 0 {
		cacheTTL = ttl[0]
	}

	key := c.getEmptyKey(postID)

	c.localCache.Set(key, true, time.Duration(cacheTTL)*time.Second)

	if c.redis != nil {
		err := c.redis.Setex(key, EmptyValueMarker, cacheTTL)
		if err != nil {
			return fmt.Errorf("cache empty value to redis failed: %w", err)
		}
	}

	return nil
}

func (c *EmptyValueCache) IsEmptyCached(ctx context.Context, postID int64) bool {
	key := c.getEmptyKey(postID)

	if value, found := c.localCache.Get(key); found && value {
		return true
	}

	if c.redis == nil {
		return false
	}

	val, err := c.redis.Get(key)
	if err != nil {
		return false
	}

	if val == EmptyValueMarker {
		c.localCache.Set(key, true, time.Duration(c.defaultTTL)*time.Second)
		return true
	}

	return false
}

func (c *EmptyValueCache) RemoveEmptyCache(ctx context.Context, postID int64) error {
	key := c.getEmptyKey(postID)

	c.localCache.Delete(key)

	if c.redis != nil {
		_, err := c.redis.Del(key)
		if err != nil {
			return fmt.Errorf("remove empty value cache from redis failed: %w", err)
		}
	}

	return nil
}

func (c *EmptyValueCache) getEmptyKey(postID int64) string {
	return fmt.Sprintf("%s%d", EmptyValuePrefix, postID)
}

func (c *EmptyValueCache) CachePostNotFound(ctx context.Context, postID int64) error {
	return c.CacheEmpty(ctx, postID)
}

func (c *EmptyValueCache) ClearAllEmptyCache(ctx context.Context) error {
	c.localCache = NewSyncMapCache()

	if c.redis != nil {
		pattern := EmptyValuePrefix + "*"
		keys, err := c.redis.Keys(pattern)
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			_, err = c.redis.Del(keys...)
			return err
		}
	}

	return nil
}

type CacheableValue struct {
	Value     interface{}
	IsEmpty   bool
	ExpiresAt int64
}

func (c *EmptyValueCache) GetWithEmptyCheck(ctx context.Context, key string, fetchFn func() (interface{}, error), ttl int) (interface{}, error) {
	if c.redis == nil {
		return fetchFn()
	}

	val, err := c.redis.Get(key)
	if err == nil {
		if val == EmptyValueMarker {
			return nil, nil
		}

		var result CacheableValue
		if err := json.Unmarshal([]byte(val), &result); err == nil {
			if result.IsEmpty {
				return nil, nil
			}
			return result.Value, nil
		}

		return val, nil
	}

	result, err := fetchFn()
	if err != nil {
		return nil, err
	}

	if result == nil {
		if err := c.CacheEmpty(ctx, extractPostIDFromKey(key), ttl); err != nil {
			return nil, err
		}
		return nil, nil
	}

	cacheVal := CacheableValue{
		Value:   result,
		IsEmpty: false,
	}
	data, _ := json.Marshal(cacheVal)
	if err := c.redis.Setex(key, string(data), ttl); err != nil {
		return result, nil
	}

	return result, nil
}

func extractPostIDFromKey(key string) int64 {
	var id int64
	fmt.Sscanf(key, "post:detail:%d", &id)
	return id
}
