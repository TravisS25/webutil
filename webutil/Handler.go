// Code generated by mockery v2.10.0. DO NOT EDIT.

package webutil

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// Handler is an autogenerated mock type for the Handler type
type Handler struct {
	mock.Mock
}

// ServeHTTP provides a mock function with given fields: _a0, _a1
func (_m *Handler) ServeHTTP(_a0 http.ResponseWriter, _a1 *http.Request) {
	_m.Called(_a0, _a1)
}
