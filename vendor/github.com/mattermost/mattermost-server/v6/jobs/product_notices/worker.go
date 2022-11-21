// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package product_notices

import (
	"github.com/mattermost/mattermost-server/v6/jobs"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

const jobName = "ProductNotices"

type AppIface interface {
	UpdateProductNotices() *model.AppError
}

func MakeWorker(jobServer *jobs.JobServer, app AppIface) model.Worker {
	isEnabled := func(cfg *model.Config) bool {
		return *cfg.AnnouncementSettings.AdminNoticesEnabled || *cfg.AnnouncementSettings.UserNoticesEnabled
	}
	execute := func(job *model.Job) error {
		if err := app.UpdateProductNotices(); err != nil {
			mlog.Error("Worker: Failed to fetch product notices", mlog.String("worker", model.JobTypeProductNotices), mlog.String("job_id", job.Id), mlog.Err(err))
			return err
		}
		return nil
	}
	worker := jobs.NewSimpleWorker(jobName, jobServer, execute, isEnabled)
	return worker
}
