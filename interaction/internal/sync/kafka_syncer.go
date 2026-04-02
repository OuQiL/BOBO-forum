package sync

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"interaction/internal/cache"
	"interaction/internal/kafka"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type KafkaLikeCountSyncer struct {
	redis      *redis.Redis
	db         sqlx.SqlConn
	likeCache  *cache.LikeCache
	batchSize  int
	processing sync.Map
}

func NewKafkaLikeCountSyncer(redis *redis.Redis, db sqlx.SqlConn, likeCache *cache.LikeCache, batchSize int) *KafkaLikeCountSyncer {
	return &KafkaLikeCountSyncer{
		redis:     redis,
		db:        db,
		likeCache: likeCache,
		batchSize: batchSize,
	}
}

func (s *KafkaLikeCountSyncer) ProcessBatch(ctx context.Context, messages []*kafka.LikeMessage) error {
	logx.Infof("Processing batch of %d like messages", len(messages))
	
	targetSet := make(map[string]bool)
	var targets []string
	
	for _, msg := range messages {
		key := fmt.Sprintf("%d:%d", msg.TargetType, msg.TargetId)
		if !targetSet[key] {
			targetSet[key] = true
			targets = append(targets, key)
		}
	}
	
	logx.Infof("Unique targets to sync: %d", len(targets))
	
	for _, target := range targets {
		parts := strings.Split(target, ":")
		if len(parts) != 2 {
			logx.Errorf("Invalid target format: %s", target)
			continue
		}
		
		targetType, err := strconv.Atoi(parts[0])
		if err != nil {
			logx.Errorf("Invalid target type: %s", parts[0])
			continue
		}
		
		targetId, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			logx.Errorf("Invalid target id: %s", parts[1])
			continue
		}
		
		if err := s.syncSingleTarget(ctx, int32(targetType), targetId); err != nil {
			logx.Errorf("Failed to sync target %s: %v", target, err)
			continue
		}
	}
	
	logx.Infof("Completed syncing batch of %d messages", len(messages))
	return nil
}

func (s *KafkaLikeCountSyncer) syncSingleTarget(ctx context.Context, targetType int32, targetId int64) error {
	redisCount, err := s.likeCache.GetLikeCount(ctx, targetType, targetId)
	if err != nil {
		return fmt.Errorf("failed to get redis count: %w", err)
	}
	
	relationCount, err := s.likeCache.GetLikeRelationCount(ctx, targetType, targetId)
	if err != nil {
		return fmt.Errorf("failed to get relation count: %w", err)
	}
	
	if redisCount != relationCount {
		logx.Warnf("Count mismatch for target %d:%d - Redis: %d, Relation: %d", 
			targetType, targetId, redisCount, relationCount)
		redisCount = relationCount
	}
	
	var tableName string
	if targetType == 1 {
		tableName = "posts"
	} else if targetType == 2 {
		tableName = "comments"
	} else {
		return fmt.Errorf("unknown target type: %d", targetType)
	}
	
	query := fmt.Sprintf("UPDATE %s SET like_count = ? WHERE id = ?", tableName)
	_, err = s.db.ExecCtx(ctx, query, redisCount, targetId)
	if err != nil {
		return fmt.Errorf("failed to update mysql: %w", err)
	}
	
	logx.Debugf("Synced like count for %s %d: %d", tableName, targetId, redisCount)
	return nil
}

type LikeDataConsistencyChecker struct {
	redis     *redis.Redis
	db        sqlx.SqlConn
	likeCache *cache.LikeCache
}

func NewLikeDataConsistencyChecker(redis *redis.Redis, db sqlx.SqlConn, likeCache *cache.LikeCache) *LikeDataConsistencyChecker {
	return &LikeDataConsistencyChecker{
		redis:     redis,
		db:        db,
		likeCache: likeCache,
	}
}

func (c *LikeDataConsistencyChecker) CheckAndFix(ctx context.Context, targetType int32, targetId int64) error {
	redisCount, err := c.likeCache.GetLikeCount(ctx, targetType, targetId)
	if err != nil {
		return fmt.Errorf("failed to get redis count: %w", err)
	}
	
	relationCount, err := c.likeCache.GetLikeRelationCount(ctx, targetType, targetId)
	if err != nil {
		return fmt.Errorf("failed to get relation count: %w", err)
	}
	
	if redisCount != relationCount {
		logx.Warnf("Redis count (%d) != Relation count (%d) for target %d:%d, fixing...", 
			targetType, targetId, redisCount, relationCount)
		
		if err := c.likeCache.SetLikeCount(ctx, targetType, targetId, relationCount); err != nil {
			return fmt.Errorf("failed to fix redis count: %w", err)
		}
		
		logx.Infof("Fixed like count for target %d:%d to %d", targetType, targetId, relationCount)
	}
	
	var tableName string
	if targetType == 1 {
		tableName = "posts"
	} else if targetType == 2 {
		tableName = "comments"
	} else {
		return fmt.Errorf("unknown target type: %d", targetType)
	}
	
	var mysqlCount int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM likes WHERE target_id = ? AND type = ? AND status = 1")
	err = c.db.QueryRowCtx(ctx, &mysqlCount, query, targetId, targetType)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query mysql count: %w", err)
	}
	
	if mysqlCount != relationCount {
		logx.Warnf("MySQL count (%d) != Relation count (%d) for target %d:%d", 
			targetType, targetId, mysqlCount, relationCount)
		
		updateQuery := fmt.Sprintf("UPDATE %s SET like_count = ? WHERE id = ?", tableName)
		_, err = c.db.ExecCtx(ctx, updateQuery, relationCount, targetId)
		if err != nil {
			return fmt.Errorf("failed to update mysql count: %w", err)
		}
		
		logx.Infof("Fixed MySQL like count for %s %d to %d", tableName, targetId, relationCount)
	}
	
	return nil
}
