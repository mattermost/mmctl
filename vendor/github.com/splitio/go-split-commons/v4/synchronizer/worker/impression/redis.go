package impression

import (
	"errors"

	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-toolkit/v5/logging"
)

type RecorderRedis struct {
	impsInMemoryStorage storage.ImpressionStorageConsumer
	impsRedisStorage    storage.ImpressionStorageProducer
	logger              logging.LoggerInterface
}

// NewRecorderRedis creates new impressionsCount synchronizer for log impressionsCount in redis
func NewRecorderRedis(impsInMemoryStorage storage.ImpressionStorageConsumer, impsRedisStorage storage.ImpressionStorageProducer, logger logging.LoggerInterface) ImpressionRecorder {
	return &RecorderRedis{
		impsInMemoryStorage: impsInMemoryStorage,
		impsRedisStorage:    impsRedisStorage,
		logger:              logger,
	}
}

// SynchronizeImpressions syncs impressions
func (i *RecorderRedis) SynchronizeImpressions(bulkSize int64) error {
	queuedImpressions, err := i.impsInMemoryStorage.PopN(bulkSize)
	if err != nil {
		i.logger.Error("Error reading impressions queue", err)
		return errors.New("Error reading impressions queue")
	}

	if len(queuedImpressions) == 0 {
		i.logger.Debug("No impressions fetched from queue. Nothing to send")
		return nil
	}

	err = i.impsRedisStorage.LogImpressions(queuedImpressions)
	if err != nil {
		i.logger.Error("Error saving impressions in redis", err)
		return errors.New("Error saving impressions in redis")
	}

	return nil
}

// FlushImpressions flushes impressions
func (i *RecorderRedis) FlushImpressions(bulkSize int64) error {
	for !i.impsInMemoryStorage.Empty() {
		err := i.SynchronizeImpressions(bulkSize)
		if err != nil {
			return err
		}
	}

	return nil
}
