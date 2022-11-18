// Code generated by mockery v2.10.4. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import (
	model "github.com/mattermost/mattermost-server/v6/model"
	mock "github.com/stretchr/testify/mock"
)

// JobStore is an autogenerated mock type for the JobStore type
type JobStore struct {
	mock.Mock
}

// Cleanup provides a mock function with given fields: expiryTime, batchSize
func (_m *JobStore) Cleanup(expiryTime int64, batchSize int) error {
	ret := _m.Called(expiryTime, batchSize)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, int) error); ok {
		r0 = rf(expiryTime, batchSize)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: id
func (_m *JobStore) Delete(id string) (string, error) {
	ret := _m.Called(id)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: id
func (_m *JobStore) Get(id string) (*model.Job, error) {
	ret := _m.Called(id)

	var r0 *model.Job
	if rf, ok := ret.Get(0).(func(string) *model.Job); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllByStatus provides a mock function with given fields: status
func (_m *JobStore) GetAllByStatus(status string) ([]*model.Job, error) {
	ret := _m.Called(status)

	var r0 []*model.Job
	if rf, ok := ret.Get(0).(func(string) []*model.Job); ok {
		r0 = rf(status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(status)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllByType provides a mock function with given fields: jobType
func (_m *JobStore) GetAllByType(jobType string) ([]*model.Job, error) {
	ret := _m.Called(jobType)

	var r0 []*model.Job
	if rf, ok := ret.Get(0).(func(string) []*model.Job); ok {
		r0 = rf(jobType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(jobType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllByTypeAndStatus provides a mock function with given fields: jobType, status
func (_m *JobStore) GetAllByTypeAndStatus(jobType string, status string) ([]*model.Job, error) {
	ret := _m.Called(jobType, status)

	var r0 []*model.Job
	if rf, ok := ret.Get(0).(func(string, string) []*model.Job); ok {
		r0 = rf(jobType, status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(jobType, status)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllByTypePage provides a mock function with given fields: jobType, offset, limit
func (_m *JobStore) GetAllByTypePage(jobType string, offset int, limit int) ([]*model.Job, error) {
	ret := _m.Called(jobType, offset, limit)

	var r0 []*model.Job
	if rf, ok := ret.Get(0).(func(string, int, int) []*model.Job); ok {
		r0 = rf(jobType, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, int, int) error); ok {
		r1 = rf(jobType, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllByTypesPage provides a mock function with given fields: jobTypes, offset, limit
func (_m *JobStore) GetAllByTypesPage(jobTypes []string, offset int, limit int) ([]*model.Job, error) {
	ret := _m.Called(jobTypes, offset, limit)

	var r0 []*model.Job
	if rf, ok := ret.Get(0).(func([]string, int, int) []*model.Job); ok {
		r0 = rf(jobTypes, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string, int, int) error); ok {
		r1 = rf(jobTypes, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllPage provides a mock function with given fields: offset, limit
func (_m *JobStore) GetAllPage(offset int, limit int) ([]*model.Job, error) {
	ret := _m.Called(offset, limit)

	var r0 []*model.Job
	if rf, ok := ret.Get(0).(func(int, int) []*model.Job); ok {
		r0 = rf(offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, int) error); ok {
		r1 = rf(offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCountByStatusAndType provides a mock function with given fields: status, jobType
func (_m *JobStore) GetCountByStatusAndType(status string, jobType string) (int64, error) {
	ret := _m.Called(status, jobType)

	var r0 int64
	if rf, ok := ret.Get(0).(func(string, string) int64); ok {
		r0 = rf(status, jobType)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(status, jobType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNewestJobByStatusAndType provides a mock function with given fields: status, jobType
func (_m *JobStore) GetNewestJobByStatusAndType(status string, jobType string) (*model.Job, error) {
	ret := _m.Called(status, jobType)

	var r0 *model.Job
	if rf, ok := ret.Get(0).(func(string, string) *model.Job); ok {
		r0 = rf(status, jobType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(status, jobType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNewestJobByStatusesAndType provides a mock function with given fields: statuses, jobType
func (_m *JobStore) GetNewestJobByStatusesAndType(statuses []string, jobType string) (*model.Job, error) {
	ret := _m.Called(statuses, jobType)

	var r0 *model.Job
	if rf, ok := ret.Get(0).(func([]string, string) *model.Job); ok {
		r0 = rf(statuses, jobType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string, string) error); ok {
		r1 = rf(statuses, jobType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: job
func (_m *JobStore) Save(job *model.Job) (*model.Job, error) {
	ret := _m.Called(job)

	var r0 *model.Job
	if rf, ok := ret.Get(0).(func(*model.Job) *model.Job); ok {
		r0 = rf(job)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Job) error); ok {
		r1 = rf(job)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateOptimistically provides a mock function with given fields: job, currentStatus
func (_m *JobStore) UpdateOptimistically(job *model.Job, currentStatus string) (bool, error) {
	ret := _m.Called(job, currentStatus)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*model.Job, string) bool); ok {
		r0 = rf(job, currentStatus)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Job, string) error); ok {
		r1 = rf(job, currentStatus)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateStatus provides a mock function with given fields: id, status
func (_m *JobStore) UpdateStatus(id string, status string) (*model.Job, error) {
	ret := _m.Called(id, status)

	var r0 *model.Job
	if rf, ok := ret.Get(0).(func(string, string) *model.Job); ok {
		r0 = rf(id, status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Job)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(id, status)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateStatusOptimistically provides a mock function with given fields: id, currentStatus, newStatus
func (_m *JobStore) UpdateStatusOptimistically(id string, currentStatus string, newStatus string) (bool, error) {
	ret := _m.Called(id, currentStatus, newStatus)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string, string) bool); ok {
		r0 = rf(id, currentStatus, newStatus)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(id, currentStatus, newStatus)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
