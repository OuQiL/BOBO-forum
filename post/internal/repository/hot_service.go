package repository

import (
	"context"
	"errors"
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

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

type HotPostService struct {
	redis           *redis.Redis
	localCache      *HotPostLocalCache
	emptyValueCache *EmptyValueCache
	circuitBreaker  *CircuitBreaker
}

type HotPostServiceConfig struct {
	Redis              *redis.Redis
	LocalCache         *HotPostLocalCache
	EmptyCacheTTL      int
	CircuitBreakerConf CircuitBreakerConfig
}

func NewHotPostService(rds *redis.Redis, localCache *HotPostLocalCache) *HotPostService {
	return &HotPostService{
		redis:           rds,
		localCache:      localCache,
		emptyValueCache: NewEmptyValueCache(rds, 60),
		circuitBreaker:  NewCircuitBreaker(CircuitBreakerConfig{}),
	}
}

func NewHotPostServiceWithConfig(cfg HotPostServiceConfig) *HotPostService {
	emptyCacheTTL := cfg.EmptyCacheTTL
	if emptyCacheTTL <= 0 {
		emptyCacheTTL = 60
	}

	return &HotPostService{
		redis:           cfg.Redis,
		localCache:      cfg.LocalCache,
		emptyValueCache: NewEmptyValueCache(cfg.Redis, emptyCacheTTL),
		circuitBreaker:  NewCircuitBreaker(cfg.CircuitBreakerConf),
	}
}

func (s *HotPostService) GetCircuitBreaker() *CircuitBreaker {
	return s.circuitBreaker
}

func (s *HotPostService) GetEmptyValueCache() *EmptyValueCache {
	return s.emptyValueCache
}

func (s *HotPostService) ShouldCachePost(viewCount int64) bool {
	return s.localCache.ShouldCache(viewCount)
}

func (s *HotPostService) CachePost(post *model.Post) {
	if s.localCache.ShouldCache(post.ViewCount) {
		s.localCache.Set(post)
	}
}

func (s *HotPostService) GetLocalCacheSize() int {
	return s.localCache.Size()
}

func (s *HotPostService) CalculateScore(post *model.Post) float64 {
	baseScore := float64(post.CommentCount)*CommentWeight +
		float64(post.ViewCount)*ViewWeight

	hours := time.Since(time.Unix(post.CreatedAt, 0)).Hours()
	decayFactor := math.Exp(-DecayFactor * hours)

	return baseScore * decayFactor
}

func (s *HotPostService) UpdatePostScore(ctx context.Context, post *model.Post) error {
	if !s.circuitBreaker.Allow() {
		return ErrCircuitBreakerOpen
	}

	err := s.updatePostScoreInternal(ctx, post)
	if err != nil {
		s.circuitBreaker.RecordFailure()
		return err
	}

	s.circuitBreaker.RecordSuccess()
	return nil
}

func (s *HotPostService) updatePostScoreInternal(ctx context.Context, post *model.Post) error {
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
	if !s.circuitBreaker.Allow() {
		return nil, ErrCircuitBreakerOpen
	}

	key := s.getTimeRangeKey(timeRange)
	start := (page - 1) * pageSize
	end := start + pageSize - 1

	result, err := s.redis.Zrevrange(key, start, end)
	if err != nil {
		s.circuitBreaker.RecordFailure()
		return nil, err
	}

	s.circuitBreaker.RecordSuccess()
	return result, nil
}

func (s *HotPostService) CachePostDetail(ctx context.Context, post *model.Post, ttl int) error {
	if s.localCache.ShouldCache(post.ViewCount) {
		s.localCache.Set(post)
	}

	s.emptyValueCache.RemoveEmptyCache(ctx, post.ID)

	if s.redis == nil {
		return nil
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

	if s.emptyValueCache.IsEmptyCached(ctx, postID) {
		return nil, ErrPostNotFound
	}

	if s.redis == nil {
		return nil, nil
	}

	if !s.circuitBreaker.Allow() {
		return nil, ErrCircuitBreakerOpen
	}

	key := fmt.Sprintf("%s:%d", PostDetailPrefix, postID)
	data, err := s.redis.Get(key)
	if err != nil {
		s.circuitBreaker.RecordFailure()
		return nil, err
	}

	if data == "" {
		s.circuitBreaker.RecordSuccess()
		return nil, nil
	}

	parts := strings.Split(data, "|")
	if len(parts) != 12 {
		s.circuitBreaker.RecordSuccess()
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

	s.circuitBreaker.RecordSuccess()
	return post, nil
}

func (s *HotPostService) GetPostDetailWithFallback(ctx context.Context, postID int64, fetchFn func(ctx context.Context, id int64) (*model.Post, error)) (*model.Post, error) {
	cached, err := s.GetCachedPostDetail(ctx, postID)
	if err != nil && !errors.Is(err, ErrPostNotFound) && !errors.Is(err, ErrCircuitBreakerOpen) {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	if s.emptyValueCache.IsEmptyCached(ctx, postID) {
		return nil, ErrPostNotFound
	}

	post, err := fetchFn(ctx, postID)
	if err != nil {
		return nil, err
	}

	if post == nil {
		if cacheErr := s.emptyValueCache.CacheEmpty(ctx, postID); cacheErr != nil {
			return nil, ErrPostNotFound
		}
		return nil, ErrPostNotFound
	}

	if cacheErr := s.CachePostDetail(ctx, post, 300); cacheErr != nil {
	}

	return post, nil
}

func (s *HotPostService) RemovePostFromHot(ctx context.Context, postID int64) error {
	s.localCache.Delete(postID)

	if s.redis == nil {
		return nil
	}

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

func (s *HotPostService) DeletePostCache(ctx context.Context, postID int64) error {
	s.localCache.Delete(postID)
	s.emptyValueCache.RemoveEmptyCache(ctx, postID)

	if s.redis == nil {
		return nil
	}

	key := fmt.Sprintf("%s:%d", PostDetailPrefix, postID)
	_, err := s.redis.Del(key)
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
