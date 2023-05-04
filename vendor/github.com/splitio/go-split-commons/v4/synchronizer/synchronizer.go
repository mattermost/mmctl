package synchronizer

import (
	"time"

	"github.com/splitio/go-split-commons/v4/conf"
	"github.com/splitio/go-split-commons/v4/healthcheck/application"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/event"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/impression"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/impressionscount"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/segment"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/split"
	"github.com/splitio/go-split-commons/v4/tasks"
	"github.com/splitio/go-split-commons/v4/telemetry"
	"github.com/splitio/go-toolkit/v5/asynctask"
	"github.com/splitio/go-toolkit/v5/logging"
)

// SplitTasks struct for tasks
type SplitTasks struct {
	SplitSyncTask            *asynctask.AsyncTask
	SegmentSyncTask          *asynctask.AsyncTask
	TelemetrySyncTask        tasks.Task
	ImpressionSyncTask       tasks.Task
	EventSyncTask            tasks.Task
	ImpressionsCountSyncTask tasks.Task
	UniqueKeysTask           tasks.Task
	CleanFilterTask          tasks.Task
	ImpsCountConsumerTask    tasks.Task
}

// Workers struct for workers
type Workers struct {
	SplitFetcher             split.Updater
	SegmentFetcher           segment.Updater
	TelemetryRecorder        telemetry.TelemetrySynchronizer
	ImpressionRecorder       impression.ImpressionRecorder
	EventRecorder            event.EventRecorder
	ImpressionsCountRecorder impressionscount.ImpressionsCountRecorder
}

// Synchronizer interface for syncing data to and from splits servers
type Synchronizer interface {
	SyncAll() error
	SynchronizeSplits(till *int64) error
	LocalKill(splitName string, defaultTreatment string, changeNumber int64)
	SynchronizeSegment(segmentName string, till *int64) error
	StartPeriodicFetching()
	StopPeriodicFetching()
	StartPeriodicDataRecording()
	StopPeriodicDataRecording()
	RefreshRates() (splits time.Duration, segments time.Duration)
}

// SynchronizerImpl implements Synchronizer
type SynchronizerImpl struct {
	splitTasks          SplitTasks
	workers             Workers
	logger              logging.LoggerInterface
	inMememoryFullQueue chan string
	impressionBulkSize  int64
	eventBulkSize       int64
	splitsRefreshRate   int
	segmentsRefreshRate int
	httpTiemoutSecs     int
}

// NewSynchronizer creates new SynchronizerImpl
func NewSynchronizer(
	confAdvanced conf.AdvancedConfig,
	splitTasks SplitTasks,
	workers Workers,
	logger logging.LoggerInterface,
	inMememoryFullQueue chan string,
	hcMonitor application.MonitorProducerInterface, // Deprecated: This is no longer used, left here only to avoid a breaking change
) Synchronizer {
	return &SynchronizerImpl{
		impressionBulkSize:  confAdvanced.ImpressionsBulkSize,
		eventBulkSize:       confAdvanced.EventsBulkSize,
		splitTasks:          splitTasks,
		workers:             workers,
		logger:              logger,
		inMememoryFullQueue: inMememoryFullQueue,
		splitsRefreshRate:   confAdvanced.SplitsRefreshRate,
		segmentsRefreshRate: confAdvanced.SegmentsRefreshRate,
		httpTiemoutSecs:     confAdvanced.HTTPTimeout,
	}
}

func (s *SynchronizerImpl) dataFlusher() {
	for {
		msg := <-s.inMememoryFullQueue
		switch msg {
		case "EVENTS_FULL":
			s.logger.Debug("FLUSHING storage queue")
			err := s.workers.EventRecorder.SynchronizeEvents(s.eventBulkSize)
			if err != nil {
				s.logger.Error("Error flushing storage queue", err)
			}
		case "IMPRESSIONS_FULL":
			s.logger.Debug("FLUSHING storage queue")
			err := s.workers.ImpressionRecorder.SynchronizeImpressions(s.impressionBulkSize)
			if err != nil {
				s.logger.Error("Error flushing storage queue", err)
			}
		}
	}
}

// SyncAll syncs splits and segments
func (s *SynchronizerImpl) SyncAll() error {
	_, err := s.workers.SplitFetcher.SynchronizeSplits(nil)
	if err != nil {
		return err
	}
	_, err = s.workers.SegmentFetcher.SynchronizeSegments()
	return err
}

// StartPeriodicFetching starts periodic fetchers tasks
func (s *SynchronizerImpl) StartPeriodicFetching() {
	if s.splitTasks.SplitSyncTask != nil {
		s.splitTasks.SplitSyncTask.Start()
	}
	if s.splitTasks.SegmentSyncTask != nil {
		s.splitTasks.SegmentSyncTask.Start()
	}
}

// StopPeriodicFetching stops periodic fetchers tasks
func (s *SynchronizerImpl) StopPeriodicFetching() {
	if s.splitTasks.SplitSyncTask != nil {
		s.splitTasks.SplitSyncTask.Stop(false)
	}
	if s.splitTasks.SegmentSyncTask != nil {
		s.splitTasks.SegmentSyncTask.Stop(true)
	}
}

// StartPeriodicDataRecording starts periodic recorders tasks
func (s *SynchronizerImpl) StartPeriodicDataRecording() {
	if s.inMememoryFullQueue != nil {
		go s.dataFlusher()
	}

	if s.splitTasks.ImpressionSyncTask != nil {
		s.splitTasks.ImpressionSyncTask.Start()
	}
	if s.splitTasks.TelemetrySyncTask != nil {
		s.splitTasks.TelemetrySyncTask.Start()
	}
	if s.splitTasks.EventSyncTask != nil {
		s.splitTasks.EventSyncTask.Start()
	}
	if s.splitTasks.ImpressionsCountSyncTask != nil {
		s.splitTasks.ImpressionsCountSyncTask.Start()
	}
	if s.splitTasks.UniqueKeysTask != nil {
		s.splitTasks.UniqueKeysTask.Start()
	}
	if s.splitTasks.CleanFilterTask != nil {
		s.splitTasks.CleanFilterTask.Start()
	}
	if s.splitTasks.ImpsCountConsumerTask != nil {
		s.splitTasks.ImpsCountConsumerTask.Start()
	}
}

// StopPeriodicDataRecording stops periodic recorders tasks
func (s *SynchronizerImpl) StopPeriodicDataRecording() {
	if s.splitTasks.ImpressionSyncTask != nil {
		s.splitTasks.ImpressionSyncTask.Stop(true)
	}
	if s.splitTasks.TelemetrySyncTask != nil {
		s.splitTasks.TelemetrySyncTask.Stop(true)
	}
	if s.splitTasks.EventSyncTask != nil {
		s.splitTasks.EventSyncTask.Stop(true)
	}
	if s.splitTasks.ImpressionsCountSyncTask != nil {
		s.splitTasks.ImpressionsCountSyncTask.Stop(true)
	}
	if s.splitTasks.UniqueKeysTask != nil {
		s.splitTasks.UniqueKeysTask.Stop(true)
	}
	if s.splitTasks.CleanFilterTask != nil {
		s.splitTasks.CleanFilterTask.Stop(true)
	}
	if s.splitTasks.ImpsCountConsumerTask != nil {
		s.splitTasks.ImpsCountConsumerTask.Stop(true)
	}
}

// SynchronizeSplits syncs splits
func (s *SynchronizerImpl) SynchronizeSplits(till *int64) error {
	result, err := s.workers.SplitFetcher.SynchronizeSplits(till)
	for _, segment := range s.filterCachedSegments(result.ReferencedSegments) {
		go s.SynchronizeSegment(segment, nil) // send segment to workerpool (queue is bypassed)
	}
	return err
}

// RefreshRates returns the refresh rates of the splits & segment tasks
func (s *SynchronizerImpl) RefreshRates() (splits time.Duration, segments time.Duration) {
	return time.Duration(s.splitsRefreshRate) * time.Second, time.Duration(s.segmentsRefreshRate) * time.Second
}

func (s *SynchronizerImpl) filterCachedSegments(segmentsReferenced []string) []string {
	toRet := make([]string, 0, len(segmentsReferenced))
	for _, name := range segmentsReferenced {
		if !s.workers.SegmentFetcher.IsSegmentCached(name) {
			toRet = append(toRet, name)
		}
	}
	return toRet
}

// LocalKill locally kills a split
func (s *SynchronizerImpl) LocalKill(splitName string, defaultTreatment string, changeNumber int64) {
	s.workers.SplitFetcher.LocalKill(splitName, defaultTreatment, changeNumber)
}

// SynchronizeSegment syncs segment
func (s *SynchronizerImpl) SynchronizeSegment(name string, till *int64) error {
	_, err := s.workers.SegmentFetcher.SynchronizeSegment(name, till)
	return err
}

var _ Synchronizer = &SynchronizerImpl{}
