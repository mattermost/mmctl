package split

import (
	"fmt"
	"time"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/healthcheck/application"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-split-commons/v4/telemetry"
	"github.com/splitio/go-toolkit/v5/backoff"
	"github.com/splitio/go-toolkit/v5/common"
	"github.com/splitio/go-toolkit/v5/logging"
)

const (
	matcherTypeInSegment           = "IN_SEGMENT"
	onDemandFetchBackoffBase       = int64(10)        // backoff base starting at 10 seconds
	onDemandFetchBackoffMaxWait    = 60 * time.Second //  don't sleep for more than 1 minute
	onDemandFetchBackoffMaxRetries = 10
)

// Updater interface
type Updater interface {
	SynchronizeSplits(till *int64) (*UpdateResult, error)
	LocalKill(splitName string, defaultTreatment string, changeNumber int64)
}

// UpdateResult encapsulates information regarding the split update performed
type UpdateResult struct {
	UpdatedSplits      []string
	ReferencedSegments []string
	NewChangeNumber    int64
}

type internalSplitSync struct {
	updateResult   *UpdateResult
	successfulSync bool
	attempt        int
}

// UpdaterImpl struct for split sync
type UpdaterImpl struct {
	splitStorage                storage.SplitStorage
	splitFetcher                service.SplitFetcher
	logger                      logging.LoggerInterface
	runtimeTelemetry            storage.TelemetryRuntimeProducer
	hcMonitor                   application.MonitorProducerInterface
	onDemandFetchBackoffBase    int64
	onDemandFetchBackoffMaxWait time.Duration
}

// NewSplitFetcher creates new split synchronizer for processing split updates
func NewSplitFetcher(
	splitStorage storage.SplitStorage,
	splitFetcher service.SplitFetcher,
	logger logging.LoggerInterface,
	runtimeTelemetry storage.TelemetryRuntimeProducer,
	hcMonitor application.MonitorProducerInterface,
) *UpdaterImpl {
	return &UpdaterImpl{
		splitStorage:                splitStorage,
		splitFetcher:                splitFetcher,
		logger:                      logger,
		runtimeTelemetry:            runtimeTelemetry,
		hcMonitor:                   hcMonitor,
		onDemandFetchBackoffBase:    onDemandFetchBackoffBase,
		onDemandFetchBackoffMaxWait: onDemandFetchBackoffMaxWait,
	}
}

func (s *UpdaterImpl) processUpdate(splits *dtos.SplitChangesDTO) {
	inactiveSplits := make([]dtos.SplitDTO, 0, len(splits.Splits))
	activeSplits := make([]dtos.SplitDTO, 0, len(splits.Splits))
	for idx := range splits.Splits {
		if splits.Splits[idx].Status == "ACTIVE" {
			activeSplits = append(activeSplits, splits.Splits[idx])
		} else {
			inactiveSplits = append(inactiveSplits, splits.Splits[idx])
		}
	}

	// Add/Update active splits
	s.splitStorage.Update(activeSplits, inactiveSplits, splits.Till)
}

// fetchUntil Hit endpoint, update storage and return when since==till.
func (s *UpdaterImpl) fetchUntil(fetchOptions *service.FetchOptions, till *int64) (*UpdateResult, error) {
	// just guessing sizes so the we don't realloc immediately
	segmentReferences := make([]string, 0, 10)
	updatedSplitNames := make([]string, 0, 10)
	var err error
	var currentSince int64

	for { // Fetch until since==till
		currentSince, _ = s.splitStorage.ChangeNumber()
		before := time.Now()
		var splits *dtos.SplitChangesDTO
		splits, err = s.splitFetcher.Fetch(currentSince, fetchOptions)
		if err != nil {
			if httpError, ok := err.(*dtos.HTTPError); ok {
				s.runtimeTelemetry.RecordSyncError(telemetry.SplitSync, httpError.Code)
			}
			break
		}
		currentSince = splits.Till
		s.runtimeTelemetry.RecordSyncLatency(telemetry.SplitSync, time.Since(before))
		s.processUpdate(splits)
		segmentReferences = appendSegmentNames(segmentReferences, splits)
		updatedSplitNames = appendSplitNames(updatedSplitNames, splits)
		if currentSince == splits.Since {
			s.runtimeTelemetry.RecordSuccessfulSync(telemetry.SplitSync, time.Now().UTC())
			break
		}
	}
	return &UpdateResult{
		UpdatedSplits:      common.DedupeStringSlice(updatedSplitNames),
		ReferencedSegments: common.DedupeStringSlice(segmentReferences),
		NewChangeNumber:    currentSince,
	}, err
}

// attemptSplitSync Hit endpoint, update storage and return True if sync is complete.
func (s *UpdaterImpl) attemptSplitSync(fetchOptions *service.FetchOptions, till *int64) (internalSplitSync, error) {
	internalBackoff := backoff.New(s.onDemandFetchBackoffBase, s.onDemandFetchBackoffMaxWait)
	remainingAttempts := onDemandFetchBackoffMaxRetries
	for {
		remainingAttempts = remainingAttempts - 1
		updateResult, err := s.fetchUntil(fetchOptions, till) // what we should do with err
		if err != nil || remainingAttempts <= 0 {
			return internalSplitSync{updateResult: updateResult, successfulSync: false, attempt: remainingAttempts}, err
		}
		if till == nil || *till <= updateResult.NewChangeNumber {
			return internalSplitSync{updateResult: updateResult, successfulSync: true, attempt: remainingAttempts}, nil
		}
		howLong := internalBackoff.Next()
		time.Sleep(howLong)
	}
}

func (s *UpdaterImpl) SynchronizeSplits(till *int64) (*UpdateResult, error) {
	s.hcMonitor.NotifyEvent(application.Splits)
	currentSince, _ := s.splitStorage.ChangeNumber()
	if till != nil && *till < currentSince { // the passed till is less than change_number, no need to perform updates
		return &UpdateResult{}, nil
	}

	fetchOptions := service.NewFetchOptions(true, nil)
	internalSyncResult, err := s.attemptSplitSync(&fetchOptions, till)
	attempts := onDemandFetchBackoffMaxRetries - internalSyncResult.attempt
	if err != nil {
		return internalSyncResult.updateResult, err
	}
	if internalSyncResult.successfulSync {
		s.logger.Debug(fmt.Sprintf("Refresh completed in %d attempts.", attempts))
		return internalSyncResult.updateResult, nil
	}
	withCDNBypass := service.NewFetchOptions(true, &internalSyncResult.updateResult.NewChangeNumber) // Set flag for bypassing CDN
	internalSyncResultCDNBypass, err := s.attemptSplitSync(&withCDNBypass, till)
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

func appendSplitNames(dst []string, splits *dtos.SplitChangesDTO) []string {
	for idx := range splits.Splits {
		dst = append(dst, splits.Splits[idx].Name)
	}
	return dst
}

func appendSegmentNames(dst []string, splits *dtos.SplitChangesDTO) []string {
	for _, split := range splits.Splits {
		for _, cond := range split.Conditions {
			for _, matcher := range cond.MatcherGroup.Matchers {
				if matcher.MatcherType == matcherTypeInSegment && matcher.UserDefinedSegment != nil {
					dst = append(dst, matcher.UserDefinedSegment.SegmentName)
				}
			}
		}
	}
	return dst
}

// LocalKill marks a spit as killed in local storage
func (s *UpdaterImpl) LocalKill(splitName string, defaultTreatment string, changeNumber int64) {
	s.splitStorage.KillLocally(splitName, defaultTreatment, changeNumber)
}
