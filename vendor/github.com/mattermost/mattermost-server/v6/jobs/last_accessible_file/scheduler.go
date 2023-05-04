// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package last_accessible_file

import (
	"strconv"
	"time"

	"github.com/mattermost/mattermost-server/v6/jobs"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

const schedFreq = 2 * time.Hour

func MakeScheduler(jobServer *jobs.JobServer, license *model.License) model.Scheduler {
	isEnabled := func(cfg *model.Config) bool {
		enabled := license != nil && *license.Features.Cloud
		mlog.Debug("Scheduler: isEnabled: "+strconv.FormatBool(enabled), mlog.String("scheduler", model.JobTypeLastAccessibleFile))
		return enabled
	}
	return jobs.NewPeriodicScheduler(jobServer, model.JobTypeLastAccessibleFile, schedFreq, isEnabled)
}
