// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package webutil

import (
	sqlx "github.com/jmoiron/sqlx"
	mock "github.com/stretchr/testify/mock"
)

// MockRecoverDB is an autogenerated mock type for the RecoverDB type
type MockRecoverDB struct {
	mock.Mock
}

// Execute provides a mock function with given fields: err
func (_m *MockRecoverDB) Execute(err error) (*sqlx.DB, error) {
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
