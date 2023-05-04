package impressionscount

import (
	"time"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/provisional/strategy"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-split-commons/v4/telemetry"
	"github.com/splitio/go-toolkit/v5/logging"
)

// ImpressionsCountRecorder interface
type ImpressionsCountRecorder interface {
	SynchronizeImpressionsCount() error
}

// RecorderSingle struct for impressionsCount sync
type RecorderSingle struct {
	impressionsCounter *strategy.ImpressionsCounter
	impressionRecorder service.ImpressionsRecorder
	metadata           dtos.Metadata
	logger             logging.LoggerInterface
	runtimeTelemetry   storage.TelemetryRuntimeProducer
}

// NewRecorderSingle creates new impressionsCount synchronizer for posting impressionsCount
func NewRecorderSingle(
	impressionsCounter *strategy.ImpressionsCounter,
	impressionRecorder service.ImpressionsRecorder,
	metadata dtos.Metadata,
	logger logging.LoggerInterface,
	runtimeTelemetry storage.TelemetryRuntimeProducer,
) ImpressionsCountRecorder {
	return &RecorderSingle{
		impressionsCounter: impressionsCounter,
		impressionRecorder: impressionRecorder,
		metadata:           metadata,
		logger:             logger,
		runtimeTelemetry:   runtimeTelemetry,
	}
}

// SynchronizeImpressionsCount syncs imp counts
func (m *RecorderSingle) SynchronizeImpressionsCount() error {
	impressionsCount := m.impressionsCounter.PopAll()

	pf := impressionsCountMapper(impressionsCount)

	before := time.Now()
	err := m.impressionRecorder.RecordImpressionsCount(pf, m.metadata)
	if err != nil {
		if httpError, ok := err.(*dtos.HTTPError); ok {
			m.runtimeTelemetry.RecordSyncError(telemetry.ImpressionCountSync, httpError.Code)
		}
		return err
	}
	m.runtimeTelemetry.RecordSyncLatency(telemetry.ImpressionCountSync, time.Since(before))
	m.runtimeTelemetry.RecordSuccessfulSync(telemetry.ImpressionCountSync, time.Now().UTC())
	return nil
}

func impressionsCountMapper(impressionsCount map[strategy.Key]int64) dtos.ImpressionsCountDTO {
	impressionsInTimeFrame := make([]dtos.ImpressionsInTimeFrameDTO, 0)
	for key, count := range impressionsCount {
		impressionInTimeFrame := dtos.ImpressionsInTimeFrameDTO{
			FeatureName: key.FeatureName,
			RawCount:    count,
			TimeFrame:   key.TimeFrame,
		}
		impressionsInTimeFrame = append(impressionsInTimeFrame, impressionInTimeFrame)
	}

	pf := dtos.ImpressionsCountDTO{
		PerFeature: impressionsInTimeFrame,
	}

	return pf
}
