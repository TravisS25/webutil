// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	http "net/http"

	webutil "github.com/TravisS25/webutil/webutil"
	mock "github.com/stretchr/testify/mock"
)

// RequestLog is an autogenerated mock type for the RequestLog type
type RequestLog struct {
	mock.Mock
}

// Execute provides a mock function with given fields: r, conf
func (_m *RequestLog) Execute(r *http.Request, conf webutil.LogConfig) {
	_m.Called(r, conf)
}
