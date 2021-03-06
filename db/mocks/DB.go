// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import models "github.com/eugeneverywhere/canopsis-test/db/models"
import types "github.com/eugeneverywhere/canopsis-test/types"

// DB is an autogenerated mock type for the DB type
type DB struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *DB) Close() {
	_m.Called()
}

// Connect provides a mock function with given fields:
func (_m *DB) Connect() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateAlarm provides a mock function with given fields: event
func (_m *DB) CreateAlarm(event *types.Event) error {
	ret := _m.Called(event)

	var r0 error
	if rf, ok := ret.Get(0).(func(*types.Event) error); ok {
		r0 = rf(event)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAlarm provides a mock function with given fields: event
func (_m *DB) GetAlarm(event *types.Event) (*models.Alarm, error) {
	ret := _m.Called(event)

	var r0 *models.Alarm
	if rf, ok := ret.Get(0).(func(*types.Event) *models.Alarm); ok {
		r0 = rf(event)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Alarm)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*types.Event) error); ok {
		r1 = rf(event)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Lock provides a mock function with given fields: key
func (_m *DB) Lock(key string) error {
	ret := _m.Called(key)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unlock provides a mock function with given fields: key
func (_m *DB) Unlock(key string) error {
	ret := _m.Called(key)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAlarm provides a mock function with given fields: event, status
func (_m *DB) UpdateAlarm(event *types.Event, status string) error {
	ret := _m.Called(event, status)

	var r0 error
	if rf, ok := ret.Get(0).(func(*types.Event, string) error); ok {
		r0 = rf(event, status)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
