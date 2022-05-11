// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package jobs

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

type SimpleWorker struct {
	name      string
	stop      chan bool
	stopped   chan bool
	jobs      chan model.Job
	jobServer *JobServer
	execute   func(job *model.Job) error
	isEnabled func(cfg *model.Config) bool
}

func NewSimpleWorker(name string, jobServer *JobServer, execute func(job *model.Job) error, isEnabled func(cfg *model.Config) bool) *SimpleWorker {
	worker := SimpleWorker{
		name:      name,
		stop:      make(chan bool, 1),
		stopped:   make(chan bool, 1),
		jobs:      make(chan model.Job),
		jobServer: jobServer,
		execute:   execute,
		isEnabled: isEnabled,
	}
	return &worker
}

func (worker *SimpleWorker) Run() {
	mlog.Debug("Worker started", mlog.String("worker", worker.name))

	defer func() {
		mlog.Debug("Worker finished", mlog.String("worker", worker.name))
		worker.stopped <- true
	}()

	for {
		select {
		case <-worker.stop:
			mlog.Debug("Worker received stop signal", mlog.String("worker", worker.name))
			return
		case job := <-worker.jobs:
			mlog.Debug("Worker received a new candidate job.", mlog.String("worker", worker.name))
			worker.DoJob(&job)
		}
	}
}

func (worker *SimpleWorker) Stop() {
	mlog.Debug("Worker stopping", mlog.String("worker", worker.name))
	worker.stop <- true
	<-worker.stopped
}

func (worker *SimpleWorker) JobChannel() chan<- model.Job {
	return worker.jobs
}

func (worker *SimpleWorker) IsEnabled(cfg *model.Config) bool {
	return worker.isEnabled(cfg)
}

func (worker *SimpleWorker) DoJob(job *model.Job) {
	if claimed, err := worker.jobServer.ClaimJob(job); err != nil {
		mlog.Warn("SimpleWorker experienced an error while trying to claim job",
			mlog.String("worker", worker.name),
			mlog.String("job_id", job.Id),
			mlog.Err(err))
		return
	} else if !claimed {
		return
	}

	err := worker.execute(job)
	if err != nil {
		mlog.Error("SimpleWorker: Failed to get active user count", mlog.String("worker", worker.name), mlog.String("job_id", job.Id), mlog.Err(err))
		worker.setJobError(job, model.NewAppError("DoJob", "app.user.get_total_users_count.app_error", nil, err.Error(), http.StatusInternalServerError))
		return
	}

	mlog.Info("SimpleWorker: Job is complete", mlog.String("worker", worker.name), mlog.String("job_id", job.Id))
	worker.setJobSuccess(job)
}

func (worker *SimpleWorker) setJobSuccess(job *model.Job) {
	if err := worker.jobServer.SetJobSuccess(job); err != nil {
		mlog.Error("SimpleWorker: Failed to set success for job", mlog.String("worker", worker.name), mlog.String("job_id", job.Id), mlog.String("error", err.Error()))
		worker.setJobError(job, err)
	}
}

func (worker *SimpleWorker) setJobError(job *model.Job, appError *model.AppError) {
	if err := worker.jobServer.SetJobError(job, appError); err != nil {
		mlog.Error("SimpleWorker: Failed to set job error", mlog.String("worker", worker.name), mlog.String("job_id", job.Id), mlog.Err(err))
	}
}
