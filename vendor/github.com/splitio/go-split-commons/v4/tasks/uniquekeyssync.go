package tasks

import (
	"github.com/splitio/go-split-commons/v4/provisional/strategy"
	"github.com/splitio/go-split-commons/v4/telemetry"
	"github.com/splitio/go-toolkit/v5/asynctask"
	"github.com/splitio/go-toolkit/v5/logging"
)

// NewRecordUniqueKeysTask constructor
func NewRecordUniqueKeysTask(
	recorder telemetry.TelemetrySynchronizer,
	uniqueTracker strategy.UniqueKeysTracker,
	period int,
	logger logging.LoggerInterface,
) *asynctask.AsyncTask {
	record := func(logger logging.LoggerInterface) error {
		return recorder.SynchronizeUniqueKeys(uniqueTracker.PopAll())
	}

	onStop := func(logger logging.LoggerInterface) {
		recorder.SynchronizeUniqueKeys(uniqueTracker.PopAll())
	}

	return asynctask.NewAsyncTask("SubmitUniqueKeys", record, period, nil, onStop, logger)
}
