package tasks

import (
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/impressionscount"
	"github.com/splitio/go-toolkit/v5/asynctask"
	"github.com/splitio/go-toolkit/v5/logging"
)

// NewRecordImpressionsCountTask creates a new impressionsCount recording task
func NewRecordImpressionsCountTask(
	recorder impressionscount.ImpressionsCountRecorder,
	logger logging.LoggerInterface,
	period int,
) *asynctask.AsyncTask {
	record := func(logger logging.LoggerInterface) error {
		return recorder.SynchronizeImpressionsCount()
	}

	onStop := func(logger logging.LoggerInterface) {
		recorder.SynchronizeImpressionsCount()
	}

	return asynctask.NewAsyncTask("SubmitImpressionsCount", record, period, nil, onStop, logger)
}
