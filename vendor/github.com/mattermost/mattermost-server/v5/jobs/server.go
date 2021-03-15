// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package jobs

import (
	"github.com/mattermost/mattermost-server/v5/einterfaces"
	ejobs "github.com/mattermost/mattermost-server/v5/einterfaces/jobs"
	tjobs "github.com/mattermost/mattermost-server/v5/jobs/interfaces"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/services/configservice"
	"github.com/mattermost/mattermost-server/v5/store"
)

type JobServer struct {
	ConfigService configservice.ConfigService
	Store         store.Store
	metrics       einterfaces.MetricsInterface
	Workers       *Workers
	Schedulers    *Schedulers

	DataRetentionJob        ejobs.DataRetentionJobInterface
	MessageExportJob        ejobs.MessageExportJobInterface
	ElasticsearchAggregator ejobs.ElasticsearchAggregatorInterface
	ElasticsearchIndexer    tjobs.IndexerJobInterface
	LdapSync                ejobs.LdapSyncInterface
	Migrations              tjobs.MigrationsJobInterface
	Plugins                 tjobs.PluginsJobInterface
	BleveIndexer            tjobs.IndexerJobInterface
	ExpiryNotify            tjobs.ExpiryNotifyJobInterface
	ProductNotices          tjobs.ProductNoticesJobInterface
	ActiveUsers             tjobs.ActiveUsersJobInterface
	ImportProcess           tjobs.ImportProcessInterface
	ImportDelete            tjobs.ImportDeleteInterface
	ExportProcess           tjobs.ExportProcessInterface
	ExportDelete            tjobs.ExportDeleteInterface
	Cloud                   ejobs.CloudJobInterface
}

func NewJobServer(configService configservice.ConfigService, store store.Store, metrics einterfaces.MetricsInterface) *JobServer {
	return &JobServer{
		ConfigService: configService,
		Store:         store,
		metrics:       metrics,
	}
}

func (srv *JobServer) Config() *model.Config {
	return srv.ConfigService.Config()
}

func (srv *JobServer) StartWorkers() {
	srv.Workers = srv.Workers.Start()
}

func (srv *JobServer) StartSchedulers() {
	srv.Schedulers = srv.Schedulers.Start()
}

func (srv *JobServer) StopWorkers() {
	if srv.Workers != nil {
		srv.Workers.Stop()
	}
}

func (srv *JobServer) StopSchedulers() {
	if srv.Schedulers != nil {
		srv.Schedulers.Stop()
	}
}
