package redis

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/go-toolkit/v5/redis"
)

type ImpressionsCountStorageImp struct {
	client   *redis.PrefixedRedisClient
	mutex    *sync.Mutex
	logger   logging.LoggerInterface
	redisKey string
}

func NewImpressionsCountStorage(client *redis.PrefixedRedisClient, logger logging.LoggerInterface) storage.ImpressionsCountStorage {
	return &ImpressionsCountStorageImp{
		client:   client,
		mutex:    &sync.Mutex{},
		logger:   logger,
		redisKey: KeyImpressionsCount,
	}
}

func (r *ImpressionsCountStorageImp) GetImpressionsCount() (*dtos.ImpressionsCountDTO, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	res, err := r.client.HGetAll(r.redisKey)
	if err != nil {
		r.logger.Error("Error reading impressions count, %w", err)
		return nil, err
	}

	r.client.Del(r.redisKey)

	toReturn := dtos.ImpressionsCountDTO{PerFeature: []dtos.ImpressionsInTimeFrameDTO{}}

	for key, value := range res {
		nameandtime := strings.Split(key, "::")
		if len(nameandtime) != 2 {
			r.logger.Error("Error spliting key from redis, %w", err)
			continue
		}
		featureName := nameandtime[0]
		timeFrame, err := strconv.ParseInt(nameandtime[1], 10, 64)
		if err != nil {
			r.logger.Error("Error parsing time frame from redis, %w", err)
			continue
		}
		rawCount, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			r.logger.Error("Error parsing raw count from redis, %w", err)
			continue
		}

		toReturn.PerFeature = append(toReturn.PerFeature, dtos.ImpressionsInTimeFrameDTO{
			FeatureName: featureName,
			TimeFrame:   timeFrame,
			RawCount:    rawCount,
		})
	}

	return &toReturn, nil
}

func (r *ImpressionsCountStorageImp) RecordImpressionsCount(impressions dtos.ImpressionsCountDTO) error {
	if len(impressions.PerFeature) < 1 {
		r.logger.Debug("Impression Count list is empty, nothing to record.")
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	pipe := r.client.Pipeline()
	for _, value := range impressions.PerFeature {
		pipe.HIncrBy(r.redisKey, fmt.Sprintf("%s::%d", value.FeatureName, value.TimeFrame), value.RawCount)
	}

	pipe.HLen(r.redisKey)
	res, err := pipe.Exec()
	if err != nil {
		r.logger.Error("Error incrementing impressions count, %w", err)
		return err
	}
	if len(res) < len(impressions.PerFeature) {
		return fmt.Errorf("Error incrementing impressions count")
	}

	// Checks if expiration needs to be set
	if shouldSetExpirationKey(&impressions, res) {
		r.logger.Debug("Proceeding to set expiration for: ", r.redisKey)
		result := r.client.Expire(r.redisKey, time.Duration(TTLImpressions)*time.Second)
		if !result {
			r.logger.Error("Something were wrong setting expiration for %s", r.redisKey)
		}
	}

	return nil
}

func shouldSetExpirationKey(impressions *dtos.ImpressionsCountDTO, res []redis.Result) bool {
	hlenRes := res[len(res)-1]

	var totalCounts int64
	var resCounts int64
	for i := 0; i < len(impressions.PerFeature); i++ {
		totalCounts += impressions.PerFeature[i].RawCount
		resCounts += res[i].Int()
	}

	return totalCounts+int64(len(impressions.PerFeature)) == hlenRes.Int()+resCounts
}
