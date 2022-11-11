// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package last_accessible_post

import (
	"github.com/mattermost/mattermost-server/v6/jobs"
	"github.com/mattermost/mattermost-server/v6/model"
)

const (
	JobName = "LastAccessiblePost"
)

type AppIface interface {
	ComputeLastAccessiblePostTime() error
}

func MakeWorker(jobServer *jobs.JobServer, license *model.License, app AppIface) model.Worker {
	isEnabled := func(_ *model.Config) bool {
		return license != nil && license.Features != nil && *license.Features.Cloud
	}
	execute := func(_ *model.Job) error {
		return app.ComputeLastAccessiblePostTime()
	}
	worker := jobs.NewSimpleWorker(JobName, jobServer, execute, isEnabled)
	return worker
}
