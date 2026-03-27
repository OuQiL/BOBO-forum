package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	PostReadCountKey = "post:view:%d"
	PostContentKey   = "post:content:%d"
)

const (
	addWhenCreateScript = `
		local key = KEYS[1]
		local userID = ARGV[1]
		local ttl = tonumber(ARGV[2])
		redis.call('PFADD', key, userID)
		redis.call('EXPIRE', key, ttl)
		return 1
	`

	addReadWithExpireScript = `
		local key = KEYS[1]
		local userID = ARGV[1]
		local ttl = tonumber(ARGV[2])
		redis.call('PFADD', key, userID)
		if ttl > 0 then
			redis.call('EXPIRE', key, ttl)
		end
		return 1
	`

	incrWithExpireScript = `
		local key = KEYS[1]
		local ttl = tonumber(ARGV[1])
		local result = redis.call('INCR', key)
		redis.call('EXPIRE', key, ttl)
		return result
	`
)

func AddWhenCreate(ctx context.Context, r *redis.Redis, postID, userID int64) error {
	key := fmt.Sprintf(PostReadCountKey, postID)
	ttl := int(24 * time.Hour / time.Second * 7)
	_, err := r.EvalCtx(ctx, addWhenCreateScript, []string{key}, userID, ttl)
	return err
}

func AddRead(ctx context.Context, r *redis.Redis, postID, userID int64) error {
	key := fmt.Sprintf(PostReadCountKey, postID)
	_, err := r.Pfadd(key, fmt.Sprintf("%d", userID))
	return err
}

func AddReadWithExpire(ctx context.Context, r *redis.Redis, postID, userID int64, ttl time.Duration) error {
	key := fmt.Sprintf(PostReadCountKey, postID)
	ttlSeconds := int(ttl / time.Second)
	_, err := r.EvalCtx(ctx, addReadWithExpireScript, []string{key}, userID, ttlSeconds)
	return err
}

func GetRC(ctx context.Context, r *redis.Redis, postID int64) (int64, error) {
	key := fmt.Sprintf(PostReadCountKey, postID)
	return r.Pfcount(key)
}

func IncrViewCount(ctx context.Context, r *redis.Redis, postID int64) (int64, error) {
	key := fmt.Sprintf(PostReadCountKey, postID)
	return r.Incr(key)
}

func IncrViewCountWithExpire(ctx context.Context, r *redis.Redis, postID int64, ttl time.Duration) (int64, error) {
	key := fmt.Sprintf(PostReadCountKey, postID)
	ttlSeconds := int(ttl / time.Second)
	result, err := r.EvalCtx(ctx, incrWithExpireScript, []string{key}, ttlSeconds)
	if err != nil {
		return 0, err
	}
	if val, ok := result.(int64); ok {
		return val, nil
	}
	return 0, fmt.Errorf("unexpected result type: %T", result)
}

func GetViewCount(ctx context.Context, r *redis.Redis, postID int64) (int64, error) {
	key := fmt.Sprintf(PostReadCountKey, postID)
	val, err := r.Get(key)
	if err != nil {
		return 0, err
	}
	var count int64
	fmt.Sscanf(val, "%d", &count)
	return count, nil
}

func SetViewCount(ctx context.Context, r *redis.Redis, postID int64, count int64) error {
	key := fmt.Sprintf(PostReadCountKey, postID)
	return r.Set(key, fmt.Sprintf("%d", count))
}

func SetViewCountWithExpire(ctx context.Context, r *redis.Redis, postID int64, count int64, expire time.Duration) error {
	key := fmt.Sprintf(PostReadCountKey, postID)
	return r.Setex(key, fmt.Sprintf("%d", count), int(expire/time.Second))
}
