package segment

import (
	"fmt"
	"sync"
	"time"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/healthcheck/application"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-split-commons/v4/telemetry"
	"github.com/splitio/go-toolkit/v5/common"
	"github.com/splitio/go-toolkit/v5/datastructures/set"
	"github.com/splitio/go-toolkit/v5/logging"
)

// Updater interface
type Updater interface {
	SynchronizeSegment(name string, till *int64, requestNoCache bool) (*UpdateResult, error)
	SynchronizeSegments(requestNoCache bool) (map[string]UpdateResult, error)
	SegmentNames() []interface{}
	IsSegmentCached(segmentName string) bool
}

// UpdaterImpl struct for segment sync
type UpdaterImpl struct {
	splitStorage     storage.SplitStorageConsumer
	segmentStorage   storage.SegmentStorage
	segmentFetcher   service.SegmentFetcher
	logger           logging.LoggerInterface
	runtimeTelemetry storage.TelemetryRuntimeProducer
	hcMonitor        application.MonitorProducerInterface
}

// UpdateResult encapsulates information regarding the segment update performed
type UpdateResult struct {
	UpdatedKeys     []string
	NewChangeNumber int64
}

// NewSegmentFetcher creates new segment synchronizer for processing segment updates
func NewSegmentFetcher(
	splitStorage storage.SplitStorage,
	segmentStorage storage.SegmentStorage,
	segmentFetcher service.SegmentFetcher,
	logger logging.LoggerInterface,
	runtimeTelemetry storage.TelemetryRuntimeProducer,
	hcMonitor application.MonitorProducerInterface,
) Updater {
	return &UpdaterImpl{
		splitStorage:     splitStorage,
		segmentStorage:   segmentStorage,
		segmentFetcher:   segmentFetcher,
		logger:           logger,
		runtimeTelemetry: runtimeTelemetry,
		hcMonitor:        hcMonitor,
	}
}

func (s *UpdaterImpl) processUpdate(segmentChanges *dtos.SegmentChangesDTO) {
	if len(segmentChanges.Added) == 0 && len(segmentChanges.Removed) == 0 && segmentChanges.Since != -1 {
		// If the Since is -1, it means the segment hasn't been fetched before.
		// In that case we need to proceed so that we store an empty list for cases that need
		// disambiguation between "segment isn't cached" & segment is empty (ie: split-proxy)
		return
	}

	toAdd := set.NewSet()
	toRemove := set.NewSet()
	// Segment exists, must add new members and remove old ones
	for _, key := range segmentChanges.Added {
		toAdd.Add(key)
	}
	for _, key := range segmentChanges.Removed {
		toRemove.Add(key)
	}

	s.logger.Debug(fmt.Sprintf("Segment [%s] exists, it will be updated. %d keys added, %d keys removed", segmentChanges.Name, toAdd.Size(), toRemove.Size()))
	s.segmentStorage.Update(segmentChanges.Name, toAdd, toRemove, segmentChanges.Till)

}

// SynchronizeSegment syncs segment
func (s *UpdaterImpl) SynchronizeSegment(name string, till *int64, requestNoCache bool) (*UpdateResult, error) {
	s.hcMonitor.NotifyEvent(application.Segments)
	var updatedKeys []string
	var err error
	var newCN int64

	for {
		s.logger.Debug(fmt.Sprintf("Synchronizing segment %s", name))
		changeNumber, _ := s.segmentStorage.ChangeNumber(name)
		if changeNumber == 0 {
			changeNumber = -1
		}
		if till != nil && *till < changeNumber {
			break
		}

		before := time.Now()
		var segmentChanges *dtos.SegmentChangesDTO
		segmentChanges, err = s.segmentFetcher.Fetch(name, changeNumber, requestNoCache)
		if err != nil {
			if httpError, ok := err.(*dtos.HTTPError); ok {
				s.runtimeTelemetry.RecordSyncError(telemetry.SegmentSync, httpError.Code)
			}
			break
		}

		newCN = segmentChanges.Till
		updatedKeys = append(updatedKeys, segmentChanges.Added...)
		updatedKeys = append(updatedKeys, segmentChanges.Removed...)
		s.runtimeTelemetry.RecordSyncLatency(telemetry.SegmentSync, time.Since(before))
		s.processUpdate(segmentChanges)
		if segmentChanges.Till == segmentChanges.Since || (till != nil && segmentChanges.Till >= *till) {
			s.runtimeTelemetry.RecordSuccessfulSync(telemetry.SegmentSync, time.Now().UTC())
			break
		}
	}

	return &UpdateResult{
		UpdatedKeys:     common.DedupeStringSlice(updatedKeys),
		NewChangeNumber: newCN,
	}, err
}

// SynchronizeSegments syncs segments at once
func (s *UpdaterImpl) SynchronizeSegments(requestNoCache bool) (map[string]UpdateResult, error) {
	segmentNames := s.splitStorage.SegmentNames().List()
	s.logger.Debug("Segment Sync", segmentNames)
	wg := sync.WaitGroup{}
	wg.Add(len(segmentNames))
	failedSegments := set.NewThreadSafeSet()

	var mtx sync.Mutex
	results := make(map[string]UpdateResult, len(segmentNames))
	for _, name := range segmentNames {
		conv, ok := name.(string)
		if !ok {
			s.logger.Warning("Skipping non-string segment present in storage at initialization-time!")
			continue
		}
		go func(segmentName string) {
			defer wg.Done() // Make sure the "finished" signal is always sent
			res, err := s.SynchronizeSegment(segmentName, nil, requestNoCache)
			if err != nil {
				failedSegments.Add(segmentName)
			}

			mtx.Lock()
			defer mtx.Unlock()
			results[segmentName] = *res
		}(conv)
	}
	wg.Wait()

	if failedSegments.Size() > 0 {
		return results, fmt.Errorf("The following segments failed to be fetched %v", failedSegments.List())
	}

	return results, nil
}

// SegmentNames returns all segments
func (s *UpdaterImpl) SegmentNames() []interface{} {
	return s.splitStorage.SegmentNames().List()
}

// IsSegmentCached returns true if a segment exists
func (s *UpdaterImpl) IsSegmentCached(segmentName string) bool {
	cn, _ := s.segmentStorage.ChangeNumber(segmentName)
	return cn != -1
}
