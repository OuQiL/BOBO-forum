package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	ViewCountSyncInterval = 5 * time.Minute
	ViewCountKeyPattern   = "post:view:*"
)

type ViewCountSyncer struct {
	redis      *redis.Redis
	db         sqlx.SqlConn
	localCache *HotPostLocalCache
	stopChan   chan struct{}
}

func NewViewCountSyncer(rds *redis.Redis, db sqlx.SqlConn, localCache *HotPostLocalCache) *ViewCountSyncer {
	return &ViewCountSyncer{
		redis:      rds,
		db:         db,
		localCache: localCache,
		stopChan:   make(chan struct{}),
	}
}

func (s *ViewCountSyncer) Start() {
	go s.run()
	logx.Info("View count syncer started")
}

func (s *ViewCountSyncer) Stop() {
	close(s.stopChan)
	logx.Info("View count syncer stopped")
}

func (s *ViewCountSyncer) run() {
	ticker := time.NewTicker(ViewCountSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.Sync(); err != nil {
				logx.Errorf("Failed to sync view counts: %v", err)
			}
		}
	}
}

func (s *ViewCountSyncer) Sync() error {
	ctx := context.Background()

	keys, err := s.redis.Keys(ViewCountKeyPattern)
	if err != nil {
		return fmt.Errorf("failed to get view count keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	var syncErr error
	syncCount := 0

	for _, key := range keys {
		var postID int64
		if _, err := fmt.Sscanf(key, "post:view:%d", &postID); err != nil {
			continue
		}

		viewCount, err := s.redis.Pfcount(key)
		if err != nil {
			logx.Errorf("Failed to get view count for post %d: %v", postID, err)
			continue
		}

		if viewCount == 0 {
			continue
		}

		_, err = s.db.ExecCtx(ctx,
			"UPDATE posts SET view_count = view_count + ? WHERE id = ?",
			viewCount, postID)
		if err != nil {
			logx.Errorf("Failed to update view count for post %d: %v", postID, err)
			syncErr = err
			continue
		}

		s.localCache.UpdateViewCount(postID, viewCount)

		if _, err := s.redis.Del(key); err != nil {
			logx.Errorf("Failed to delete view count key for post %d: %v", postID, err)
		}

		syncCount++
	}

	if syncCount > 0 {
		logx.Infof("Synced view counts for %d posts", syncCount)
	}

	return syncErr
}

func (s *ViewCountSyncer) SyncSingle(ctx context.Context, postID int64) error {
	key := fmt.Sprintf("post:view:%d", postID)

	viewCount, err := s.redis.Pfcount(key)
	if err != nil {
		return fmt.Errorf("failed to get view count: %w", err)
	}

	if viewCount == 0 {
		return nil
	}

	_, err = s.db.ExecCtx(ctx,
		"UPDATE posts SET view_count = view_count + ? WHERE id = ?",
		viewCount, postID)
	if err != nil {
		return fmt.Errorf("failed to update view count: %w", err)
	}

	s.localCache.UpdateViewCount(postID, viewCount)

	if _, err := s.redis.Del(key); err != nil {
		return fmt.Errorf("failed to delete view count key: %w", err)
	}

	return nil
}
