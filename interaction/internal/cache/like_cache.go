package cache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	LikeCountPrefix   = "like:count:"
	LikeRelationPrefix = "like:relation:"
	LikeUserPrefix    = "like:user:"
	LikePendingSync   = "like:pending:sync"
)

type LikeCache struct {
	redis *redis.Redis
}

func NewLikeCache(redis *redis.Redis) *LikeCache {
	return &LikeCache{
		redis: redis,
	}
}

func (c *LikeCache) buildCountKey(targetType int32, targetId int64) string {
	return fmt.Sprintf("%s%d:%d", LikeCountPrefix, targetType, targetId)
}

func (c *LikeCache) buildRelationKey(targetType int32, targetId int64) string {
	return fmt.Sprintf("%s%d:%d", LikeRelationPrefix, targetType, targetId)
}

func (c *LikeCache) buildUserKey(userId int64) string {
	return fmt.Sprintf("%s%d", LikeUserPrefix, userId)
}

func (c *LikeCache) GetLikeCount(ctx context.Context, targetType int32, targetId int64) (int64, error) {
	key := c.buildCountKey(targetType, targetId)
	val, err := c.redis.Get(key)
	if err != nil {
		if err == redis.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	
	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

func (c *LikeCache) IncrLikeCount(ctx context.Context, targetType int32, targetId int64) (int64, error) {
	key := c.buildCountKey(targetType, targetId)
	count, err := c.redis.Incr(key)
	if err != nil {
		return 0, err
	}
	
	c.redis.Setex(key, strconv.FormatInt(count, 10), 86400)
	
	return count, nil
}

func (c *LikeCache) DecrLikeCount(ctx context.Context, targetType int32, targetId int64) (int64, error) {
	key := c.buildCountKey(targetType, targetId)
	count, err := c.redis.Decr(key)
	if err != nil {
		return 0, err
	}
	
	if count < 0 {
		count = 0
		c.redis.Setex(key, "0", 86400)
	}
	
	return count, nil
}

func (c *LikeCache) SetLikeCount(ctx context.Context, targetType int32, targetId int64, count int64) error {
	key := c.buildCountKey(targetType, targetId)
	return c.redis.Setex(key, strconv.FormatInt(count, 10), 86400)
}

func (c *LikeCache) IsLiked(ctx context.Context, targetType int32, targetId, userId int64) (bool, error) {
	key := c.buildRelationKey(targetType, targetId)
	return c.redis.Sismember(key, strconv.FormatInt(userId, 10))
}

func (c *LikeCache) AddLikeRelation(ctx context.Context, targetType int32, targetId, userId int64) error {
	relationKey := c.buildRelationKey(targetType, targetId)
	_, err := c.redis.Sadd(relationKey, strconv.FormatInt(userId, 10))
	if err != nil {
		return err
	}
	
	userKey := c.buildUserKey(userId)
	_, err = c.redis.Sadd(userKey, fmt.Sprintf("%d:%d", targetType, targetId))
	if err != nil {
		return err
	}
	
	c.redis.Setex(relationKey, c.redis.Get(key), 86400)
	c.redis.Setex(userKey, c.redis.Get(key), 86400)
	
	return nil
}

func (c *LikeCache) RemoveLikeRelation(ctx context.Context, targetType int32, targetId, userId int64) error {
	relationKey := c.buildRelationKey(targetType, targetId)
	_, err := c.redis.Srem(relationKey, strconv.FormatInt(userId, 10))
	if err != nil {
		return err
	}
	
	userKey := c.buildUserKey(userId)
	_, err = c.redis.Srem(userKey, fmt.Sprintf("%d:%d", targetType, targetId))
	if err != nil {
		return err
	}
	
	return nil
}

func (c *LikeCache) GetLikeRelationCount(ctx context.Context, targetType int32, targetId int64) (int64, error) {
	key := c.buildRelationKey(targetType, targetId)
	count, err := c.redis.Scard(key)
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

func (c *LikeCache) MarkPendingSync(ctx context.Context, targetType int32, targetId int64) error {
	_, err := c.redis.Sadd(LikePendingSync, fmt.Sprintf("%d:%d", targetType, targetId))
	return err
}

func (c *LikeCache) GetPendingSyncTargets(ctx context.Context, limit int64) ([]string, error) {
	members, err := c.redis.Smembers(LikePendingSync)
	if err != nil {
		return nil, err
	}
	
	if int64(len(members)) > limit {
		return members[:limit], nil
	}
	
	return members, nil
}

func (c *LikeCache) RemovePendingSync(ctx context.Context, targetType int32, targetId int64) error {
	_, err := c.redis.Srem(LikePendingSync, fmt.Sprintf("%d:%d", targetType, targetId))
	return err
}

func (c *LikeCache) BatchGetLikeStatus(ctx context.Context, targetType int32, userId int64, targetIds []int64) (map[int64]bool, error) {
	result := make(map[int64]bool)
	
	for _, targetId := range targetIds {
		key := c.buildRelationKey(targetType, targetId)
		isMember, err := c.redis.Sismember(key, strconv.FormatInt(userId, 10))
		if err != nil {
			continue
		}
		result[targetId] = isMember
	}
	
	return result, nil
}
