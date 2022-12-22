package storage

import (
	"time"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/datastructures/set"
)

// SplitStorageProducer should be implemented by structs that offer writing splits in storage
type SplitStorageProducer interface {
	ChangeNumber() (int64, error)
	Update(toAdd []dtos.SplitDTO, toRemove []dtos.SplitDTO, changeNumber int64)
	KillLocally(splitName string, defaultTreatment string, changeNumber int64)
	SetChangeNumber(changeNumber int64) error
}

// SplitStorageConsumer should be implemented by structs that offer reading splits from storage
type SplitStorageConsumer interface {
	ChangeNumber() (int64, error)
	All() []dtos.SplitDTO
	FetchMany(splitNames []string) map[string]*dtos.SplitDTO
	SegmentNames() *set.ThreadUnsafeSet // Not in Spec
	Split(splitName string) *dtos.SplitDTO
	SplitNames() []string
	TrafficTypeExists(trafficType string) bool
}

// SegmentStorageProducer interface should be implemented by all structs that offer writing segments
type SegmentStorageProducer interface {
	Update(name string, toAdd *set.ThreadUnsafeSet, toRemove *set.ThreadUnsafeSet, changeNumber int64) error
	SetChangeNumber(segmentName string, till int64) error
}

// SegmentStorageConsumer interface should be implemented by all structs that ofer reading segments
type SegmentStorageConsumer interface {
	ChangeNumber(segmentName string) (int64, error)
	Keys(segmentName string) *set.ThreadUnsafeSet
	SegmentContainsKey(segmentName string, key string) (bool, error)
	SegmentKeysCount() int64
}

// ImpressionStorageProducer interface should be impemented by structs that accept incoming impressions
type ImpressionStorageProducer interface {
	LogImpressions(impressions []dtos.Impression) error
}

// DataDropper interface is used by dependants who need to drop data from a collection
type DataDropper interface {
	Drop(size int64) error
}

// ImpressionMultiSdkConsumer defines the methods required to consume impressions
// from a stored shared by many sdks
type ImpressionMultiSdkConsumer interface {
	Count() int64
	PopNRaw(int64) ([]string, int64, error)
	PopNWithMetadata(n int64) ([]dtos.ImpressionQueueObject, error)
}

// EventMultiSdkConsumer defines the methods required to consume events
// from a stored shared by many sdks
type EventMultiSdkConsumer interface {
	Count() int64
	PopNRaw(int64) ([]string, int64, error)
	PopNWithMetadata(n int64) ([]dtos.QueueStoredEventDTO, error)
}

// UniqueKeysMultiSdkConsumer defines the methods required to consume unique keys
// from a stored shared by many sdks
type UniqueKeysMultiSdkConsumer interface {
	Count() int64
	PopNRaw(int64) ([]string, int64, error)
}

// ImpressionStorageConsumer interface should be implemented by structs that offer popping impressions
type ImpressionStorageConsumer interface {
	Empty() bool
	PopN(n int64) ([]dtos.Impression, error)
}

// EventStorageProducer interface should be implemented by structs that accept incoming events
type EventStorageProducer interface {
	Push(event dtos.EventDTO, size int) error
}

// EventStorageConsumer interface should be implemented by structs that offer popping impressions
type EventStorageConsumer interface {
	Empty() bool
	PopN(n int64) ([]dtos.EventDTO, error)
}

// TelemetryStorageProducer interface should be implemented by struct that accepts incoming telemetry
type TelemetryStorageProducer interface {
	TelemetryConfigProducer
	TelemetryEvaluationProducer
	TelemetryRuntimeProducer
}

// TelemetryRedisProducer interface redis
type TelemetryRedisProducer interface {
	TelemetryConfigProducer
	TelemetryEvaluationProducer
}

// TelemetryConfigProducer interface for config data
type TelemetryConfigProducer interface {
	RecordConfigData(configData dtos.Config) error
	RecordNonReadyUsage()
	RecordBURTimeout()
	RecordUniqueKeys(uniques dtos.Uniques) error
}

// TelemetryEvaluationProducer for evaluation
type TelemetryEvaluationProducer interface {
	RecordLatency(method string, latency time.Duration)
	RecordException(method string)
}

// TelemetryRuntimeProducer for runtime stats
type TelemetryRuntimeProducer interface {
	AddTag(tag string)
	RecordImpressionsStats(dataType int, count int64)
	RecordEventsStats(dataType int, count int64)
	RecordSuccessfulSync(resource int, when time.Time)
	RecordSyncError(resource int, status int)
	RecordSyncLatency(resource int, latency time.Duration)
	RecordAuthRejections()
	RecordTokenRefreshes()
	RecordStreamingEvent(streamingEvent *dtos.StreamingEvent)
	RecordSessionLength(session int64)
}

// TelemetryStorageConsumer interface should be implemented by structs that offer popping telemetry
type TelemetryStorageConsumer interface {
	TelemetryConfigConsumer
	TelemetryEvaluationConsumer
	TelemetryRuntimeConsumer
}

// TelemetryConfigConsumer interface for config data
type TelemetryConfigConsumer interface {
	GetNonReadyUsages() int64
	GetBURTimeouts() int64
}

// TelemetryEvaluationConsumer for evaluation
type TelemetryEvaluationConsumer interface {
	PopLatencies() dtos.MethodLatencies
	PopExceptions() dtos.MethodExceptions
}

// TelemetryRuntimeConsumer for runtime stats
type TelemetryRuntimeConsumer interface {
	GetImpressionsStats(dataType int) int64
	GetEventsStats(dataType int) int64
	GetLastSynchronization() dtos.LastSynchronization
	PopHTTPErrors() dtos.HTTPErrors
	PopHTTPLatencies() dtos.HTTPLatencies
	PopAuthRejections() int64
	PopTokenRefreshes() int64
	PopStreamingEvents() []dtos.StreamingEvent
	PopTags() []string
	GetSessionLength() int64
}

// TelemetryPeeker interface
type TelemetryPeeker interface {
	PeekHTTPLatencies(resource int) []int64
	PeekHTTPErrors(resource int) map[int]int
}

// ImpressionsCountProducer interface
type ImpressionsCountProducer interface {
	RecordImpressionsCount(impressions dtos.ImpressionsCountDTO) error
}

// ImpressionsCountProducer interface
type ImpressionsCountConsumer interface {
	GetImpressionsCount() (*dtos.ImpressionsCountDTO, error)
}

type ImpressionsCountStorage interface {
	ImpressionsCountConsumer
	ImpressionsCountProducer
}

// --- Wide Interfaces

// SplitStorage wraps consumer & producer interfaces
// Note: Since go's interface composition does not (yet) support interface method overlap,
// extracting a common subset (.ChangeNumber()), embedding it in Both consumer & Producer,
// and then having a wide interface that embeds both (diamond composition), results in a compilation error.
// The only workaround so far is to explicitly define all the methods that make up the composed interface
type SplitStorage interface {
	ChangeNumber() (int64, error)
	Update(toAdd []dtos.SplitDTO, toRemove []dtos.SplitDTO, changeNumber int64)
	KillLocally(splitName string, defaultTreatment string, changeNumber int64)
	SetChangeNumber(changeNumber int64) error
	All() []dtos.SplitDTO
	FetchMany(splitNames []string) map[string]*dtos.SplitDTO
	SegmentNames() *set.ThreadUnsafeSet // Not in Spec
	Split(splitName string) *dtos.SplitDTO
	SplitNames() []string
	TrafficTypeExists(trafficType string) bool
}

// SegmentStorage wraps consumer and producer interfaces
type SegmentStorage interface {
	SegmentStorageProducer
	SegmentStorageConsumer
}

// ImpressionStorage wraps consumer & producer interfaces
type ImpressionStorage interface {
	ImpressionStorageConsumer
	ImpressionStorageProducer
	ImpressionMultiSdkConsumer
}

// EventsStorage wraps consumer and producer interfaces
type EventsStorage interface {
	EventStorageConsumer
	EventStorageProducer
}

// TelemetryStorage wraps consumer and producer interfaces
type TelemetryStorage interface {
	TelemetryStorageConsumer
	TelemetryStorageProducer
}

// Filter interfaces
type Filter interface {
	Add(data string)
	Contains(data string) bool
	Clear()
}
