package repository

import (
	"testing"
)

func TestPostBloomFilter_AddAndCheck(t *testing.T) {
	bf := NewPostBloomFilter(1000, 0.01, nil)

	postIDs := []int64{1, 100, 1000, 10000, 99999}
	for _, id := range postIDs {
		bf.Add(id)
	}

	for _, id := range postIDs {
		if !bf.MightContain(id) {
			t.Errorf("PostID %d should be in bloom filter", id)
		}
	}

	notExistsIDs := []int64{2, 200, 2000}
	falsePositiveCount := 0
	for _, id := range notExistsIDs {
		if bf.MightContain(id) {
			falsePositiveCount++
		}
	}

	falsePositiveRate := float64(falsePositiveCount) / float64(len(notExistsIDs))
	t.Logf("False positive rate: %.2f%%", falsePositiveRate*100)

	if falsePositiveRate > 0.1 {
		t.Logf("Warning: false positive rate %.2f%% is higher than expected", falsePositiveRate*100)
	}
}

func TestPostBloomFilter_AddMultiple(t *testing.T) {
	bf := NewPostBloomFilter(10000, 0.01, nil)

	postIDs := make([]int64, 100)
	for i := 0; i < 100; i++ {
		postIDs[i] = int64(i + 1)
	}

	bf.AddMultiple(postIDs)

	for _, id := range postIDs {
		if !bf.MightContain(id) {
			t.Errorf("PostID %d should be in bloom filter", id)
		}
	}
}

func TestPostBloomFilter_Clear(t *testing.T) {
	bf := NewPostBloomFilter(1000, 0.01, nil)

	bf.Add(1)
	bf.Add(2)
	bf.Add(3)

	if !bf.MightContain(1) {
		t.Error("PostID 1 should be in bloom filter before clear")
	}

	bf.Clear()

	if bf.MightContain(1) {
		t.Error("PostID 1 should not be in bloom filter after clear")
	}
	if bf.MightContain(2) {
		t.Error("PostID 2 should not be in bloom filter after clear")
	}
	if bf.MightContain(3) {
		t.Error("PostID 3 should not be in bloom filter after clear")
	}
}

func TestPostBloomFilter_Size(t *testing.T) {
	bf := NewPostBloomFilter(1000, 0.01, nil)

	initialSize := bf.Size()
	t.Logf("Initial size: %d", initialSize)

	bf.Add(1)
	bf.Add(2)
	bf.Add(3)

	afterAddSize := bf.Size()
	t.Logf("Size after adding 3 items: %d", afterAddSize)

	if afterAddSize < initialSize {
		t.Error("Size should increase after adding items")
	}
}

func TestPostBloomFilter_FalsePositiveRate(t *testing.T) {
	expectedItems := uint(10000)
	falsePositiveRate := 0.01

	bf := NewPostBloomFilter(expectedItems, falsePositiveRate, nil)

	for i := int64(1); i <= int64(expectedItems); i++ {
		bf.Add(i)
	}

	testCount := 10000
	falsePositiveCount := 0

	for i := int64(expectedItems) + 1; i <= int64(expectedItems)+int64(testCount); i++ {
		if bf.MightContain(i) {
			falsePositiveCount++
		}
	}

	actualFalsePositiveRate := float64(falsePositiveCount) / float64(testCount)
	t.Logf("Expected false positive rate: %.2f%%", falsePositiveRate*100)
	t.Logf("Actual false positive rate: %.2f%%", actualFalsePositiveRate*100)

	tolerance := falsePositiveRate * 3
	if actualFalsePositiveRate > falsePositiveRate+tolerance {
		t.Errorf("False positive rate %.2f%% exceeds tolerance", actualFalsePositiveRate*100)
	}
}

func TestPostBloomFilter_Int64ToBytes(t *testing.T) {
	bf := &PostBloomFilter{}

	testCases := []int64{0, 1, -1, 123456789, -987654321}

	for _, tc := range testCases {
		bytes := bf.int64ToBytes(tc)
		if len(bytes) != 8 {
			t.Errorf("int64ToBytes(%d) returned %d bytes, expected 8", tc, len(bytes))
		}
	}
}
