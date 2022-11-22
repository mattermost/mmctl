package push

// Borrowed synchronizer interface to break circular dependencies
type synchronizerInterface interface {
	SyncAll() error
	SynchronizeSplits(till *int64) error
	LocalKill(splitName string, defaultTreatment string, changeNumber int64)
	SynchronizeSegment(segmentName string, till *int64) error
	StartPeriodicFetching()
	StopPeriodicFetching()
	StartPeriodicDataRecording()
	StopPeriodicDataRecording()
}
