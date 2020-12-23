// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package webutil

import (
	http "net/http"

	sessions "github.com/gorilla/sessions"
	mock "github.com/stretchr/testify/mock"
)

// MockSessionStore is an autogenerated mock type for the SessionStore type
type MockSessionStore struct {
	mock.Mock
}

// Get provides a mock function with given fields: r, name
func (_m *MockSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	ret := _m.Called(r, name)

	var r0 *sessions.Session
	if rf, ok := ret.Get(0).(func(*http.Request, string) *sessions.Session); ok {
		r0 = rf(r, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sessions.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Request, string) error); ok {
		r1 = rf(r, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// New provides a mock function with given fields: r, name
func (_m *MockSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	ret := _m.Called(r, name)

	var r0 *sessions.Session
	if rf, ok := ret.Get(0).(func(*http.Request, string) *sessions.Session); ok {
		r0 = rf(r, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sessions.Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Request, string) error); ok {
		r1 = rf(r, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ping provides a mock function with given fields:
func (_m *MockSessionStore) Ping() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Save provides a mock function with given fields: r, w, s
func (_m *MockSessionStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	ret := _m.Called(r, w, s)

	var r0 error
	if rf, ok := ret.Get(0).(func(*http.Request, http.ResponseWriter, *sessions.Session) error); ok {
		r0 = rf(r, w, s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
