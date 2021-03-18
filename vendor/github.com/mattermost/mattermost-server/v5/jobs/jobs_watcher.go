// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package jobs

import (
	"math/rand"
	"time"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
)

// Default polling interval for jobs termination.
// (Defining as `var` rather than `const` allows tests to lower the interval.)
var DefaultWatcherPollingInterval = 15000

type Watcher struct {
	srv     *JobServer
	workers *Workers

	stop            chan struct{}
	stopped         chan struct{}
	pollingInterval int
}

func (srv *JobServer) MakeWatcher(workers *Workers, pollingInterval int) *Watcher {
	return &Watcher{
		stop:            make(chan struct{}),
		stopped:         make(chan struct{}),
		pollingInterval: pollingInterval,
		workers:         workers,
		srv:             srv,
	}
}

func (watcher *Watcher) Start() {
	mlog.Debug("Watcher Started")

	// Delay for some random number of milliseconds before starting to ensure that multiple
	// instances of the jobserver  don't poll at a time too close to each other.
	rand.Seed(time.Now().UTC().UnixNano())
	<-time.After(time.Duration(rand.Intn(watcher.pollingInterval)) * time.Millisecond)

	defer func() {
		mlog.Debug("Watcher Finished")
		close(watcher.stopped)
	}()

	for {
		select {
		case <-watcher.stop:
			mlog.Debug("Watcher: Received stop signal")
			return
		case <-time.After(time.Duration(watcher.pollingInterval) * time.Millisecond):
			watcher.PollAndNotify()
		}
	}
}

func (watcher *Watcher) Stop() {
	mlog.Debug("Watcher Stopping")
	close(watcher.stop)
	<-watcher.stopped
}

func (watcher *Watcher) PollAndNotify() {
	jobs, err := watcher.srv.Store.Job().GetAllByStatus(model.JOB_STATUS_PENDING)
	if err != nil {
		mlog.Error("Error occurred getting all pending statuses.", mlog.Err(err))
		return
	}

	for _, job := range jobs {
		if job.Type == model.JOB_TYPE_DATA_RETENTION {
			if watcher.workers.DataRetention != nil {
				select {
				case watcher.workers.DataRetention.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_MESSAGE_EXPORT {
			if watcher.workers.MessageExport != nil {
				select {
				case watcher.workers.MessageExport.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_ELASTICSEARCH_POST_INDEXING {
			if watcher.workers.ElasticsearchIndexing != nil {
				select {
				case watcher.workers.ElasticsearchIndexing.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_ELASTICSEARCH_POST_AGGREGATION {
			if watcher.workers.ElasticsearchAggregation != nil {
				select {
				case watcher.workers.ElasticsearchAggregation.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_BLEVE_POST_INDEXING {
			if watcher.workers.BleveIndexing != nil {
				select {
				case watcher.workers.BleveIndexing.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_LDAP_SYNC {
			if watcher.workers.LdapSync != nil {
				select {
				case watcher.workers.LdapSync.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_MIGRATIONS {
			if watcher.workers.Migrations != nil {
				select {
				case watcher.workers.Migrations.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_PLUGINS {
			if watcher.workers.Plugins != nil {
				select {
				case watcher.workers.Plugins.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_EXPIRY_NOTIFY {
			if watcher.workers.ExpiryNotify != nil {
				select {
				case watcher.workers.ExpiryNotify.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_PRODUCT_NOTICES {
			if watcher.workers.ProductNotices != nil {
				select {
				case watcher.workers.ProductNotices.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_ACTIVE_USERS {
			if watcher.workers.ActiveUsers != nil {
				select {
				case watcher.workers.ActiveUsers.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_IMPORT_PROCESS {
			if watcher.workers.ImportProcess != nil {
				select {
				case watcher.workers.ImportProcess.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_IMPORT_DELETE {
			if watcher.workers.ImportDelete != nil {
				select {
				case watcher.workers.ImportDelete.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_EXPORT_PROCESS {
			if watcher.workers.ExportProcess != nil {
				select {
				case watcher.workers.ExportProcess.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_EXPORT_DELETE {
			if watcher.workers.ExportDelete != nil {
				select {
				case watcher.workers.ExportDelete.JobChannel() <- *job:
				default:
				}
			}
		} else if job.Type == model.JOB_TYPE_CLOUD {
			if watcher.workers.Cloud != nil {
				select {
				case watcher.workers.Cloud.JobChannel() <- *job:
				default:
				}
			}
		}
	}
}
