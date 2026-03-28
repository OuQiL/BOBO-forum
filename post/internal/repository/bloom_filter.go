package repository

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	BloomFilterKey = "post:bloom:filter"
)

type PostBloomFilter struct {
	filter    *bloom.BloomFilter
	redis     *redis.Redis
	mu        sync.RWMutex
	dirty     bool
}

func NewPostBloomFilter(expectedItems uint, falsePositiveRate float64, rds *redis.Redis) *PostBloomFilter {
	filter := bloom.NewWithEstimates(expectedItems, falsePositiveRate)

	pbf := &PostBloomFilter{
		filter: filter,
		redis:  rds,
		dirty:  false,
	}

	return pbf
}

func (pbf *PostBloomFilter) Add(postID int64) {
	pbf.mu.Lock()
	defer pbf.mu.Unlock()

	data := pbf.int64ToBytes(postID)
	pbf.filter.Add(data)
	pbf.dirty = true
}

func (pbf *PostBloomFilter) MightContain(postID int64) bool {
	pbf.mu.RLock()
	defer pbf.mu.RUnlock()

	data := pbf.int64ToBytes(postID)
	return pbf.filter.Test(data)
}

func (pbf *PostBloomFilter) AddMultiple(postIDs []int64) {
	pbf.mu.Lock()
	defer pbf.mu.Unlock()

	for _, id := range postIDs {
		data := pbf.int64ToBytes(id)
		pbf.filter.Add(data)
	}
	pbf.dirty = true
}

func (pbf *PostBloomFilter) PersistToRedis(ctx context.Context) error {
	pbf.mu.Lock()
	defer pbf.mu.Unlock()

	if !pbf.dirty {
		return nil
	}

	data, err := pbf.filter.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal bloom filter: %w", err)
	}

	if err := pbf.redis.Setex(BloomFilterKey, string(data), 0); err != nil {
		return fmt.Errorf("failed to persist bloom filter to redis: %w", err)
	}

	pbf.dirty = false
	return nil
}

func (pbf *PostBloomFilter) LoadFromRedis(ctx context.Context) error {
	pbf.mu.Lock()
	defer pbf.mu.Unlock()

	data, err := pbf.redis.Get(BloomFilterKey)
	if err != nil {
		return nil
	}

	if data == "" {
		return nil
	}

	if err := pbf.filter.UnmarshalBinary([]byte(data)); err != nil {
		return fmt.Errorf("failed to unmarshal bloom filter: %w", err)
	}

	return nil
}

func (pbf *PostBloomFilter) Clear() {
	pbf.mu.Lock()
	defer pbf.mu.Unlock()

	pbf.filter.ClearAll()
	pbf.dirty = true
}

func (pbf *PostBloomFilter) Size() uint {
	pbf.mu.RLock()
	defer pbf.mu.RUnlock()

	return uint(pbf.filter.ApproximatedSize())
}

func (pbf *PostBloomFilter) int64ToBytes(n int64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(n))
	return data
}
