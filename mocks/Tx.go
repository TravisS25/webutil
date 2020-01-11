// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	sql "database/sql"

	mock "github.com/stretchr/testify/mock"
)

// Tx is an autogenerated mock type for the Tx type
type Tx struct {
	mock.Mock
}

// Begin provides a mock function with given fields:
func (_m *Tx) Begin() (*sql.Tx, error) {
	ret := _m.Called()

	var r0 *sql.Tx
	if rf, ok := ret.Get(0).(func() *sql.Tx); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sql.Tx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Commit provides a mock function with given fields: tx
func (_m *Tx) Commit(tx *sql.Tx) error {
	ret := _m.Called(tx)

	var r0 error
	if rf, ok := ret.Get(0).(func(*sql.Tx) error); ok {
		r0 = rf(tx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}