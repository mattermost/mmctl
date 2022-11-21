package synchronizer

import (
	"time"

	"github.com/splitio/go-split-commons/v4/healthcheck/application"
	"github.com/splitio/go-split-commons/v4/service/api"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/split"
	"github.com/splitio/go-split-commons/v4/tasks"
	"github.com/splitio/go-toolkit/v5/logging"
)

// Local implements Local Synchronizer
type Local struct {
	splitTasks SplitTasks
	workers    Workers
	logger     logging.LoggerInterface
}

// NewLocal creates new Local
func NewLocal(period int, splitAPI *api.SplitAPI, splitStorage storage.SplitStorage, logger logging.LoggerInterface, runtimeTelemetry storage.TelemetryRuntimeProducer, hcMonitor application.MonitorProducerInterface) Synchronizer {
	workers := Workers{
		SplitFetcher: split.NewSplitFetcher(splitStorage, splitAPI.SplitFetcher, logger, runtimeTelemetry, hcMonitor),
	}
	return &Local{
		splitTasks: SplitTasks{
			SplitSyncTask: tasks.NewFetchSplitsTask(workers.SplitFetcher, period, logger),
		},
		workers: workers,
		logger:  logger,
	}
}

// SyncAll syncs splits and segments
func (s *Local) SyncAll() error {
	_, err := s.workers.SplitFetcher.SynchronizeSplits(nil)
	return err
}

// StartPeriodicFetching starts periodic fetchers tasks
func (s *Local) StartPeriodicFetching() {
	s.splitTasks.SplitSyncTask.Start()
}

// StopPeriodicFetching stops periodic fetchers tasks
func (s *Local) StopPeriodicFetching() {
	s.splitTasks.SplitSyncTask.Stop(false)
}

// StartPeriodicDataRecording starts periodic recorders tasks
func (s *Local) StartPeriodicDataRecording() {
}

// StopPeriodicDataRecording stops periodic recorders tasks
func (s *Local) StopPeriodicDataRecording() {
}

// RefreshRates returns anything
func (s *Local) RefreshRates() (time.Duration, time.Duration) {
	return 10 * time.Minute, 10 * time.Minute
}

// SynchronizeSplits syncs splits
func (s *Local) SynchronizeSplits(till *int64) error {
	_, err := s.workers.SplitFetcher.SynchronizeSplits(nil)
	return err
}

// SynchronizeSegment syncs segment
func (s *Local) SynchronizeSegment(name string, till *int64) error {
	return nil
}

// LocalKill does nothing
func (s *Local) LocalKill(splitName string, defaultTreatment string, changeNumber int64) {
}
