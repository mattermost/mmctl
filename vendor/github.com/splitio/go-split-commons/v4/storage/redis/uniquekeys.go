package redis

import (
	"sync"

	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/go-toolkit/v5/redis"
)

type UniqueKeysMultiSdkConsumer struct {
	client   *redis.PrefixedRedisClient
	logger   logging.LoggerInterface
	mutex    *sync.RWMutex
	redisKey string
}

func NewUniqueKeysMultiSdkConsumer(
	redisClient *redis.PrefixedRedisClient,
	logger logging.LoggerInterface,
) storage.UniqueKeysMultiSdkConsumer {
	return &UniqueKeysMultiSdkConsumer{
		client:   redisClient,
		logger:   logger,
		mutex:    &sync.RWMutex{},
		redisKey: KeyUniquekeys,
	}
}

func (u *UniqueKeysMultiSdkConsumer) Count() int64 {
	val, err := u.client.LLen(u.redisKey)
	if err != nil {
		return 0
	}

	return val
}

func (u *UniqueKeysMultiSdkConsumer) PopNRaw(n int64) ([]string, int64, error) {
	lrange, left, err := u.pop(n)
	if err != nil {
		return nil, left, err
	}

	return lrange, left, nil
}

func (u *UniqueKeysMultiSdkConsumer) pop(n int64) ([]string, int64, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	lrange, err := u.client.LRange(u.redisKey, 0, n-1)
	if err != nil {
		u.logger.Error("Error fetching unique keys")
		return nil, 0, err
	}

	fetchedCount := int64(len(lrange))
	if fetchedCount == 0 {
		return nil, 0, nil
	}

	pipe := u.client.Pipeline()
	pipe.LTrim(u.redisKey, fetchedCount, int64(-1))
	pipe.LLen(u.redisKey)
	res, err := pipe.Exec()
	if len(res) < 2 || err != nil {
		u.logger.Error("Error trimming unique keys")
		return nil, 0, err
	}

	return lrange, res[1].Int(), err
}
