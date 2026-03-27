package repository

import (
	"testing"
	"time"

	"post/internal/model"
)

func newTestHotPostService() *HotPostService {
	localCache := NewHotPostLocalCache(1000, 5*time.Minute, 1000)
	return NewHotPostService(nil, localCache)
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
