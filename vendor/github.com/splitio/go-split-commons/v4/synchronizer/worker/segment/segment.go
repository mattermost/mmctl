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
	"github.com/splitio/go-toolkit/v5/backoff"
	"github.com/splitio/go-toolkit/v5/common"
	"github.com/splitio/go-toolkit/v5/datastructures/set"
	"github.com/splitio/go-toolkit/v5/logging"
)

const (
	onDemandFetchBackoffBase       = int64(10)        // backoff base starting at 10 seconds
	onDemandFetchBackoffMaxWait    = 60 * time.Second //  don't sleep for more than 1 minute
	onDemandFetchBackoffMaxRetries = 10
)

// Updater interface
type Updater interface {
	SynchronizeSegment(name string, till *int64) (*UpdateResult, error)
	SynchronizeSegments() (map[string]UpdateResult, error)
	SegmentNames() []interface{}
	IsSegmentCached(segmentName string) bool
}

// UpdaterImpl struct for segment sync
type UpdaterImpl struct {
	splitStorage                storage.SplitStorageConsumer
	segmentStorage              storage.SegmentStorage
	segmentFetcher              service.SegmentFetcher
	logger                      logging.LoggerInterface
	runtimeTelemetry            storage.TelemetryRuntimeProducer
	hcMonitor                   application.MonitorProducerInterface
	onDemandFetchBackoffBase    int64
	onDemandFetchBackoffMaxWait time.Duration
}

// UpdateResult encapsulates information regarding the segment update performed
type UpdateResult struct {
	UpdatedKeys     []string
	NewChangeNumber int64
}

type internalSegmentSync struct {
	updateResult   *UpdateResult
	successfulSync bool
	attempt        int
}

// NewSegmentFetcher creates new segment synchronizer for processing segment updates
func NewSegmentFetcher(
	splitStorage storage.SplitStorage,
	segmentStorage storage.SegmentStorage,
	segmentFetcher service.SegmentFetcher,
	logger logging.LoggerInterface,
	runtimeTelemetry storage.TelemetryRuntimeProducer,
	hcMonitor application.MonitorProducerInterface,
) *UpdaterImpl {
	return &UpdaterImpl{
		splitStorage:                splitStorage,
		segmentStorage:              segmentStorage,
		segmentFetcher:              segmentFetcher,
		logger:                      logger,
		runtimeTelemetry:            runtimeTelemetry,
		hcMonitor:                   hcMonitor,
		onDemandFetchBackoffBase:    onDemandFetchBackoffBase,
		onDemandFetchBackoffMaxWait: onDemandFetchBackoffMaxWait,
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

func (s *UpdaterImpl) fetchUntil(name string, till *int64, fetchOptions *service.FetchOptions) (*UpdateResult, error) {
	var updatedKeys []string
	var err error
	var currentSince int64

	for {
		s.logger.Debug(fmt.Sprintf("Synchronizing segment %s", name))
		currentSince, _ = s.segmentStorage.ChangeNumber(name)

		before := time.Now()
		var segmentChanges *dtos.SegmentChangesDTO
		segmentChanges, err = s.segmentFetcher.Fetch(name, currentSince, fetchOptions)
		if err != nil {
			if httpError, ok := err.(*dtos.HTTPError); ok {
				s.runtimeTelemetry.RecordSyncError(telemetry.SegmentSync, httpError.Code)
			}
			break
		}

		currentSince = segmentChanges.Till
		updatedKeys = append(updatedKeys, segmentChanges.Added...)
		updatedKeys = append(updatedKeys, segmentChanges.Removed...)
		s.runtimeTelemetry.RecordSyncLatency(telemetry.SegmentSync, time.Since(before))
		s.processUpdate(segmentChanges)
		if currentSince == segmentChanges.Since {
			s.runtimeTelemetry.RecordSuccessfulSync(telemetry.SegmentSync, time.Now().UTC())
			break
		}
	}

	return &UpdateResult{
		UpdatedKeys:     common.DedupeStringSlice(updatedKeys),
		NewChangeNumber: currentSince,
	}, err
}

func (s *UpdaterImpl) attemptSegmentSync(name string, till *int64, fetchOptions *service.FetchOptions) (internalSegmentSync, error) {
	internalBackoff := backoff.New(s.onDemandFetchBackoffBase, s.onDemandFetchBackoffMaxWait)
	remainingAttempts := onDemandFetchBackoffMaxRetries
	for {
		remainingAttempts = remainingAttempts - 1
		updateResult, err := s.fetchUntil(name, till, fetchOptions) // what we should do with err
		if err != nil || remainingAttempts <= 0 {
			return internalSegmentSync{updateResult: updateResult, successfulSync: false, attempt: remainingAttempts}, err
		}
		if till == nil || *till <= updateResult.NewChangeNumber {
			return internalSegmentSync{updateResult: updateResult, successfulSync: true, attempt: remainingAttempts}, nil
		}
		howLong := internalBackoff.Next()
		time.Sleep(howLong)
	}
}

// SynchronizeSegment syncs segment
func (s *UpdaterImpl) SynchronizeSegment(name string, till *int64) (*UpdateResult, error) {
	fetchOptions := service.NewFetchOptions(true, nil)
	s.hcMonitor.NotifyEvent(application.Segments)

	currentSince, _ := s.segmentStorage.ChangeNumber(name)
	if till != nil && *till <= currentSince { // the passed till is less than change_number, no need to perform updates
		return &UpdateResult{}, nil
	}

	internalSyncResult, err := s.attemptSegmentSync(name, till, &fetchOptions)
	attempts := onDemandFetchBackoffMaxRetries - internalSyncResult.attempt
	if err != nil {
		return internalSyncResult.updateResult, err
	}
	if internalSyncResult.successfulSync {
		s.logger.Debug(fmt.Sprintf("Refresh completed in %d attempts.", attempts))
		return internalSyncResult.updateResult, nil
	}
	withCDNBypass := service.NewFetchOptions(true, &internalSyncResult.updateResult.NewChangeNumber) // Set flag for bypassing CDN
	internalSyncResultCDNBypass, err := s.attemptSegmentSync(name, till, &withCDNBypass)
	withoutCDNattempts := onDemandFetchBackoffMaxRetries - internalSyncResultCDNBypass.attempt
	if err != nil {
		return internalSyncResultCDNBypass.updateResult, err
	}
	if internalSyncResultCDNBypass.successfulSync {
		s.logger.Debug(fmt.Sprintf("Refresh completed bypassing the CDN in %d attempts.", withoutCDNattempts))
		return internalSyncResultCDNBypass.updateResult, nil
	}
	s.logger.Debug(fmt.Sprintf("No changes fetched after %d attempts with CDN bypassed.", withoutCDNattempts))
	return internalSyncResultCDNBypass.updateResult, nil
}

// SynchronizeSegments syncs segments at once
func (s *UpdaterImpl) SynchronizeSegments() (map[string]UpdateResult, error) {
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
			res, err := s.SynchronizeSegment(segmentName, nil)
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
		return results, fmt.Errorf("the following segments failed to be fetched %v", failedSegments.List())
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
