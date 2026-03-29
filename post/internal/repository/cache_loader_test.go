package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"post/internal/model"

	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestCacheLoader_GetOrLoad_CacheHit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	localCache := NewHotPostLocalCache(100, 5*time.Minute, 10)
	loader := NewCacheLoader(r, localCache)

	post := &model.Post{
		ID:        1,
		Title:     "Test Post",
		ViewCount: 100,
	}
	localCache.Set(post)

	loadCalled := false
	result, err := loader.GetOrLoad(context.Background(), 1, func(ctx context.Context) (*model.Post, error) {
		loadCalled = true
		return nil, nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected post from cache")
	}

	if loadCalled {
		t.Error("Load function should not be called when cache hit")
	}
}

func TestCacheLoader_GetOrLoad_CacheMiss(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	localCache := NewHotPostLocalCache(100, 5*time.Minute, 10)
	loader := NewCacheLoader(r, localCache)

	post := &model.Post{
		ID:        1,
		Title:     "Test Post",
		ViewCount: 100,
	}

	loadCalled := false
	result, err := loader.GetOrLoad(context.Background(), 1, func(ctx context.Context) (*model.Post, error) {
		loadCalled = true
		return post, nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected post from loader")
	}

	if !loadCalled {
		t.Error("Load function should be called when cache miss")
	}

	cached, found := localCache.Get(1)
	if !found {
		t.Error("Expected post to be cached after load")
	}

	if cached.ID != post.ID {
		t.Errorf("Expected cached post ID %d, got %d", post.ID, cached.ID)
	}
}

func TestCacheLoader_GetOrLoad_LoadError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	localCache := NewHotPostLocalCache(100, 5*time.Minute, 10)
	loader := NewCacheLoader(r, localCache)

	loadError := errors.New("load error")
	result, err := loader.GetOrLoad(context.Background(), 1, func(ctx context.Context) (*model.Post, error) {
		return nil, loadError
	})

	if err != loadError {
		t.Errorf("Expected load error, got: %v", err)
	}

	if result != nil {
		t.Error("Expected nil result on error")
	}
}

func TestCacheLoader_GetOrLoad_Concurrent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	localCache := NewHotPostLocalCache(100, 5*time.Minute, 10)
	loader := NewCacheLoader(r, localCache)

	post := &model.Post{
		ID:        1,
		Title:     "Test Post",
		ViewCount: 100,
	}

	var loadCount int
	var mu sync.Mutex
	var successCount int

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := loader.GetOrLoad(context.Background(), 1, func(ctx context.Context) (*model.Post, error) {
				mu.Lock()
				loadCount++
				mu.Unlock()
				time.Sleep(50 * time.Millisecond)
				return post, nil
			})
			if err == nil && result != nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successCount < 1 {
		t.Errorf("Expected at least 1 successful load, got %d", successCount)
	}

	t.Logf("Load called %d times, success count %d", loadCount, successCount)
}

func TestCacheLoader_GetOrLoad_ContextCancel(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	localCache := NewHotPostLocalCache(100, 5*time.Minute, 10)
	loader := NewCacheLoader(r, localCache)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := loader.GetOrLoad(ctx, 1, func(ctx context.Context) (*model.Post, error) {
		return nil, nil
	})

	if result != nil {
		t.Error("Expected nil result on context cancel")
	}
}

func TestCacheLoader_GetOrLoad_LowViewCount(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	localCache := NewHotPostLocalCache(100, 5*time.Minute, 10)
	loader := NewCacheLoader(r, localCache)

	post := &model.Post{
		ID:        1,
		Title:     "Low View Post",
		ViewCount: 5,
	}

	result, err := loader.GetOrLoad(context.Background(), 1, func(ctx context.Context) (*model.Post, error) {
		return post, nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected post from loader")
	}

	_, found := localCache.Get(1)
	if found {
		t.Error("Low view count post should not be cached")
	}
}
