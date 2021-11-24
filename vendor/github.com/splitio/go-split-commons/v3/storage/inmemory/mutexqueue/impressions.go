package mutexqueue

import (
	"container/list"
	"sync"

	"github.com/splitio/go-split-commons/v3/dtos"
	"github.com/splitio/go-split-commons/v3/storage"
	"github.com/splitio/go-split-commons/v3/telemetry"
	"github.com/splitio/go-toolkit/v4/logging"
)

// NewMQImpressionsStorage returns an instance of MQEventsStorage
func NewMQImpressionsStorage(queueSize int, isFull chan<- string, logger logging.LoggerInterface, runtimeTelemetry storage.TelemetryRuntimeProducer) *MQImpressionsStorage {
	return &MQImpressionsStorage{
		queue:            list.New(),
		size:             queueSize,
		mutexQueue:       &sync.Mutex{},
		fullChan:         isFull,
		logger:           logger,
		runtimeTelemetry: runtimeTelemetry,
	}
}

// MQImpressionsStorage in memory events storage
type MQImpressionsStorage struct {
	queue            *list.List
	size             int
	mutexQueue       *sync.Mutex
	fullChan         chan<- string //only write channel
	logger           logging.LoggerInterface
	runtimeTelemetry storage.TelemetryRuntimeProducer
}

func (s *MQImpressionsStorage) sendSignalIsFull() {
	// Nom blocking select
	select {
	case s.fullChan <- "IMPRESSIONS_FULL":
		// Send "queue is full" signal
		break
	default:
		s.logger.Debug("Some error occurred on sending signal for impressions")
		break
	}
}

// Empty returns if slice len if zero
func (s *MQImpressionsStorage) Empty() bool {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()
	return s.queue.Len() == 0
}

// Count returns len
func (s *MQImpressionsStorage) Count() int64 {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()
	return int64(s.queue.Len())
}

// LogImpressions inserts impressions into the queue
func (s *MQImpressionsStorage) LogImpressions(impressions []dtos.Impression) error {
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	impressionsToAdd := len(impressions)

	for _, impression := range impressions {
		if s.queue.Len()+1 > s.size {
			s.sendSignalIsFull()
			s.runtimeTelemetry.RecordImpressionsStats(telemetry.ImpressionsDropped, int64(impressionsToAdd))
			return ErrorMaxSizeReached
		}
		// Add element
		s.queue.PushBack(impression)
		s.runtimeTelemetry.RecordImpressionsStats(telemetry.ImpressionsQueued, 1)
		impressionsToAdd--

		if s.queue.Len() == s.size {
			s.sendSignalIsFull()
		}
	}
	return nil
}

// PopN pop N elements from queue
func (s *MQImpressionsStorage) PopN(n int64) ([]dtos.Impression, error) {
	var toReturn []dtos.Impression
	var totalItems int

	// Mutexing queue
	s.mutexQueue.Lock()
	defer s.mutexQueue.Unlock()

	if int64(s.queue.Len()) >= n {
		totalItems = int(n)
	} else {
		totalItems = s.queue.Len()
	}

	toReturn = make([]dtos.Impression, totalItems)
	for i := 0; i < totalItems; i++ {
		toReturn[i] = s.queue.Remove(s.queue.Front()).(dtos.Impression)
	}

	return toReturn, nil
}

// PopNWithMetadata pop N elements from queue
func (s *MQImpressionsStorage) PopNWithMetadata(n int64) ([]dtos.ImpressionQueueObject, error) {
	panic("Not implemented for inmemory")
}

// Drop drops
func (s *MQImpressionsStorage) Drop(size *int64) error {
	panic("Not implemented for inmemory")
}
