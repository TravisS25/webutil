// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	sql "database/sql"

	mock "github.com/stretchr/testify/mock"
)

// DBInterface is an autogenerated mock type for the DBInterface type
type DBInterface struct {
	mock.Mock
}

// Begin provides a mock function with given fields:
func (_m *DBInterface) Begin() (*sql.Tx, error) {
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

// Exec provides a mock function with given fields: _a0, _a1
func (_m *DBInterface) Exec(_a0 string, _a1 ...interface{}) (sql.Result, error) {
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _a1...)
	ret := _m.Called(_ca...)

	var r0 sql.Result
	if rf, ok := ret.Get(0).(func(string, ...interface{}) sql.Result); ok {
		r0 = rf(_a0, _a1...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sql.Result)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, ...interface{}) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: dest, query, args
func (_m *DBInterface) Get(dest interface{}, query string, args ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, dest, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}, string, ...interface{}) error); ok {
		r0 = rf(dest, query, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Query provides a mock function with given fields: query, args
func (_m *DBInterface) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var _ca []interface{}
	_ca = append(_ca, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 *sql.Rows
	if rf, ok := ret.Get(0).(func(string, ...interface{}) *sql.Rows); ok {
		r0 = rf(query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sql.Rows)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, ...interface{}) error); ok {
		r1 = rf(query, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueryRow provides a mock function with given fields: query, args
func (_m *DBInterface) QueryRow(query string, args ...interface{}) *sql.Row {
	var _ca []interface{}
	_ca = append(_ca, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 *sql.Row
	if rf, ok := ret.Get(0).(func(string, ...interface{}) *sql.Row); ok {
		r0 = rf(query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sql.Row)
		}
	}

	return r0
}

// Select provides a mock function with given fields: dest, query, args
func (_m *DBInterface) Select(dest interface{}, query string, args ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, dest, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}, string, ...interface{}) error); ok {
		r0 = rf(dest, query, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
