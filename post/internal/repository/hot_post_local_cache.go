package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"post/internal/model"

	"github.com/allegro/bigcache/v3"
)

type HotPostLocalCache struct {
	cache     *bigcache.BigCache
	threshold int
}

func NewHotPostLocalCache(maxEntries int, ttl time.Duration, threshold int) *HotPostLocalCache {
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(ttl))
	if err != nil {
		panic(fmt.Sprintf("failed to create bigcache: %v", err))
	}

	return &HotPostLocalCache{
		cache:     cache,
		threshold: threshold,
	}
}

func (c *HotPostLocalCache) Get(postID int64) (*model.Post, bool) {
	key := c.formatKey(postID)
	data, err := c.cache.Get(key)
	if err != nil {
		return nil, false
	}

	var post model.Post
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, false
	}

	return &post, true
}

func (c *HotPostLocalCache) Set(post *model.Post) {
	if post == nil {
		return
	}

	if post.ViewCount < int64(c.threshold) {
		return
	}

	key := c.formatKey(post.ID)
	data, err := json.Marshal(post)
	if err != nil {
		return
	}

	c.cache.Set(key, data)
}

func (c *HotPostLocalCache) Delete(postID int64) {
	key := c.formatKey(postID)
	c.cache.Delete(key)
}

func (c *HotPostLocalCache) UpdateViewCount(postID int64, viewCount int64) {
	post, found := c.Get(postID)
	if !found {
		return
	}

	post.ViewCount = viewCount
	c.Set(post)
}

func (c *HotPostLocalCache) ShouldCache(viewCount int64) bool {
	return viewCount >= int64(c.threshold)
}

func (c *HotPostLocalCache) Size() int {
	return c.cache.Len()
}

func (c *HotPostLocalCache) formatKey(postID int64) string {
	return fmt.Sprintf("%d", postID)
}
