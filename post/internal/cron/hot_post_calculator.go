package cron

import (
	"context"
	"fmt"

	"post/internal/repository"

	"github.com/robfig/cron/v3"
)

type CronScheduler struct {
	postRepo   repository.PostRepository
	hotPostSvc *repository.HotPostService
	cron       *cron.Cron
	stopChan   chan struct{}
}

func NewCronScheduler(postRepo repository.PostRepository, hotPostSvc *repository.HotPostService) *CronScheduler {
	return &CronScheduler{
		postRepo:   postRepo,
		hotPostSvc: hotPostSvc,
		cron:       cron.New(),
		stopChan:   make(chan struct{}),
	}
}

func (s *CronScheduler) Start() {
	ctx := context.Background()
	s.cron.AddFunc("0 * * * *", func() {
		s.Run(ctx)
	})
	s.cron.Start()
	fmt.Println("[Cron] Hot post scheduler started: runs every hour at minute 0")
}

func (s *CronScheduler) Stop() {
	s.cron.Stop()
	close(s.stopChan)
	fmt.Println("[Cron] Hot post scheduler stopped")
}

func (s *CronScheduler) Run(ctx context.Context) {
	fmt.Println("[Cron] Starting hot post calculation job...")

	posts, err := s.postRepo.GetRecentPosts(ctx, 7)
	if err != nil {
		fmt.Printf("[Cron] Failed to get recent posts: %v\n", err)
		return
	}

	fmt.Printf("[Cron] Processing %d posts...\n", len(posts))

	successCount := 0
	for _, post := range posts {
		if err := s.hotPostSvc.UpdatePostScore(ctx, post); err != nil {
			fmt.Printf("[Cron] Failed to update post score: postID=%d, err=%v\n", post.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("[Cron] Updated scores for %d/%d posts\n", successCount, len(posts))

	cachedCount := 0
	for _, post := range posts {
		if err := s.hotPostSvc.CachePostDetail(ctx, post, 3600); err != nil {
			fmt.Printf("[Cron] Failed to cache post detail: postID=%d, err=%v\n", post.ID, err)
			continue
		}
		cachedCount++
	}

	fmt.Printf("[Cron] Cached details for %d/%d posts\n", cachedCount, len(posts))

	fmt.Println("[Cron] Starting cleanup expired posts...")
	timeRanges := []string{"hour", "day", "week", "month"}
	for _, timeRange := range timeRanges {
		if err := s.hotPostSvc.CleanupExpiredPosts(ctx, timeRange); err != nil {
			fmt.Printf("[Cron] Failed to cleanup expired posts for %s: %v\n", timeRange, err)
		} else {
			fmt.Printf("[Cron] Cleaned up expired posts for %s\n", timeRange)
		}
	}

	fmt.Println("[Cron] Hot posts calculation completed successfully")
}
