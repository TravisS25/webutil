// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	sql "database/sql"

	mock "github.com/stretchr/testify/mock"

	sqlx "github.com/jmoiron/sqlx"
)

// QuerierExec is an autogenerated mock type for the QuerierExec type
type QuerierExec struct {
	mock.Mock
}

// Exec provides a mock function with given fields: _a0, _a1
func (_m *QuerierExec) Exec(_a0 string, _a1 ...interface{}) (sql.Result, error) {
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

// QueryRowx provides a mock function with given fields: query, args
func (_m *QuerierExec) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	var _ca []interface{}
	_ca = append(_ca, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 *sqlx.Row
	if rf, ok := ret.Get(0).(func(string, ...interface{}) *sqlx.Row); ok {
		r0 = rf(query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlx.Row)
		}
	}

	return r0
}

// Queryx provides a mock function with given fields: query, args
func (_m *QuerierExec) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	var _ca []interface{}
	_ca = append(_ca, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 *sqlx.Rows
	if rf, ok := ret.Get(0).(func(string, ...interface{}) *sqlx.Rows); ok {
		r0 = rf(query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sqlx.Rows)
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
