// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package webutil

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// MockRequestValidator is an autogenerated mock type for the RequestValidator type
type MockRequestValidator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: req, instance
func (_m *MockRequestValidator) Validate(req *http.Request, instance interface{}) (interface{}, error) {
	ret := _m.Called(req, instance)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(*http.Request, interface{}) interface{}); ok {
		r0 = rf(req, instance)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Request, interface{}) error); ok {
		r1 = rf(req, instance)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
