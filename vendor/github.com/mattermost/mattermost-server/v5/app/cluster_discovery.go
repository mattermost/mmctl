// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"time"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	DISCOVERY_SERVICE_WRITE_PING = 60 * time.Second
)

type ClusterDiscoveryService struct {
	model.ClusterDiscovery
	srv  *Server
	stop chan bool
}

func (s *Server) NewClusterDiscoveryService() *ClusterDiscoveryService {
	ds := &ClusterDiscoveryService{
		ClusterDiscovery: model.ClusterDiscovery{},
		srv:              s,
		stop:             make(chan bool),
	}

	return ds
}

func (a *App) NewClusterDiscoveryService() *ClusterDiscoveryService {
	return a.Srv().NewClusterDiscoveryService()
}

func (me *ClusterDiscoveryService) Start() {
	err := me.srv.Store.ClusterDiscovery().Cleanup()
	if err != nil {
		mlog.Error("ClusterDiscoveryService failed to cleanup the outdated cluster discovery information", mlog.Err(err))
	}

	exists, err := me.srv.Store.ClusterDiscovery().Exists(&me.ClusterDiscovery)
	if err != nil {
		mlog.Error("ClusterDiscoveryService failed to check if row exists", mlog.String("ClusterDiscovery", me.ClusterDiscovery.ToJson()), mlog.Err(err))
	} else {
		if exists {
			if _, err := me.srv.Store.ClusterDiscovery().Delete(&me.ClusterDiscovery); err != nil {
				mlog.Error("ClusterDiscoveryService failed to start clean", mlog.String("ClusterDiscovery", me.ClusterDiscovery.ToJson()), mlog.Err(err))
			}
		}
	}

	if err := me.srv.Store.ClusterDiscovery().Save(&me.ClusterDiscovery); err != nil {
		mlog.Error("ClusterDiscoveryService failed to save", mlog.String("ClusterDiscovery", me.ClusterDiscovery.ToJson()), mlog.Err(err))
		return
	}

	go func() {
		mlog.Debug("ClusterDiscoveryService ping writer started", mlog.String("ClusterDiscovery", me.ClusterDiscovery.ToJson()))
		ticker := time.NewTicker(DISCOVERY_SERVICE_WRITE_PING)
		defer func() {
			ticker.Stop()
			if _, err := me.srv.Store.ClusterDiscovery().Delete(&me.ClusterDiscovery); err != nil {
				mlog.Error("ClusterDiscoveryService failed to cleanup", mlog.String("ClusterDiscovery", me.ClusterDiscovery.ToJson()), mlog.Err(err))
			}
			mlog.Debug("ClusterDiscoveryService ping writer stopped", mlog.String("ClusterDiscovery", me.ClusterDiscovery.ToJson()))
		}()

		for {
			select {
			case <-ticker.C:
				if err := me.srv.Store.ClusterDiscovery().SetLastPingAt(&me.ClusterDiscovery); err != nil {
					mlog.Error("ClusterDiscoveryService failed to write ping", mlog.String("ClusterDiscovery", me.ClusterDiscovery.ToJson()), mlog.Err(err))
				}
			case <-me.stop:
				return
			}
		}
	}()
}

func (me *ClusterDiscoveryService) Stop() {
	me.stop <- true
}

func (s *Server) IsLeader() bool {
	if s.License() != nil && *s.Config().ClusterSettings.Enable && s.Cluster != nil {
		return s.Cluster.IsLeader()
	}
	return true
}

func (a *App) IsLeader() bool {
	return a.Srv().IsLeader()
}

func (a *App) GetClusterId() string {
	if a.Cluster() == nil {
		return ""
	}

	return a.Cluster().GetClusterId()
}
