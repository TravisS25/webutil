// Code generated by mockery v1.0.0. DO NOT EDIT.

package webutil

import (
	"github.com/jmoiron/sqlx"
	mock "github.com/stretchr/testify/mock"
)

// MockRecover is an autogenerated mock type for the Recover type
type MockRecover struct {
	mock.Mock
}

// RecoverError provides a mock function with given fields: err
func (_m *MockRecover) RecoverError(err error) (*sqlx.DB, error) {
	ret := _m.Called(err)

	var r0 *sqlx.DB
	if rf, ok := ret.Get(0).(func(error) *sqlx.DB); ok {
		r0 = rf(err)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlx.DB)
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
