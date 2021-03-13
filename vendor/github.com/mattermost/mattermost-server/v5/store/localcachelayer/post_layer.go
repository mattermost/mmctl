// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package localcachelayer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

type LocalCachePostStore struct {
	store.PostStore
	rootStore *LocalCacheStore
}

func (s *LocalCachePostStore) handleClusterInvalidateLastPostTime(msg *model.ClusterMessage) {
	if msg.Data == ClearCacheMessageData {
		s.rootStore.lastPostTimeCache.Purge()
	} else {
		s.rootStore.lastPostTimeCache.Remove(msg.Data)
	}
}

func (s *LocalCachePostStore) handleClusterInvalidateLastPosts(msg *model.ClusterMessage) {
	if msg.Data == ClearCacheMessageData {
		s.rootStore.postLastPostsCache.Purge()
	} else {
		s.rootStore.postLastPostsCache.Remove(msg.Data)
	}
}

func (s LocalCachePostStore) ClearCaches() {
	s.rootStore.doClearCacheCluster(s.rootStore.lastPostTimeCache)
	s.rootStore.doClearCacheCluster(s.rootStore.postLastPostsCache)
	s.PostStore.ClearCaches()

	if s.rootStore.metrics != nil {
		s.rootStore.metrics.IncrementMemCacheInvalidationCounter("Last Post Time - Purge")
		s.rootStore.metrics.IncrementMemCacheInvalidationCounter("Last Posts Cache - Purge")
	}
}

func (s LocalCachePostStore) InvalidateLastPostTimeCache(channelId string) {
	s.rootStore.doInvalidateCacheCluster(s.rootStore.lastPostTimeCache, channelId)

	// Keys are "{channelid}{limit}" and caching only occurs on limits of 30 and 60
	s.rootStore.doInvalidateCacheCluster(s.rootStore.postLastPostsCache, channelId+"30")
	s.rootStore.doInvalidateCacheCluster(s.rootStore.postLastPostsCache, channelId+"60")

	s.PostStore.InvalidateLastPostTimeCache(channelId)

	if s.rootStore.metrics != nil {
		s.rootStore.metrics.IncrementMemCacheInvalidationCounter("Last Post Time - Remove by Channel Id")
		s.rootStore.metrics.IncrementMemCacheInvalidationCounter("Last Posts Cache - Remove by Channel Id")
	}
}

func (s LocalCachePostStore) GetEtag(channelId string, allowFromCache, collapsedThreads bool) string {
	if allowFromCache {
		var lastTime int64
		if err := s.rootStore.doStandardReadCache(s.rootStore.lastPostTimeCache, channelId, &lastTime); err == nil {
			return fmt.Sprintf("%v.%v", model.CurrentVersion, lastTime)
		}
	}

	result := s.PostStore.GetEtag(channelId, allowFromCache, collapsedThreads)

	splittedResult := strings.Split(result, ".")

	lastTime, _ := strconv.ParseInt((splittedResult[len(splittedResult)-1]), 10, 64)

	s.rootStore.doStandardAddToCache(s.rootStore.lastPostTimeCache, channelId, lastTime)

	return result
}

func (s LocalCachePostStore) GetPostsSince(options model.GetPostsSinceOptions, allowFromCache bool) (*model.PostList, error) {
	if allowFromCache {
		// If the last post in the channel's time is less than or equal to the time we are getting posts since,
		// we can safely return no posts.
		var lastTime int64
		if err := s.rootStore.doStandardReadCache(s.rootStore.lastPostTimeCache, options.ChannelId, &lastTime); err == nil && lastTime <= options.Time {
			list := model.NewPostList()
			return list, nil
		}
	}

	list, err := s.PostStore.GetPostsSince(options, allowFromCache)

	latestUpdate := options.Time
	if err == nil {
		for _, p := range list.ToSlice() {
			if latestUpdate < p.UpdateAt {
				latestUpdate = p.UpdateAt
			}
		}
		s.rootStore.doStandardAddToCache(s.rootStore.lastPostTimeCache, options.ChannelId, latestUpdate)
	}

	return list, err
}

func (s LocalCachePostStore) GetPosts(options model.GetPostsOptions, allowFromCache bool) (*model.PostList, error) {
	if !allowFromCache {
		return s.PostStore.GetPosts(options, allowFromCache)
	}

	offset := options.PerPage * options.Page
	// Caching only occurs on limits of 30 and 60, the common limits requested by MM clients
	if offset == 0 && (options.PerPage == 60 || options.PerPage == 30) {
		var cacheItem *model.PostList
		if err := s.rootStore.doStandardReadCache(s.rootStore.postLastPostsCache, fmt.Sprintf("%s%v", options.ChannelId, options.PerPage), &cacheItem); err == nil {
			return cacheItem, nil
		}
	}

	list, err := s.PostStore.GetPosts(options, false)
	if err != nil {
		return nil, err
	}

	// Caching only occurs on limits of 30 and 60, the common limits requested by MM clients
	if offset == 0 && (options.PerPage == 60 || options.PerPage == 30) {
		s.rootStore.doStandardAddToCache(s.rootStore.postLastPostsCache, fmt.Sprintf("%s%v", options.ChannelId, options.PerPage), list)
	}

	return list, err
}
