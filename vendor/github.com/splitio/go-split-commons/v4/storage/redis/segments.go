package redis

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-toolkit/v5/datastructures/set"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/go-toolkit/v5/redis"
)

// SegmentStorage is a redis implementation of a storage for segments
type SegmentStorage struct {
	client redis.PrefixedRedisClient
	logger logging.LoggerInterface
	mutext *sync.RWMutex
}

// NewSegmentStorage creates a new RedisSegmentStorage and returns a reference to it
// TODO(mredolatti): This should return the concrete type, not the implementation
func NewSegmentStorage(redisClient *redis.PrefixedRedisClient, logger logging.LoggerInterface) storage.SegmentStorage {
	return &SegmentStorage{
		client: *redisClient,
		logger: logger,
		mutext: &sync.RWMutex{},
	}
}

// ChangeNumber returns the changeNumber for a particular segment
func (r *SegmentStorage) ChangeNumber(segmentName string) (int64, error) {
	segmentKey := strings.Replace(KeySegmentTill, "{segment}", segmentName, 1)
	tillStr, err := r.client.Get(segmentKey)
	if err != nil {
		return -1, err
	}

	asInt, err := strconv.ParseInt(tillStr, 10, 64)
	if err != nil {
		r.logger.Error("Error retrieving till. Returning -1: ", err.Error())
		return -1, err
	}
	return asInt, nil
}

// Keys returns segments keys for segment if it's present
func (r *SegmentStorage) Keys(segmentName string) *set.ThreadUnsafeSet {
	keyToFetch := strings.Replace(KeySegment, "{segment}", segmentName, 1)
	segmentKeys, err := r.client.SMembers(keyToFetch)
	if len(segmentKeys) <= 0 {
		r.logger.Debug(fmt.Sprintf("Nonexsitent segment requested: %s", segmentName))
		return nil
	}
	if err != nil {
		r.logger.Error(fmt.Sprintf("Error retrieving members from set %s", segmentName))
		return nil
	}
	segment := set.NewSet()
	for _, member := range segmentKeys {
		segment.Add(member)
	}
	return segment
}

// SetChangeNumber sets the till value belong to segmentName
func (r *SegmentStorage) SetChangeNumber(segmentName string, changeNumber int64) error {
	segmentKey := strings.Replace(KeySegmentTill, "{segment}", segmentName, 1)
	return r.client.Set(segmentKey, changeNumber, 0)
}

// UpdateWithSummary returns errors instead of just logging them.
// This method should replace the current Update which should be deprecated in the next breaking change
func (r *SegmentStorage) UpdateWithSummary(name string, toAdd *set.ThreadUnsafeSet, toRemove *set.ThreadUnsafeSet, till int64) (int, int, error) {
	segmentKey := strings.Replace(KeySegment, "{segment}", name, 1)

	var mainErr SegmentUpdateError

	var removed int64
	if !toRemove.IsEmpty() {
		removed, mainErr.FailureToRemove = r.client.SRem(segmentKey, toRemove.List()...)
	}

	var added int64
	if !toAdd.IsEmpty() {
		added, mainErr.FailureToAdd = r.client.SAdd(segmentKey, toAdd.List()...)
	}
	r.SetChangeNumber(name, till)
	if mainErr.FailureToAdd != nil || mainErr.FailureToRemove != nil {
		return int(added), int(removed), &mainErr
	}
	return int(added), int(removed), nil
}

// Update adds a new segment
func (r *SegmentStorage) Update(name string, toAdd *set.ThreadUnsafeSet, toRemove *set.ThreadUnsafeSet, till int64) error {

	// TODO(mredolatti): Remove this mutex. This makes no sense here, but we need to make sure that no usage of `Update`
	// expects this sort of atomicity
	r.mutext.Lock()
	defer r.mutext.Unlock()

	if _, _, err := r.UpdateWithSummary(name, toAdd, toRemove, till); err != nil {
		r.logger.Error(fmt.Sprintf("error updating segment %s in redis: %s", name, err))
	}

	return nil
}

// Size returns the number of keys in a segment
func (r *SegmentStorage) Size(name string) (int, error) {
	res, err := r.client.SCard(strings.Replace(KeySegment, "{segment}", name, 1))
	return int(res), err

}

// SegmentContainsKey returns true if the segment contains a specific key
func (r *SegmentStorage) SegmentContainsKey(segmentName string, key string) (bool, error) {
	segmentKey := strings.Replace(KeySegment, "{segment}", segmentName, 1)
	exists := r.client.SIsMember(segmentKey, key)
	return exists, nil
}

// SegmentKeysCount method
func (r *SegmentStorage) SegmentKeysCount() int64 { return 0 }

// static interface compliance assertions
var _ storage.SegmentStorage = (*SegmentStorage)(nil)
var _ storage.SegmentStorageConsumer = (*SegmentStorage)(nil)
var _ storage.SegmentStorageProducer = (*SegmentStorage)(nil)
