// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package webutil

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// MockRequestLog is an autogenerated mock type for the RequestLog type
type MockRequestLog struct {
	mock.Mock
}

// Execute provides a mock function with given fields: r, conf
func (_m *MockRequestLog) Execute(r *http.Request, conf LogConfig) {
	_m.Called(r, conf)
}
