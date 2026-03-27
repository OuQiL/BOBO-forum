package repository

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"post/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	CommentWeight = 3.0
	ViewWeight    = 0.1
	DecayFactor   = 0.02
)

const (
	HotPostsHour     = "hot_posts:hour"
	HotPostsDay      = "hot_posts:day"
	HotPostsWeek     = "hot_posts:week"
	HotPostsMonth    = "hot_posts:month"
	PostDetailPrefix = "post:detail"
)

type HotPostService struct {
	redis      *redis.Redis
	localCache *HotPostLocalCache
}

func NewHotPostService(rds *redis.Redis, localCache *HotPostLocalCache) *HotPostService {
	return &HotPostService{
		redis:      rds,
		localCache: localCache,
	}
}

func (s *HotPostService) CalculateScore(post *model.Post) float64 {
	baseScore := float64(post.CommentCount)*CommentWeight +
		float64(post.ViewCount)*ViewWeight

	hours := time.Since(time.Unix(post.CreatedAt, 0)).Hours()
	decayFactor := math.Exp(-DecayFactor * hours)

	return baseScore * decayFactor
}

func (s *HotPostService) UpdatePostScore(ctx context.Context, post *model.Post) error {
	score := s.CalculateScore(post)
	postIDStr := fmt.Sprintf("%d", post.ID)

	keys := []string{HotPostsHour, HotPostsDay, HotPostsWeek, HotPostsMonth}
	for _, key := range keys {
		_, err := s.redis.Zadd(key, int64(score*1000), postIDStr)
		if err != nil {
			return err
		}
	}

	scoreKey := fmt.Sprintf("post:hot:score:%d", post.ID)
	return s.redis.Setex(scoreKey, fmt.Sprintf("%.2f", score), 3600)
}

func (s *HotPostService) GetHotPosts(ctx context.Context, timeRange string, page, pageSize int64) ([]string, error) {
	key := s.getTimeRangeKey(timeRange)
	start := (page - 1) * pageSize
	end := start + pageSize - 1

	return s.redis.Zrevrange(key, start, end)
}

func (s *HotPostService) CachePostDetail(ctx context.Context, post *model.Post, ttl int) error {
	if s.localCache.ShouldCache(post.ViewCount) {
		s.localCache.Set(post)
	}

	key := fmt.Sprintf("%s:%d", PostDetailPrefix, post.ID)
	data := fmt.Sprintf("%d|%d|%d|%s|%s|%s|%s|%d|%d|%d|%d|%d",
		post.ID, post.UserID, post.CommunityID, post.Username, post.Title,
		post.Content, post.Tags, post.LikeCount, post.CommentCount,
		post.ViewCount, post.CreatedAt, post.UpdatedAt)

	return s.redis.Setex(key, data, ttl)
}

func (s *HotPostService) GetCachedPostDetail(ctx context.Context, postID int64) (*model.Post, error) {
	if post, found := s.localCache.Get(postID); found {
		return post, nil
	}

	key := fmt.Sprintf("%s:%d", PostDetailPrefix, postID)
	data, err := s.redis.Get(key)
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}

	parts := strings.Split(data, "|")
	if len(parts) != 12 {
		return nil, fmt.Errorf("invalid cached data")
	}

	post := &model.Post{
		ID:           parseInt64(parts[0]),
		UserID:       parseInt64(parts[1]),
		CommunityID:  parseInt64(parts[2]),
		Username:     parts[3],
		Title:        parts[4],
		Content:      parts[5],
		Tags:         parts[6],
		LikeCount:    parseInt64(parts[7]),
		CommentCount: parseInt64(parts[8]),
		ViewCount:    parseInt64(parts[9]),
		CreatedAt:    parseInt64(parts[10]),
		UpdatedAt:    parseInt64(parts[11]),
	}

	if s.localCache.ShouldCache(post.ViewCount) {
		s.localCache.Set(post)
	}

	return post, nil
}

func (s *HotPostService) RemovePostFromHot(ctx context.Context, postID int64) error {
	s.localCache.Delete(postID)

	postIDStr := fmt.Sprintf("%d", postID)

	keys := []string{HotPostsHour, HotPostsDay, HotPostsWeek, HotPostsMonth}
	for _, key := range keys {
		_, err := s.redis.Zrem(key, postIDStr)
		if err != nil {
			return err
		}
	}

	scoreKey := fmt.Sprintf("post:hot:score:%d", postID)
	_, err := s.redis.Del(scoreKey)
	return err
}

func (s *HotPostService) CleanupExpiredPosts(ctx context.Context, timeRange string) error {
	key := s.getTimeRangeKey(timeRange)

	var retentionHours int
	switch timeRange {
	case "hour":
		retentionHours = 24
	case "day":
		retentionHours = 72
	case "week":
		retentionHours = 168
	case "month":
		retentionHours = 720
	default:
		retentionHours = 168
	}

	cutoffTime := time.Now().Add(-time.Duration(retentionHours) * time.Hour).Unix()

	_, err := s.redis.Zremrangebyscore(key, 0, cutoffTime)
	return err
}

func (s *HotPostService) getTimeRangeKey(timeRange string) string {
	switch timeRange {
	case "hour":
		return HotPostsHour
	case "day":
		return HotPostsDay
	case "week":
		return HotPostsWeek
	case "month":
		return HotPostsMonth
	default:
		return HotPostsWeek
	}
}

func parseInt64(s string) int64 {
	var id int64
	fmt.Sscanf(s, "%d", &id)
	return id
}
