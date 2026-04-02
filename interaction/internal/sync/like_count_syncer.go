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

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type LikeCountSyncer struct {
	redis     *redis.Redis
	db        sqlx.SqlConn
	likeCache *cache.LikeCache
	interval  time.Duration
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

func NewLikeCountSyncer(redis *redis.Redis, db sqlx.SqlConn, likeCache *cache.LikeCache, intervalSeconds int) *LikeCountSyncer {
	return &LikeCountSyncer{
		redis:     redis,
		db:        db,
		likeCache: likeCache,
		interval:  time.Duration(intervalSeconds) * time.Second,
		stopChan:  make(chan struct{}),
	}
}

func (s *LikeCountSyncer) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.syncToMySQL()
			case <-s.stopChan:
				logx.Info("LikeCountSyncer stopping...")
				return
			}
		}
	}()

	logx.Infof("LikeCountSyncer started with interval %v", s.interval)
}

func (s *LikeCountSyncer) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	logx.Info("LikeCountSyncer stopped")
}

func (s *LikeCountSyncer) syncToMySQL() {
	ctx := context.Background()

	targets, err := s.likeCache.GetPendingSyncTargets(ctx, 1000)
	if err != nil {
		logx.Errorf("Failed to get pending sync targets: %v", err)
		return
	}

	if len(targets) == 0 {
		return
	}

	logx.Infof("Syncing %d like count targets to MySQL", len(targets))

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

		if err := s.likeCache.RemovePendingSync(ctx, int32(targetType), targetId); err != nil {
			logx.Errorf("Failed to remove pending sync for %s: %v", target, err)
		}
	}

	logx.Infof("Completed syncing %d targets", len(targets))
}

func (s *LikeCountSyncer) syncSingleTarget(ctx context.Context, targetType int32, targetId int64) error {
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

func (s *LikeCountSyncer) ForceSync(ctx context.Context, targetType int32, targetId int64) error {
	redisCount, err := s.likeCache.GetLikeCount(ctx, targetType, targetId)
	if err != nil {
		return fmt.Errorf("failed to get redis count: %w", err)
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

	logx.Infof("Force synced like count for %s %d: %d", tableName, targetId, redisCount)
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
	query := fmt.Sprintf("SELECT COUNT(*) FROM likes WHERE target_id = ? AND type = ? AND status = 1", tableName)
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
