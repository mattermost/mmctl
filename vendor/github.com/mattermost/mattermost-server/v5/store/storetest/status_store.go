// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package storetest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

func TestStatusStore(t *testing.T, ss store.Store) {
	t.Run("", func(t *testing.T) { testStatusStore(t, ss) })
	t.Run("ActiveUserCount", func(t *testing.T) { testActiveUserCount(t, ss) })
	t.Run("UpdateExpiredDNDStatuses", func(t *testing.T) { testUpdateExpiredDNDStatuses(t, ss) })
}

func testStatusStore(t *testing.T, ss store.Store) {
	status := &model.Status{UserId: model.NewId(), Status: model.STATUS_ONLINE, Manual: false, LastActivityAt: 0, ActiveChannel: ""}
	require.NoError(t, ss.Status().SaveOrUpdate(status))

	status.LastActivityAt = 10

	_, err := ss.Status().Get(status.UserId)
	require.NoError(t, err)

	status2 := &model.Status{UserId: model.NewId(), Status: model.STATUS_AWAY, Manual: false, LastActivityAt: 0, ActiveChannel: ""}
	require.NoError(t, ss.Status().SaveOrUpdate(status2))

	status3 := &model.Status{UserId: model.NewId(), Status: model.STATUS_OFFLINE, Manual: false, LastActivityAt: 0, ActiveChannel: ""}
	require.NoError(t, ss.Status().SaveOrUpdate(status3))

	statuses, err := ss.Status().GetByIds([]string{status.UserId, "junk"})
	require.NoError(t, err)
	require.Len(t, statuses, 1, "should only have 1 status")

	err = ss.Status().ResetAll()
	require.NoError(t, err)

	statusParameter, err := ss.Status().Get(status.UserId)
	require.NoError(t, err)
	require.Equal(t, statusParameter.Status, model.STATUS_OFFLINE, "should be offline")

	err = ss.Status().UpdateLastActivityAt(status.UserId, 10)
	require.NoError(t, err)
}

func testActiveUserCount(t *testing.T, ss store.Store) {
	status := &model.Status{UserId: model.NewId(), Status: model.STATUS_ONLINE, Manual: false, LastActivityAt: model.GetMillis(), ActiveChannel: ""}
	require.NoError(t, ss.Status().SaveOrUpdate(status))

	count, err := ss.Status().GetTotalActiveUsersCount()
	require.NoError(t, err)
	require.True(t, count > 0, "expected count > 0, got %d", count)
}

type ByUserId []*model.Status

func (s ByUserId) Len() int           { return len(s) }
func (s ByUserId) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByUserId) Less(i, j int) bool { return s[i].UserId < s[j].UserId }

func testUpdateExpiredDNDStatuses(t *testing.T, ss store.Store) {
	userID := NewTestId()

	status := &model.Status{UserId: userID, Status: model.STATUS_DND, Manual: true,
		DNDEndTime: time.Now().Add(5 * time.Second).Unix(), PrevStatus: model.STATUS_ONLINE}
	require.NoError(t, ss.Status().SaveOrUpdate(status))

	time.Sleep(2 * time.Second)

	// after 2 seconds no statuses should be expired
	statuses, err := ss.Status().UpdateExpiredDNDStatuses()
	require.NoError(t, err)
	require.Len(t, statuses, 0)

	time.Sleep(3 * time.Second)

	// after 3 more seconds test status should be updated
	statuses, err = ss.Status().UpdateExpiredDNDStatuses()
	require.NoError(t, err)
	require.Len(t, statuses, 1)

	updatedStatus := *statuses[0]
	require.Equal(t, updatedStatus.UserId, userID)
	require.Equal(t, updatedStatus.Status, model.STATUS_ONLINE)
	require.Equal(t, updatedStatus.DNDEndTime, int64(0))
	require.Equal(t, updatedStatus.PrevStatus, model.STATUS_DND)
	require.Equal(t, updatedStatus.Manual, false)
}
