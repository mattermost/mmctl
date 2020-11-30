// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import (
	model "github.com/mattermost/mattermost-server/v5/model"
	mock "github.com/stretchr/testify/mock"
)

// TermsOfServiceStore is an autogenerated mock type for the TermsOfServiceStore type
type TermsOfServiceStore struct {
	mock.Mock
}

// Get provides a mock function with given fields: id, allowFromCache
func (_m *TermsOfServiceStore) Get(id string, allowFromCache bool) (*model.TermsOfService, error) {
	ret := _m.Called(id, allowFromCache)

	var r0 *model.TermsOfService
	if rf, ok := ret.Get(0).(func(string, bool) *model.TermsOfService); ok {
		r0 = rf(id, allowFromCache)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.TermsOfService)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, bool) error); ok {
		r1 = rf(id, allowFromCache)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLatest provides a mock function with given fields: allowFromCache
func (_m *TermsOfServiceStore) GetLatest(allowFromCache bool) (*model.TermsOfService, error) {
	ret := _m.Called(allowFromCache)

	var r0 *model.TermsOfService
	if rf, ok := ret.Get(0).(func(bool) *model.TermsOfService); ok {
		r0 = rf(allowFromCache)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.TermsOfService)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(bool) error); ok {
		r1 = rf(allowFromCache)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: termsOfService
func (_m *TermsOfServiceStore) Save(termsOfService *model.TermsOfService) (*model.TermsOfService, error) {
	ret := _m.Called(termsOfService)

	var r0 *model.TermsOfService
	if rf, ok := ret.Get(0).(func(*model.TermsOfService) *model.TermsOfService); ok {
		r0 = rf(termsOfService)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.TermsOfService)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.TermsOfService) error); ok {
		r1 = rf(termsOfService)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
