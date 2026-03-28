package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"post/internal/model"

	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func newTestHotPostService() *HotPostService {
	localCache := NewHotPostLocalCache(1000, 5*time.Minute, 1000)
	return NewHotPostService(nil, localCache)
}

func newTestHotPostServiceWithRedis() (*HotPostService, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	r := redis.New(mr.Addr())
	localCache := NewHotPostLocalCache(1000, 5*time.Minute, 100)
	return NewHotPostService(r, localCache), mr
}

func TestCalculateScore(t *testing.T) {
	service := newTestHotPostService()

	tests := []struct {
		name        string
		post        *model.Post
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "New post with high engagement",
			post: &model.Post{
				ID:           1,
				CommentCount: 10,
				ViewCount:    1000,
				CreatedAt:    time.Now().Unix(),
			},
			expectedMin: 100.0,
			expectedMax: 150.0,
		},
		{
			name: "Old post with high engagement",
			post: &model.Post{
				ID:           2,
				CommentCount: 20,
				ViewCount:    2000,
				CreatedAt:    time.Now().Add(-24 * time.Hour).Unix(),
			},
			expectedMin: 150.0,
			expectedMax: 170.0,
		},
		{
			name: "New post with low engagement",
			post: &model.Post{
				ID:           3,
				CommentCount: 1,
				ViewCount:    10,
				CreatedAt:    time.Now().Unix(),
			},
			expectedMin: 0.0,
			expectedMax: 10.0,
		},
		{
			name: "Very old post",
			post: &model.Post{
				ID:           4,
				CommentCount: 100,
				ViewCount:    10000,
				CreatedAt:    time.Now().Add(-168 * time.Hour).Unix(),
			},
			expectedMin: 0.0,
			expectedMax: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.CalculateScore(tt.post)
			t.Logf("Post %s: score=%.2f (expected range: %.2f-%.2f)",
				tt.name, score, tt.expectedMin, tt.expectedMax)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("Score %.2f is out of expected range [%.2f, %.2f]",
					score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestCalculateScoreDecay(t *testing.T) {
	service := newTestHotPostService()

	basePost := &model.Post{
		ID:           1,
		CommentCount: 10,
		ViewCount:    1000,
	}

	now := time.Now().Unix()
	hourAgo := time.Now().Add(-1 * time.Hour).Unix()
	dayAgo := time.Now().Add(-24 * time.Hour).Unix()
	weekAgo := time.Now().Add(-168 * time.Hour).Unix()

	scoreNow := service.CalculateScore(&model.Post{
		CommentCount: basePost.CommentCount,
		ViewCount:    basePost.ViewCount,
		CreatedAt:    now,
	})

	scoreHourAgo := service.CalculateScore(&model.Post{
		CommentCount: basePost.CommentCount,
		ViewCount:    basePost.ViewCount,
		CreatedAt:    hourAgo,
	})

	scoreDayAgo := service.CalculateScore(&model.Post{
		CommentCount: basePost.CommentCount,
		ViewCount:    basePost.ViewCount,
		CreatedAt:    dayAgo,
	})

	scoreWeekAgo := service.CalculateScore(&model.Post{
		CommentCount: basePost.CommentCount,
		ViewCount:    basePost.ViewCount,
		CreatedAt:    weekAgo,
	})

	t.Logf("Score now: %.2f", scoreNow)
	t.Logf("Score 1 hour ago: %.2f", scoreHourAgo)
	t.Logf("Score 1 day ago: %.2f", scoreDayAgo)
	t.Logf("Score 1 week ago: %.2f", scoreWeekAgo)

	if scoreNow <= scoreHourAgo {
		t.Error("Recent post should have higher score than older post")
	}

	if scoreHourAgo <= scoreDayAgo {
		t.Error("1-hour-old post should have higher score than 1-day-old post")
	}

	if scoreDayAgo <= scoreWeekAgo {
		t.Error("1-day-old post should have higher score than 1-week-old post")
	}
}

func TestGetTimeRangeKey(t *testing.T) {
	service := newTestHotPostService()

	tests := []struct {
		timeRange string
		expected  string
	}{
		{"hour", HotPostsHour},
		{"day", HotPostsDay},
		{"week", HotPostsWeek},
		{"month", HotPostsMonth},
		{"unknown", HotPostsWeek},
	}

	for _, tt := range tests {
		t.Run(tt.timeRange, func(t *testing.T) {
			result := service.getTimeRangeKey(tt.timeRange)
			if result != tt.expected {
				t.Errorf("getTimeRangeKey(%s) = %s, want %s",
					tt.timeRange, result, tt.expected)
			}
		})
	}
}

func TestHotPostLocalCache(t *testing.T) {
	cache := NewHotPostLocalCache(100, 5*time.Minute, 100)

	post := &model.Post{
		ID:           1,
		UserID:       100,
		Title:        "Test Post",
		Content:      "Test Content",
		ViewCount:    150,
		CommentCount: 10,
	}

	cache.Set(post)

	cached, found := cache.Get(1)
	if !found {
		t.Fatal("Expected to find cached post")
	}

	if cached.ID != post.ID {
		t.Errorf("Expected ID %d, got %d", post.ID, cached.ID)
	}

	if cached.Title != post.Title {
		t.Errorf("Expected Title %s, got %s", post.Title, cached.Title)
	}

	cache.Delete(1)

	_, found = cache.Get(1)
	if found {
		t.Error("Expected post to be deleted from cache")
	}
}

func TestHotPostLocalCacheThreshold(t *testing.T) {
	cache := NewHotPostLocalCache(100, 5*time.Minute, 100)

	lowViewPost := &model.Post{
		ID:        1,
		ViewCount: 50,
	}

	cache.Set(lowViewPost)

	_, found := cache.Get(1)
	if found {
		t.Error("Low view count post should not be cached")
	}

	highViewPost := &model.Post{
		ID:        2,
		ViewCount: 150,
	}

	cache.Set(highViewPost)

	_, found = cache.Get(2)
	if !found {
		t.Error("High view count post should be cached")
	}
}

func TestHotPostLocalCacheShouldCache(t *testing.T) {
	cache := NewHotPostLocalCache(100, 5*time.Minute, 1000)

	if cache.ShouldCache(500) {
		t.Error("ViewCount 500 should not be cached with threshold 1000")
	}

	if !cache.ShouldCache(1000) {
		t.Error("ViewCount 1000 should be cached with threshold 1000")
	}

	if !cache.ShouldCache(1500) {
		t.Error("ViewCount 1500 should be cached with threshold 1000")
	}
}

func TestDeletePostCache(t *testing.T) {
	cache := NewHotPostLocalCache(100, 5*time.Minute, 100)
	service := NewHotPostService(nil, cache)

	post := &model.Post{
		ID:           1,
		UserID:       100,
		Title:        "Test Post",
		Content:      "Test Content",
		ViewCount:    150,
		CommentCount: 10,
	}

	cache.Set(post)

	_, found := cache.Get(1)
	if !found {
		t.Fatal("Expected to find cached post before delete")
	}

	service.localCache.Delete(1)

	_, found = cache.Get(1)
	if found {
		t.Error("Expected post to be deleted from local cache")
	}
}

func TestCachePostDetail(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	post := &model.Post{
		ID:           1,
		UserID:       100,
		CommunityID:  1,
		Username:     "testuser",
		Title:        "Test Post",
		Content:      "Test Content",
		Tags:         `["tag1","tag2"]`,
		LikeCount:    10,
		CommentCount: 5,
		ViewCount:    150,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	err := service.CachePostDetail(context.Background(), post, 300)
	if err != nil {
		t.Errorf("CachePostDetail returned error: %v", err)
	}

	cached, found := service.localCache.Get(1)
	if !found {
		t.Error("Expected post to be in local cache")
	} else {
		if cached.ID != post.ID {
			t.Errorf("Expected ID %d, got %d", post.ID, cached.ID)
		}
	}
}

func TestGetCachedPostDetail_LocalCache(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	post := &model.Post{
		ID:           1,
		UserID:       100,
		CommunityID:  1,
		Username:     "testuser",
		Title:        "Test Post",
		Content:      "Test Content",
		Tags:         `["tag1"]`,
		LikeCount:    10,
		CommentCount: 5,
		ViewCount:    150,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	service.localCache.Set(post)

	cached, err := service.GetCachedPostDetail(context.Background(), 1)
	if err != nil {
		t.Errorf("GetCachedPostDetail returned error: %v", err)
	}

	if cached == nil {
		t.Fatal("Expected to get cached post")
	}

	if cached.ID != post.ID {
		t.Errorf("Expected ID %d, got %d", post.ID, cached.ID)
	}
}

func TestUpdatePostScore(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	post := &model.Post{
		ID:           1,
		CommentCount: 10,
		ViewCount:    1000,
		CreatedAt:    time.Now().Unix(),
	}

	err := service.UpdatePostScore(context.Background(), post)
	if err != nil {
		t.Errorf("UpdatePostScore returned error: %v", err)
	}

	t.Log("UpdatePostScore completed successfully")
}

func TestRemovePostFromHot(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	post := &model.Post{
		ID:           1,
		UserID:       100,
		Title:        "Test Post",
		Content:      "Test Content",
		ViewCount:    150,
		CommentCount: 10,
	}

	service.localCache.Set(post)

	_, found := service.localCache.Get(1)
	if !found {
		t.Fatal("Expected to find cached post before remove")
	}

	err := service.RemovePostFromHot(context.Background(), 1)
	if err != nil {
		t.Errorf("RemovePostFromHot returned error: %v", err)
	}

	_, found = service.localCache.Get(1)
	if found {
		t.Error("Expected post to be removed from local cache")
	}
}

func TestGetPostDetailWithFallback_CacheHit(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	post := &model.Post{
		ID:           1,
		UserID:       100,
		CommunityID:  1,
		Username:     "testuser",
		Title:        "Test Post",
		Content:      "Test Content",
		ViewCount:    150,
		CommentCount: 10,
		CreatedAt:    time.Now().Unix(),
	}

	service.localCache.Set(post)

	fetchCalled := false
	fetchFn := func(ctx context.Context, id int64) (*model.Post, error) {
		fetchCalled = true
		return nil, nil
	}

	result, err := service.GetPostDetailWithFallback(context.Background(), 1, fetchFn)
	if err != nil {
		t.Errorf("GetPostDetailWithFallback returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected to get cached post")
	}

	if result.ID != post.ID {
		t.Errorf("Expected ID %d, got %d", post.ID, result.ID)
	}

	if fetchCalled {
		t.Error("Fetch function should not be called when cache hit")
	}
}

func TestGetPostDetailWithFallback_FetchFromDB(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	post := &model.Post{
		ID:           1,
		UserID:       100,
		CommunityID:  1,
		Username:     "testuser",
		Title:        "Test Post",
		Content:      "Test Content",
		ViewCount:    150,
		CommentCount: 10,
		CreatedAt:    time.Now().Unix(),
	}

	fetchFn := func(ctx context.Context, id int64) (*model.Post, error) {
		return post, nil
	}

	result, err := service.GetPostDetailWithFallback(context.Background(), 1, fetchFn)
	if err != nil {
		t.Errorf("GetPostDetailWithFallback returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected to get post from fetch function")
	}

	if result.ID != post.ID {
		t.Errorf("Expected ID %d, got %d", post.ID, result.ID)
	}

	cached, found := service.localCache.Get(1)
	if !found {
		t.Error("Expected post to be cached after fetch")
	} else if cached.ID != post.ID {
		t.Errorf("Cached post ID mismatch: expected %d, got %d", post.ID, cached.ID)
	}
}

func TestGetPostDetailWithFallback_PostNotFound(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	fetchFn := func(ctx context.Context, id int64) (*model.Post, error) {
		return nil, nil
	}

	result, err := service.GetPostDetailWithFallback(context.Background(), 999, fetchFn)
	if err == nil {
		t.Error("Expected error for post not found")
	}

	if !errors.Is(err, ErrPostNotFound) {
		t.Errorf("Expected ErrPostNotFound, got %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for post not found")
	}

	if !service.emptyValueCache.IsEmptyCached(context.Background(), 999) {
		t.Error("Expected post 999 to be cached as empty")
	}
}

func TestGetPostDetailWithFallback_EmptyCacheHit(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	service.emptyValueCache.CacheEmpty(context.Background(), 999)

	fetchCalled := false
	fetchFn := func(ctx context.Context, id int64) (*model.Post, error) {
		fetchCalled = true
		return nil, nil
	}

	result, err := service.GetPostDetailWithFallback(context.Background(), 999, fetchFn)
	if err == nil {
		t.Error("Expected error for empty cache hit")
	}

	if !errors.Is(err, ErrPostNotFound) {
		t.Errorf("Expected ErrPostNotFound, got %v", err)
	}

	if fetchCalled {
		t.Error("Fetch function should not be called when empty cache hit")
	}

	_ = result
}

func TestGetCircuitBreaker(t *testing.T) {
	service := newTestHotPostService()

	cb := service.GetCircuitBreaker()
	if cb == nil {
		t.Fatal("Expected circuit breaker to be returned")
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected circuit breaker to be closed, got %s", cb.State())
	}
}

func TestGetEmptyValueCache(t *testing.T) {
	service := newTestHotPostService()

	evc := service.GetEmptyValueCache()
	if evc == nil {
		t.Fatal("Expected empty value cache to be returned")
	}
}

func TestNewHotPostServiceWithConfig(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	localCache := NewHotPostLocalCache(1000, 5*time.Minute, 100)

	cfg := HotPostServiceConfig{
		Redis:         r,
		LocalCache:    localCache,
		EmptyCacheTTL: 120,
		CircuitBreakerConf: CircuitBreakerConfig{
			FailureThreshold: 10,
			Timeout:          60 * time.Second,
		},
	}

	service := NewHotPostServiceWithConfig(cfg)
	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.GetCircuitBreaker() == nil {
		t.Error("Expected circuit breaker to be initialized")
	}

	if service.GetEmptyValueCache() == nil {
		t.Error("Expected empty value cache to be initialized")
	}
}

func TestCircuitBreakerIntegration_OpenCircuit(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	for i := 0; i < 5; i++ {
		service.circuitBreaker.RecordFailure()
	}

	if service.circuitBreaker.State() != StateOpen {
		t.Errorf("Expected circuit breaker to be open, got %s", service.circuitBreaker.State())
	}

	post := &model.Post{
		ID:           1,
		CommentCount: 10,
		ViewCount:    1000,
		CreatedAt:    time.Now().Unix(),
	}

	err := service.UpdatePostScore(context.Background(), post)
	if !errors.Is(err, ErrCircuitBreakerOpen) {
		t.Errorf("Expected ErrCircuitBreakerOpen, got %v", err)
	}
}

func TestCircuitBreakerIntegration_HalfOpenToClosed(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	cbConfig := CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
	}
	service.circuitBreaker = NewCircuitBreaker(cbConfig)

	service.circuitBreaker.RecordFailure()
	service.circuitBreaker.RecordFailure()

	if service.circuitBreaker.State() != StateOpen {
		t.Fatalf("Expected circuit breaker to be open, got %s", service.circuitBreaker.State())
	}

	time.Sleep(150 * time.Millisecond)

	post := &model.Post{
		ID:           1,
		CommentCount: 10,
		ViewCount:    1000,
		CreatedAt:    time.Now().Unix(),
	}

	err := service.UpdatePostScore(context.Background(), post)
	if err != nil {
		t.Errorf("First request after timeout should succeed: %v", err)
	}

	err = service.UpdatePostScore(context.Background(), post)
	if err != nil {
		t.Errorf("Second request should succeed: %v", err)
	}

	if service.circuitBreaker.State() != StateClosed {
		t.Errorf("Expected circuit breaker to be closed after successes, got %s", service.circuitBreaker.State())
	}
}

func TestEmptyValueCacheIntegration_DeletePostCache(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	service.emptyValueCache.CacheEmpty(context.Background(), 999)

	if !service.emptyValueCache.IsEmptyCached(context.Background(), 999) {
		t.Fatal("Expected post 999 to be cached as empty")
	}

	err := service.DeletePostCache(context.Background(), 999)
	if err != nil {
		t.Errorf("DeletePostCache returned error: %v", err)
	}

	if service.emptyValueCache.IsEmptyCached(context.Background(), 999) {
		t.Error("Expected empty cache to be removed after DeletePostCache")
	}
}

func TestGetCachedPostDetail_EmptyCacheCheck(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	service.emptyValueCache.CacheEmpty(context.Background(), 999)

	result, err := service.GetCachedPostDetail(context.Background(), 999)
	if err == nil {
		t.Error("Expected error for empty cached post")
	}

	if !errors.Is(err, ErrPostNotFound) {
		t.Errorf("Expected ErrPostNotFound, got %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for empty cached post")
	}
}

func TestCachePostDetail_RemovesEmptyCache(t *testing.T) {
	service, mr := newTestHotPostServiceWithRedis()
	defer mr.Close()

	service.emptyValueCache.CacheEmpty(context.Background(), 1)

	if !service.emptyValueCache.IsEmptyCached(context.Background(), 1) {
		t.Fatal("Expected post 1 to be cached as empty initially")
	}

	post := &model.Post{
		ID:           1,
		UserID:       100,
		CommunityID:  1,
		Username:     "testuser",
		Title:        "Test Post",
		Content:      "Test Content",
		ViewCount:    150,
		CommentCount: 5,
		CreatedAt:    time.Now().Unix(),
	}

	err := service.CachePostDetail(context.Background(), post, 300)
	if err != nil {
		t.Errorf("CachePostDetail returned error: %v", err)
	}

	if service.emptyValueCache.IsEmptyCached(context.Background(), 1) {
		t.Error("Expected empty cache to be removed after caching post")
	}
}
