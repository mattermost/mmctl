// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package storetest

import (
	"errors"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

func TestJobStore(t *testing.T, ss store.Store) {
	t.Run("JobSaveGet", func(t *testing.T) { testJobSaveGet(t, ss) })
	t.Run("JobGetAllByType", func(t *testing.T) { testJobGetAllByType(t, ss) })
	t.Run("JobGetAllByTypePage", func(t *testing.T) { testJobGetAllByTypePage(t, ss) })
	t.Run("JobGetAllPage", func(t *testing.T) { testJobGetAllPage(t, ss) })
	t.Run("JobGetAllByStatus", func(t *testing.T) { testJobGetAllByStatus(t, ss) })
	t.Run("GetNewestJobByStatusAndType", func(t *testing.T) { testJobStoreGetNewestJobByStatusAndType(t, ss) })
	t.Run("GetNewestJobByStatusesAndType", func(t *testing.T) { testJobStoreGetNewestJobByStatusesAndType(t, ss) })
	t.Run("GetCountByStatusAndType", func(t *testing.T) { testJobStoreGetCountByStatusAndType(t, ss) })
	t.Run("JobUpdateOptimistically", func(t *testing.T) { testJobUpdateOptimistically(t, ss) })
	t.Run("JobUpdateStatusUpdateStatusOptimistically", func(t *testing.T) { testJobUpdateStatusUpdateStatusOptimistically(t, ss) })
	t.Run("JobDelete", func(t *testing.T) { testJobDelete(t, ss) })
}

func testJobSaveGet(t *testing.T, ss store.Store) {
	job := &model.Job{
		Id:     model.NewId(),
		Type:   model.NewId(),
		Status: model.NewId(),
		Data: map[string]string{
			"Processed":     "0",
			"Total":         "12345",
			"LastProcessed": "abcd",
		},
	}

	_, err := ss.Job().Save(job)
	require.NoError(t, err)

	defer ss.Job().Delete(job.Id)

	received, err := ss.Job().Get(job.Id)
	require.NoError(t, err)
	require.Equal(t, job.Id, received.Id, "received incorrect job after save")
	require.Equal(t, "12345", received.Data["Total"])
}

func testJobGetAllByType(t *testing.T, ss store.Store) {
	jobType := model.NewId()

	jobs := []*model.Job{
		{
			Id:   model.NewId(),
			Type: jobType,
		},
		{
			Id:   model.NewId(),
			Type: jobType,
		},
		{
			Id:   model.NewId(),
			Type: model.NewId(),
		},
	}

	for _, job := range jobs {
		_, err := ss.Job().Save(job)
		require.NoError(t, err)
		defer ss.Job().Delete(job.Id)
	}

	received, err := ss.Job().GetAllByType(jobType)
	require.NoError(t, err)
	require.Len(t, received, 2)
	require.ElementsMatch(t, []string{jobs[0].Id, jobs[1].Id}, []string{received[0].Id, received[1].Id})
}

func testJobGetAllByTypePage(t *testing.T, ss store.Store) {
	jobType := model.NewId()

	jobs := []*model.Job{
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: 1000,
		},
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: 999,
		},
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: 1001,
		},
		{
			Id:       model.NewId(),
			Type:     model.NewId(),
			CreateAt: 1002,
		},
	}

	for _, job := range jobs {
		_, err := ss.Job().Save(job)
		require.NoError(t, err)
		defer ss.Job().Delete(job.Id)
	}

	received, err := ss.Job().GetAllByTypePage(jobType, 0, 2)
	require.NoError(t, err)
	require.Len(t, received, 2)
	require.Equal(t, received[0].Id, jobs[2].Id, "should've received newest job first")
	require.Equal(t, received[1].Id, jobs[0].Id, "should've received second newest job second")

	received, err = ss.Job().GetAllByTypePage(jobType, 2, 2)
	require.NoError(t, err)
	require.Len(t, received, 1)
	require.Equal(t, received[0].Id, jobs[1].Id, "should've received oldest job last")
}

func testJobGetAllPage(t *testing.T, ss store.Store) {
	jobType := model.NewId()
	createAtTime := model.GetMillis()

	jobs := []*model.Job{
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: createAtTime + 1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: createAtTime,
		},
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: createAtTime + 2,
		},
	}

	for _, job := range jobs {
		_, err := ss.Job().Save(job)
		require.NoError(t, err)
		defer ss.Job().Delete(job.Id)
	}

	received, err := ss.Job().GetAllPage(0, 2)
	require.NoError(t, err)
	require.Len(t, received, 2)
	require.Equal(t, received[0].Id, jobs[2].Id, "should've received newest job first")
	require.Equal(t, received[1].Id, jobs[0].Id, "should've received second newest job second")

	received, err = ss.Job().GetAllPage(2, 2)
	require.NoError(t, err)
	require.NotEmpty(t, received)
	require.Equal(t, received[0].Id, jobs[1].Id, "should've received oldest job last")
}

func testJobGetAllByStatus(t *testing.T, ss store.Store) {
	jobType := model.NewId()
	status := model.NewId()

	jobs := []*model.Job{
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: 1000,
			Status:   status,
			Data: map[string]string{
				"test": "data",
			},
		},
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: 999,
			Status:   status,
		},
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: 1001,
			Status:   status,
		},
		{
			Id:       model.NewId(),
			Type:     jobType,
			CreateAt: 1002,
			Status:   model.NewId(),
		},
	}

	for _, job := range jobs {
		_, err := ss.Job().Save(job)
		require.NoError(t, err)
		defer ss.Job().Delete(job.Id)
	}

	received, err := ss.Job().GetAllByStatus(status)
	require.NoError(t, err)
	require.Len(t, received, 3)
	require.Equal(t, received[0].Id, jobs[1].Id)
	require.Equal(t, received[1].Id, jobs[0].Id)
	require.Equal(t, received[2].Id, jobs[2].Id)
	require.Equal(t, "data", received[1].Data["test"], "should've received job data field back as saved")
}

func testJobStoreGetNewestJobByStatusAndType(t *testing.T, ss store.Store) {
	jobType1 := model.NewId()
	jobType2 := model.NewId()
	status1 := model.NewId()
	status2 := model.NewId()

	jobs := []*model.Job{
		{
			Id:       model.NewId(),
			Type:     jobType1,
			CreateAt: 1001,
			Status:   status1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType1,
			CreateAt: 1000,
			Status:   status1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType2,
			CreateAt: 1003,
			Status:   status1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType1,
			CreateAt: 1004,
			Status:   status2,
		},
	}

	for _, job := range jobs {
		_, err := ss.Job().Save(job)
		require.NoError(t, err)
		defer ss.Job().Delete(job.Id)
	}

	received, err := ss.Job().GetNewestJobByStatusAndType(status1, jobType1)
	assert.NoError(t, err)
	assert.EqualValues(t, jobs[0].Id, received.Id)

	received, err = ss.Job().GetNewestJobByStatusAndType(model.NewId(), model.NewId())
	assert.Error(t, err)
	var nfErr *store.ErrNotFound
	assert.True(t, errors.As(err, &nfErr))
	assert.Nil(t, received)
}

func testJobStoreGetNewestJobByStatusesAndType(t *testing.T, ss store.Store) {
	jobType1 := model.NewId()
	jobType2 := model.NewId()
	status1 := model.NewId()
	status2 := model.NewId()

	jobs := []*model.Job{
		{
			Id:       model.NewId(),
			Type:     jobType1,
			CreateAt: 1001,
			Status:   status1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType1,
			CreateAt: 1000,
			Status:   status1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType2,
			CreateAt: 1003,
			Status:   status1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType1,
			CreateAt: 1004,
			Status:   status2,
		},
	}

	for _, job := range jobs {
		_, err := ss.Job().Save(job)
		require.NoError(t, err)
		defer ss.Job().Delete(job.Id)
	}

	received, err := ss.Job().GetNewestJobByStatusesAndType([]string{status1, status2}, jobType1)
	assert.NoError(t, err)
	assert.EqualValues(t, jobs[3].Id, received.Id)

	received, err = ss.Job().GetNewestJobByStatusesAndType([]string{model.NewId(), model.NewId()}, model.NewId())
	assert.Error(t, err)
	var nfErr *store.ErrNotFound
	assert.True(t, errors.As(err, &nfErr))
	assert.Nil(t, received)

	received, err = ss.Job().GetNewestJobByStatusesAndType([]string{status2}, jobType2)
	assert.Error(t, err)
	assert.True(t, errors.As(err, &nfErr))
	assert.Nil(t, received)

	received, err = ss.Job().GetNewestJobByStatusesAndType([]string{status1}, jobType2)
	assert.NoError(t, err)
	assert.EqualValues(t, jobs[2].Id, received.Id)

	received, err = ss.Job().GetNewestJobByStatusesAndType([]string{}, jobType1)
	assert.Error(t, err)
	assert.True(t, errors.As(err, &nfErr))
	assert.Nil(t, received)
}

func testJobStoreGetCountByStatusAndType(t *testing.T, ss store.Store) {
	jobType1 := model.NewId()
	jobType2 := model.NewId()
	status1 := model.NewId()
	status2 := model.NewId()

	jobs := []*model.Job{
		{
			Id:       model.NewId(),
			Type:     jobType1,
			CreateAt: 1000,
			Status:   status1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType1,
			CreateAt: 999,
			Status:   status1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType2,
			CreateAt: 1001,
			Status:   status1,
		},
		{
			Id:       model.NewId(),
			Type:     jobType1,
			CreateAt: 1002,
			Status:   status2,
		},
	}

	for _, job := range jobs {
		_, err := ss.Job().Save(job)
		require.NoError(t, err)
		defer ss.Job().Delete(job.Id)
	}

	count, err := ss.Job().GetCountByStatusAndType(status1, jobType1)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, count)

	count, err = ss.Job().GetCountByStatusAndType(status2, jobType2)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, count)

	count, err = ss.Job().GetCountByStatusAndType(status1, jobType2)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, count)

	count, err = ss.Job().GetCountByStatusAndType(status2, jobType1)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, count)
}

func testJobUpdateOptimistically(t *testing.T, ss store.Store) {
	job := &model.Job{
		Id:       model.NewId(),
		Type:     model.JOB_TYPE_DATA_RETENTION,
		CreateAt: model.GetMillis(),
		Status:   model.JOB_STATUS_PENDING,
	}

	_, err := ss.Job().Save(job)
	require.NoError(t, err)
	defer ss.Job().Delete(job.Id)

	job.LastActivityAt = model.GetMillis()
	job.Status = model.JOB_STATUS_IN_PROGRESS
	job.Progress = 50
	job.Data = map[string]string{
		"Foo": "Bar",
	}

	updated, err := ss.Job().UpdateOptimistically(job, model.JOB_STATUS_SUCCESS)
	require.False(t, err != nil && updated)

	time.Sleep(2 * time.Millisecond)

	updated, err = ss.Job().UpdateOptimistically(job, model.JOB_STATUS_PENDING)
	require.NoError(t, err)
	require.True(t, updated)

	updatedJob, err := ss.Job().Get(job.Id)
	require.NoError(t, err)

	require.Equal(t, updatedJob.Type, job.Type)
	require.Equal(t, updatedJob.CreateAt, job.CreateAt)
	require.Equal(t, updatedJob.Status, job.Status)
	require.Greater(t, updatedJob.LastActivityAt, job.LastActivityAt)
	require.Equal(t, updatedJob.Progress, job.Progress)
	require.Equal(t, updatedJob.Data["Foo"], job.Data["Foo"])
}

func testJobUpdateStatusUpdateStatusOptimistically(t *testing.T, ss store.Store) {
	job := &model.Job{
		Id:       model.NewId(),
		Type:     model.JOB_TYPE_DATA_RETENTION,
		CreateAt: model.GetMillis(),
		Status:   model.JOB_STATUS_SUCCESS,
	}

	var lastUpdateAt int64
	received, err := ss.Job().Save(job)
	require.NoError(t, err)
	lastUpdateAt = received.LastActivityAt

	defer ss.Job().Delete(job.Id)

	time.Sleep(2 * time.Millisecond)

	received, err = ss.Job().UpdateStatus(job.Id, model.JOB_STATUS_PENDING)
	require.NoError(t, err)

	require.Equal(t, model.JOB_STATUS_PENDING, received.Status)
	require.Greater(t, received.LastActivityAt, lastUpdateAt)
	lastUpdateAt = received.LastActivityAt

	time.Sleep(2 * time.Millisecond)

	updated, err := ss.Job().UpdateStatusOptimistically(job.Id, model.JOB_STATUS_IN_PROGRESS, model.JOB_STATUS_SUCCESS)
	require.NoError(t, err)
	require.False(t, updated)

	received, err = ss.Job().Get(job.Id)
	require.NoError(t, err)

	require.Equal(t, model.JOB_STATUS_PENDING, received.Status)
	require.Equal(t, received.LastActivityAt, lastUpdateAt)

	time.Sleep(2 * time.Millisecond)

	updated, err = ss.Job().UpdateStatusOptimistically(job.Id, model.JOB_STATUS_PENDING, model.JOB_STATUS_IN_PROGRESS)
	require.NoError(t, err)
	require.True(t, updated, "should have succeeded")

	var startAtSet int64
	received, err = ss.Job().Get(job.Id)
	require.NoError(t, err)
	require.Equal(t, model.JOB_STATUS_IN_PROGRESS, received.Status)
	require.NotEqual(t, 0, received.StartAt)
	require.Greater(t, received.LastActivityAt, lastUpdateAt)
	lastUpdateAt = received.LastActivityAt
	startAtSet = received.StartAt

	time.Sleep(2 * time.Millisecond)

	updated, err = ss.Job().UpdateStatusOptimistically(job.Id, model.JOB_STATUS_IN_PROGRESS, model.JOB_STATUS_SUCCESS)
	require.NoError(t, err)
	require.True(t, updated, "should have succeeded")

	received, err = ss.Job().Get(job.Id)
	require.NoError(t, err)
	require.Equal(t, model.JOB_STATUS_SUCCESS, received.Status)
	require.Equal(t, startAtSet, received.StartAt)
	require.Greater(t, received.LastActivityAt, lastUpdateAt)
}

func testJobDelete(t *testing.T, ss store.Store) {
	job, err := ss.Job().Save(&model.Job{Id: model.NewId()})
	require.NoError(t, err)

	_, err = ss.Job().Delete(job.Id)
	assert.NoError(t, err)
}
