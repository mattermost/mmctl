package redis

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/go-toolkit/v5/queuecache"
	"github.com/splitio/go-toolkit/v5/redis"
)

// EventsStorage redis implementation of EventsStorage interface
type EventsStorage struct {
	cache       queuecache.InMemoryQueueCacheOverlay
	client      *redis.PrefixedRedisClient
	logger      logging.LoggerInterface
	metadata    dtos.Metadata
	redisKey    string
	refillMutex *sync.RWMutex
	mutex       *sync.RWMutex
}

// maxAccumulatedSize is the maximum number of bytes to be fetched from cache before posting to the backend
const maxAccumulatedSize = 5 * 1024 * 1024

// maxEventSize is the maximum allowed event size
const maxEventSize = 32 * 1024

// NewEventStorageConsumer storage for consumer
func NewEventStorageConsumer(redisClient *redis.PrefixedRedisClient, metadata dtos.Metadata, logger logging.LoggerInterface) *EventsStorage {
	return &EventsStorage{
		cache:       queuecache.InMemoryQueueCacheOverlay{},
		client:      redisClient,
		logger:      logger,
		metadata:    metadata,
		redisKey:    KeyEvents,
		refillMutex: &sync.RWMutex{},
		mutex:       &sync.RWMutex{},
	}
}

// NewEventsStorage returns an instance of RedisEventsStorage
func NewEventsStorage(redisClient *redis.PrefixedRedisClient, metadata dtos.Metadata, logger logging.LoggerInterface) *EventsStorage {
	refillMutex := &sync.RWMutex{}
	refillFunc := func(count int) ([]interface{}, error) {
		refillMutex.Lock()
		defer refillMutex.Unlock()
		lrange, err := redisClient.LRange(KeyEvents, 0, int64(count-1))
		if err != nil {
			logger.Error("Fetching events", err)
			return nil, err
		}
		totalFetchedEvents := len(lrange)

		idxFrom := count
		if totalFetchedEvents < count {
			idxFrom = totalFetchedEvents
		}

		err = redisClient.LTrim(KeyEvents, int64(idxFrom), -1)
		if err != nil {
			logger.Error("Trim events", err)
			return nil, err
		}

		toReturn := make([]interface{}, len(lrange))
		for index, item := range lrange {
			toReturn[index] = item
		}
		return toReturn, nil
	}

	return &EventsStorage{
		cache:       *queuecache.New(10000, refillFunc),
		client:      redisClient,
		logger:      logger,
		metadata:    metadata,
		redisKey:    KeyEvents,
		refillMutex: refillMutex,
		mutex:       &sync.RWMutex{},
	}
}

// Push events into Redis LIST data type with RPUSH command
func (r *EventsStorage) Push(event dtos.EventDTO, _ int) error {
	var queueMessage = dtos.QueueStoredEventDTO{Metadata: r.metadata, Event: event}

	eventJSON, err := json.Marshal(queueMessage)
	if err != nil {
		r.logger.Error("Something were wrong marshaling provided event to JSON", err.Error())
		return err
	}

	r.logger.Debug("Pushing events to:", r.redisKey, string(eventJSON))

	_, errPush := r.client.RPush(r.redisKey, eventJSON)
	if errPush != nil {
		r.logger.Error("Something were wrong pushing event to redis", errPush)
		return errPush
	}

	return nil
}

func (r *EventsStorage) pop(n int64) ([]string, int64, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	lrange, err := r.client.LRange(r.redisKey, 0, n-1)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading events: %w", err)
	}

	fetchedCount := int64(len(lrange))
	if fetchedCount == 0 {
		return nil, 0, nil
	}

	pipe := r.client.Pipeline()
	pipe.LTrim(r.redisKey, fetchedCount, int64(-1))
	pipe.LLen(r.redisKey)
	res, err := pipe.Exec()
	if len(res) < 2 || err != nil {
		r.logger.Error("Error trimming impressions")
		return nil, 0, err
	}

	return lrange, res[1].Int(), err
}

// PopNWithMetadata pop N elements from queue
func (r *EventsStorage) PopNWithMetadata(n int64) ([]dtos.QueueStoredEventDTO, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	toReturn := make([]dtos.QueueStoredEventDTO, n)
	var err error
	fetchedCount := 0
	accumulatedSize := 0
	writeIndex := 0
	for r.Count() > 0 && int64(fetchedCount) < n && accumulatedSize+maxEventSize < maxAccumulatedSize && err == nil {
		numberOfItemsToFetch := int(math.Min(
			float64((maxAccumulatedSize-accumulatedSize)/maxEventSize),
			float64(n-int64(fetchedCount)),
		))
		elems, err := r.cache.Fetch(numberOfItemsToFetch)
		if err != nil {
			r.logger.Error("Error fetching events", err.Error())
			break
		}

		for _, elem := range elems {
			asStr, ok := elem.(string)
			if !ok {
				r.logger.Error("Error type-asserting event as string", err.Error())
				continue
			}

			storedEventDTO := dtos.QueueStoredEventDTO{}
			err = json.Unmarshal([]byte(asStr), &storedEventDTO)
			if err != nil {
				r.logger.Error("Error decoding event JSON", err.Error())
				continue
			}
			accumulatedSize += storedEventDTO.Event.Size()
			toReturn[writeIndex] = storedEventDTO
			writeIndex++
		}
		fetchedCount += len(elems)
	}
	return toReturn[0:writeIndex], nil
}

// Count returns the number of items in the redis list
func (r *EventsStorage) Count() int64 {
	val, err := r.client.LLen(r.redisKey)
	if err != nil {
		return 0
	}
	return val
}

// Empty returns true if redis list is zero length
func (r *EventsStorage) Empty() bool {
	return r.Count() == 0
}

// Drop drops events from queue
func (r *EventsStorage) Drop(size int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if size == -1 {
		_, err := r.client.Del(r.redisKey)
		return err
	}
	return r.client.LTrim(r.redisKey, size, -1)
}

// PopNRaw pops N elements and returns them as raw strings
func (r *EventsStorage) PopNRaw(n int64) ([]string, int64, error) {
	lrange, left, err := r.pop(n)
	if err != nil {
		return nil, 0, err
	}

	return lrange, left, nil
}
