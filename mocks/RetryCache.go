// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	webutil "github.com/TravisS25/webutil/webutil"
	mock "github.com/stretchr/testify/mock"
)

// RetryCache is an autogenerated mock type for the RetryCache type
type RetryCache struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0
func (_m *RetryCache) Execute(_a0 webutil.CacheStore) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(webutil.CacheStore) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
