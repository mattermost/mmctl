// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package jobs

import (
	"sync"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/services/configservice"
)

type Workers struct {
	startOnce     sync.Once
	ConfigService configservice.ConfigService
	Watcher       *Watcher

	DataRetention            model.Worker
	MessageExport            model.Worker
	ElasticsearchIndexing    model.Worker
	ElasticsearchAggregation model.Worker
	LdapSync                 model.Worker
	Migrations               model.Worker
	Plugins                  model.Worker
	BleveIndexing            model.Worker
	ExpiryNotify             model.Worker
	ProductNotices           model.Worker
	ActiveUsers              model.Worker
	ImportProcess            model.Worker
	Cloud                    model.Worker

	listenerId string
}

func (srv *JobServer) InitWorkers() *Workers {
	workers := &Workers{
		ConfigService: srv.ConfigService,
	}
	workers.Watcher = srv.MakeWatcher(workers, DEFAULT_WATCHER_POLLING_INTERVAL)

	if srv.DataRetentionJob != nil {
		workers.DataRetention = srv.DataRetentionJob.MakeWorker()
	}

	if srv.MessageExportJob != nil {
		workers.MessageExport = srv.MessageExportJob.MakeWorker()
	}

	if elasticsearchIndexerInterface := srv.ElasticsearchIndexer; elasticsearchIndexerInterface != nil {
		workers.ElasticsearchIndexing = elasticsearchIndexerInterface.MakeWorker()
	}

	if elasticsearchAggregatorInterface := srv.ElasticsearchAggregator; elasticsearchAggregatorInterface != nil {
		workers.ElasticsearchAggregation = elasticsearchAggregatorInterface.MakeWorker()
	}

	if ldapSyncInterface := srv.LdapSync; ldapSyncInterface != nil {
		workers.LdapSync = ldapSyncInterface.MakeWorker()
	}

	if migrationsInterface := srv.Migrations; migrationsInterface != nil {
		workers.Migrations = migrationsInterface.MakeWorker()
	}

	if pluginsInterface := srv.Plugins; pluginsInterface != nil {
		workers.Plugins = pluginsInterface.MakeWorker()
	}

	if bleveIndexerInterface := srv.BleveIndexer; bleveIndexerInterface != nil {
		workers.BleveIndexing = bleveIndexerInterface.MakeWorker()
	}

	if expiryNotifyInterface := srv.ExpiryNotify; expiryNotifyInterface != nil {
		workers.ExpiryNotify = expiryNotifyInterface.MakeWorker()
	}

	if activeUsersInterface := srv.ActiveUsers; activeUsersInterface != nil {
		workers.ActiveUsers = activeUsersInterface.MakeWorker()
	}

	if productNoticesInterface := srv.ProductNotices; productNoticesInterface != nil {
		workers.ProductNotices = productNoticesInterface.MakeWorker()
	}

	if importProcessInterface := srv.ImportProcess; importProcessInterface != nil {
		workers.ImportProcess = importProcessInterface.MakeWorker()
	}

	if cloudInterface := srv.Cloud; cloudInterface != nil {
		workers.Cloud = cloudInterface.MakeWorker()
	}

	return workers
}

func (workers *Workers) Start() *Workers {
	mlog.Info("Starting workers")

	workers.startOnce.Do(func() {
		if workers.DataRetention != nil && (*workers.ConfigService.Config().DataRetentionSettings.EnableMessageDeletion || *workers.ConfigService.Config().DataRetentionSettings.EnableFileDeletion) {
			go workers.DataRetention.Run()
		}

		if workers.MessageExport != nil && *workers.ConfigService.Config().MessageExportSettings.EnableExport {
			go workers.MessageExport.Run()
		}

		if workers.ElasticsearchIndexing != nil && *workers.ConfigService.Config().ElasticsearchSettings.EnableIndexing {
			go workers.ElasticsearchIndexing.Run()
		}

		if workers.ElasticsearchAggregation != nil && *workers.ConfigService.Config().ElasticsearchSettings.EnableIndexing {
			go workers.ElasticsearchAggregation.Run()
		}

		if workers.LdapSync != nil && *workers.ConfigService.Config().LdapSettings.EnableSync {
			go workers.LdapSync.Run()
		}

		if workers.Migrations != nil {
			go workers.Migrations.Run()
		}

		if workers.Plugins != nil {
			go workers.Plugins.Run()
		}

		if workers.BleveIndexing != nil && *workers.ConfigService.Config().BleveSettings.EnableIndexing && *workers.ConfigService.Config().BleveSettings.IndexDir != "" {
			go workers.BleveIndexing.Run()
		}

		if workers.ExpiryNotify != nil {
			go workers.ExpiryNotify.Run()
		}

		if workers.ActiveUsers != nil {
			go workers.ActiveUsers.Run()
		}

		if workers.ProductNotices != nil {
			go workers.ProductNotices.Run()
		}

		if workers.ImportProcess != nil {
			go workers.ImportProcess.Run()
		}

		if workers.Cloud != nil {
			go workers.Cloud.Run()
		}

		go workers.Watcher.Start()
	})

	workers.listenerId = workers.ConfigService.AddConfigListener(workers.handleConfigChange)

	return workers
}

func (workers *Workers) handleConfigChange(oldConfig *model.Config, newConfig *model.Config) {
	mlog.Debug("Workers received config change.")

	if workers.DataRetention != nil {
		if (!*oldConfig.DataRetentionSettings.EnableMessageDeletion && !*oldConfig.DataRetentionSettings.EnableFileDeletion) && (*newConfig.DataRetentionSettings.EnableMessageDeletion || *newConfig.DataRetentionSettings.EnableFileDeletion) {
			go workers.DataRetention.Run()
		} else if (*oldConfig.DataRetentionSettings.EnableMessageDeletion || *oldConfig.DataRetentionSettings.EnableFileDeletion) && (!*newConfig.DataRetentionSettings.EnableMessageDeletion && !*newConfig.DataRetentionSettings.EnableFileDeletion) {
			workers.DataRetention.Stop()
		}
	}

	if workers.MessageExport != nil {
		if !*oldConfig.MessageExportSettings.EnableExport && *newConfig.MessageExportSettings.EnableExport {
			go workers.MessageExport.Run()
		} else if *oldConfig.MessageExportSettings.EnableExport && !*newConfig.MessageExportSettings.EnableExport {
			workers.MessageExport.Stop()
		}
	}

	if workers.ElasticsearchIndexing != nil {
		if !*oldConfig.ElasticsearchSettings.EnableIndexing && *newConfig.ElasticsearchSettings.EnableIndexing {
			go workers.ElasticsearchIndexing.Run()
		} else if *oldConfig.ElasticsearchSettings.EnableIndexing && !*newConfig.ElasticsearchSettings.EnableIndexing {
			workers.ElasticsearchIndexing.Stop()
		}
	}

	if workers.ElasticsearchAggregation != nil {
		if !*oldConfig.ElasticsearchSettings.EnableIndexing && *newConfig.ElasticsearchSettings.EnableIndexing {
			go workers.ElasticsearchAggregation.Run()
		} else if *oldConfig.ElasticsearchSettings.EnableIndexing && !*newConfig.ElasticsearchSettings.EnableIndexing {
			workers.ElasticsearchAggregation.Stop()
		}
	}

	if workers.LdapSync != nil {
		if !*oldConfig.LdapSettings.EnableSync && *newConfig.LdapSettings.EnableSync {
			go workers.LdapSync.Run()
		} else if *oldConfig.LdapSettings.EnableSync && !*newConfig.LdapSettings.EnableSync {
			workers.LdapSync.Stop()
		}
	}

	if workers.BleveIndexing != nil {
		if !*oldConfig.BleveSettings.EnableIndexing && *newConfig.BleveSettings.EnableIndexing {
			go workers.BleveIndexing.Run()
		} else if *oldConfig.BleveSettings.EnableIndexing && !*newConfig.BleveSettings.EnableIndexing {
			workers.BleveIndexing.Stop()
		}
	}
}

func (workers *Workers) Stop() *Workers {
	workers.ConfigService.RemoveConfigListener(workers.listenerId)

	workers.Watcher.Stop()

	if workers.DataRetention != nil && (*workers.ConfigService.Config().DataRetentionSettings.EnableMessageDeletion || *workers.ConfigService.Config().DataRetentionSettings.EnableFileDeletion) {
		workers.DataRetention.Stop()
	}

	if workers.MessageExport != nil && *workers.ConfigService.Config().MessageExportSettings.EnableExport {
		workers.MessageExport.Stop()
	}

	if workers.ElasticsearchIndexing != nil && *workers.ConfigService.Config().ElasticsearchSettings.EnableIndexing {
		workers.ElasticsearchIndexing.Stop()
	}

	if workers.ElasticsearchAggregation != nil && *workers.ConfigService.Config().ElasticsearchSettings.EnableIndexing {
		workers.ElasticsearchAggregation.Stop()
	}

	if workers.LdapSync != nil && *workers.ConfigService.Config().LdapSettings.EnableSync {
		workers.LdapSync.Stop()
	}

	if workers.Migrations != nil {
		workers.Migrations.Stop()
	}

	if workers.Plugins != nil {
		workers.Plugins.Stop()
	}

	if workers.BleveIndexing != nil && *workers.ConfigService.Config().BleveSettings.EnableIndexing {
		workers.BleveIndexing.Stop()
	}

	if workers.ExpiryNotify != nil {
		workers.ExpiryNotify.Stop()
	}

	if workers.ActiveUsers != nil {
		workers.ActiveUsers.Stop()
	}

	if workers.ProductNotices != nil {
		workers.ProductNotices.Stop()
	}

	if workers.ImportProcess != nil {
		workers.ImportProcess.Stop()
	}

	if workers.Cloud != nil {
		workers.Cloud.Stop()
	}

	mlog.Info("Stopped workers")

	return workers
}
