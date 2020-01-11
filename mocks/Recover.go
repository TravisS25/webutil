// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	webutil "github.com/TravisS25/webutil/webutil"
	mock "github.com/stretchr/testify/mock"
)

// Recover is an autogenerated mock type for the Recover type
type Recover struct {
	mock.Mock
}

// RecoverError provides a mock function with given fields: err
func (_m *Recover) RecoverError(err error) (*webutil.DB, error) {
	ret := _m.Called(err)

	var r0 *webutil.DB
	if rf, ok := ret.Get(0).(func(error) *webutil.DB); ok {
		r0 = rf(err)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*webutil.DB)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(error) error); ok {
		r1 = rf(err)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}