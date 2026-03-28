package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestEmptyValueCache_New(t *testing.T) {
	cache := NewEmptyValueCache(nil, 60)
	if cache == nil {
		t.Fatal("Expected empty value cache to be created")
	}
}

func TestEmptyValueCache_CacheEmpty(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	cache := NewEmptyValueCache(r, 60)

	err = cache.CacheEmpty(context.Background(), 123)
	if err != nil {
		t.Errorf("Failed to cache empty value: %v", err)
	}

	if !cache.IsEmptyCached(context.Background(), 123) {
		t.Error("Expected post 123 to be cached as empty")
	}
}

func TestEmptyValueCache_IsEmptyCached_LocalCache(t *testing.T) {
	cache := NewEmptyValueCache(nil, 60)

	if cache.IsEmptyCached(context.Background(), 123) {
		t.Error("Expected post 123 to not be cached as empty")
	}

	cache.CacheEmpty(context.Background(), 123)

	if !cache.IsEmptyCached(context.Background(), 123) {
		t.Error("Expected post 123 to be cached as empty in local cache")
	}
}

func TestEmptyValueCache_RemoveEmptyCache(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	cache := NewEmptyValueCache(r, 60)

	cache.CacheEmpty(context.Background(), 123)

	if !cache.IsEmptyCached(context.Background(), 123) {
		t.Fatal("Expected post 123 to be cached as empty")
	}

	err = cache.RemoveEmptyCache(context.Background(), 123)
	if err != nil {
		t.Errorf("Failed to remove empty cache: %v", err)
	}

	if cache.IsEmptyCached(context.Background(), 123) {
		t.Error("Expected post 123 to not be cached as empty after removal")
	}
}

func TestEmptyValueCache_CustomTTL(t *testing.T) {
	cache := NewEmptyValueCache(nil, 60)

	err := cache.CacheEmpty(context.Background(), 456, 120)
	if err != nil {
		t.Errorf("Failed to cache empty value with custom TTL: %v", err)
	}

	if !cache.IsEmptyCached(context.Background(), 456) {
		t.Error("Expected post 456 to be cached as empty")
	}
}

func TestEmptyValueCache_CachePostNotFound(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	cache := NewEmptyValueCache(r, 60)

	err = cache.CachePostNotFound(context.Background(), 789)
	if err != nil {
		t.Errorf("Failed to cache post not found: %v", err)
	}

	if !cache.IsEmptyCached(context.Background(), 789) {
		t.Error("Expected post 789 to be cached as not found")
	}
}

func TestSyncMapCache_SetGet(t *testing.T) {
	cache := NewSyncMapCache()

	cache.Set("test_key", true, 60*time.Second)

	value, found := cache.Get("test_key")
	if !found {
		t.Error("Expected to find key in cache")
	}
	if !value {
		t.Error("Expected value to be true")
	}
}

func TestSyncMapCache_Expired(t *testing.T) {
	cache := NewSyncMapCache()

	cache.Set("test_key", true, 100*time.Millisecond)

	time.Sleep(150 * time.Millisecond)

	_, found := cache.Get("test_key")
	if found {
		t.Error("Expected key to be expired")
	}
}

func TestSyncMapCache_Delete(t *testing.T) {
	cache := NewSyncMapCache()

	cache.Set("test_key", true, 60*time.Second)

	cache.Delete("test_key")

	_, found := cache.Get("test_key")
	if found {
		t.Error("Expected key to be deleted")
	}
}

func TestEmptyValueCache_ClearAllEmptyCache(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}
	defer mr.Close()

	r := redis.New(mr.Addr())
	cache := NewEmptyValueCache(r, 60)

	cache.CacheEmpty(context.Background(), 1)
	cache.CacheEmpty(context.Background(), 2)
	cache.CacheEmpty(context.Background(), 3)

	err = cache.ClearAllEmptyCache(context.Background())
	if err != nil {
		t.Errorf("Failed to clear all empty cache: %v", err)
	}

	if cache.IsEmptyCached(context.Background(), 1) {
		t.Error("Expected post 1 to not be cached after clear")
	}
	if cache.IsEmptyCached(context.Background(), 2) {
		t.Error("Expected post 2 to not be cached after clear")
	}
}
