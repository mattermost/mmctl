package filter

import (
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/splitio/go-split-commons/v4/storage"
)

// BloomFilter description
type BloomFilter struct {
	mutex  *sync.RWMutex
	filter *bloom.BloomFilter
}

// NewBloomFilter description
func NewBloomFilter(expectedElemenets uint, falsePositiveProbability float64) storage.Filter {
	return &BloomFilter{
		mutex:  &sync.RWMutex{},
		filter: bloom.NewWithEstimates(expectedElemenets, falsePositiveProbability),
	}
}

// Add description
func (bf *BloomFilter) Add(data string) {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	bf.filter.Add([]byte(data))
}

// Contains description
func (bf *BloomFilter) Contains(data string) bool {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	return bf.filter.Test([]byte(data))
}

// Clear description
func (bf *BloomFilter) Clear() {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	bf.filter.ClearAll()
}
