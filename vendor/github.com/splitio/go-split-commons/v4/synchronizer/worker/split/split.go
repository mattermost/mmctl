package split

import (
	"time"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/healthcheck/application"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-split-commons/v4/telemetry"
	"github.com/splitio/go-toolkit/v5/common"
	"github.com/splitio/go-toolkit/v5/logging"
)

const (
	matcherTypeInSegment = "IN_SEGMENT"
)

// Updater interface
type Updater interface {
	SynchronizeSplits(till *int64, requestNoCache bool) (*UpdateResult, error)
	LocalKill(splitName string, defaultTreatment string, changeNumber int64)
}

// UpdateResult encapsulates information regarding the split update performed
type UpdateResult struct {
	UpdatedSplits      []string
	ReferencedSegments []string
	NewChangeNumber    int64
}

// UpdaterImpl struct for split sync
type UpdaterImpl struct {
	splitStorage     storage.SplitStorage
	splitFetcher     service.SplitFetcher
	logger           logging.LoggerInterface
	runtimeTelemetry storage.TelemetryRuntimeProducer
	hcMonitor        application.MonitorProducerInterface
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
		splitStorage:     splitStorage,
		splitFetcher:     splitFetcher,
		logger:           logger,
		runtimeTelemetry: runtimeTelemetry,
		hcMonitor:        hcMonitor,
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

// SynchronizeSplits syncs splits
func (s *UpdaterImpl) SynchronizeSplits(till *int64, requestNoCache bool) (*UpdateResult, error) {
	// @TODO: add delays

	// just guessing sizes so the we don't realloc immediately
	segmentReferences := make([]string, 0, 10)
	updatedSplitNames := make([]string, 0, 10)
	var newCN int64

	s.hcMonitor.NotifyEvent(application.Splits)
	var err error
	for {
		changeNumber, _ := s.splitStorage.ChangeNumber()
		if changeNumber == 0 {
			changeNumber = -1
		}
		if till != nil && *till < changeNumber {
			break
		}

		before := time.Now()
		var splits *dtos.SplitChangesDTO
		splits, err = s.splitFetcher.Fetch(changeNumber, requestNoCache)
		if err != nil {
			if httpError, ok := err.(*dtos.HTTPError); ok {
				s.runtimeTelemetry.RecordSyncError(telemetry.SplitSync, httpError.Code)
			}
			break
		}
		newCN = splits.Till
		s.runtimeTelemetry.RecordSyncLatency(telemetry.SplitSync, time.Since(before))
		s.processUpdate(splits)
		segmentReferences = appendSegmentNames(segmentReferences, splits)
		updatedSplitNames = appendSplitNames(updatedSplitNames, splits)
		if splits.Till == splits.Since || (till != nil && splits.Till >= *till) {
			s.runtimeTelemetry.RecordSuccessfulSync(telemetry.SplitSync, time.Now().UTC())
			break
		}
	}

	return &UpdateResult{
		UpdatedSplits:      common.DedupeStringSlice(updatedSplitNames),
		ReferencedSegments: common.DedupeStringSlice(segmentReferences),
		NewChangeNumber:    newCN,
	}, err
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
