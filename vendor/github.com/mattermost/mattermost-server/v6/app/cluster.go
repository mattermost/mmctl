// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

type clusterWrapper struct {
	srv *Server
}

func (s *clusterWrapper) PublishPluginClusterEvent(productID string, ev model.PluginClusterEvent,
	opts model.PluginClusterEventSendOptions) error {
	if s.srv.Cluster == nil {
		return nil
	}

	msg := &model.ClusterMessage{
		Event:            model.ClusterEventPluginEvent,
		SendType:         opts.SendType,
		WaitForAllToSend: false,
		Props: map[string]string{
			"ProductID": productID,
			"EventID":   ev.Id,
		},
		Data: ev.Data,
	}

	// If TargetId is empty we broadcast to all other cluster nodes.
	if opts.TargetId == "" {
		s.srv.Cluster.SendClusterMessage(msg)
	} else {
		if err := s.srv.Cluster.SendClusterMessageToNode(opts.TargetId, msg); err != nil {
			return fmt.Errorf("failed to send message to cluster node %q: %w", opts.TargetId, err)
		}
	}

	return nil
}

func (s *clusterWrapper) SetPluginKeyWithOptions(productID string, key string, value []byte, options model.PluginKVSetOptions) (bool, *model.AppError) {
	return s.srv.setPluginKeyWithOptions(productID, key, value, options)
}

func (s *clusterWrapper) LogError(productID, msg string, keyValuePairs ...interface{}) {
	s.srv.Log.Error(msg, mlog.String("product_id", productID), mlog.Map("key-value pairs", keyValuePairs))
}

func (s *clusterWrapper) KVGet(productID, key string) ([]byte, *model.AppError) {
	return s.srv.getPluginKey(productID, key)
}

func (s *clusterWrapper) KVDelete(productID, key string) *model.AppError {
	return s.srv.deletePluginKey(productID, key)
}

func (s *clusterWrapper) KVList(productID string, page, perPage int) ([]string, *model.AppError) {
	return s.srv.listPluginKeys(productID, page, perPage)
}

// Registers a given function to be called when the cluster leader may have changed. Returns a unique ID for the
// listener which can later be used to remove it. If clustering is not enabled in this build, the callback will never
// be called.
func (s *Server) AddClusterLeaderChangedListener(listener func()) string {
	id := model.NewId()
	s.clusterLeaderListeners.Store(id, listener)
	return id
}

// Removes a listener function by the unique ID returned when AddConfigListener was called
func (s *Server) RemoveClusterLeaderChangedListener(id string) {
	s.clusterLeaderListeners.Delete(id)
}

func (s *Server) InvokeClusterLeaderChangedListeners() {
	s.Log.Info("Cluster leader changed. Invoking ClusterLeaderChanged listeners.")
	// This needs to be run in a separate goroutine otherwise a recursive lock happens
	// because the listener function eventually ends up calling .IsLeader().
	// Fixing this would require the changed event to pass the leader directly, but that
	// requires a lot of work.
	s.Go(func() {
		s.clusterLeaderListeners.Range(func(_, listener interface{}) bool {
			listener.(func())()
			return true
		})
	})
}
