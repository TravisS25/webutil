// Code generated by mockery v2.10.0. DO NOT EDIT.

package webutiltest

import mock "github.com/stretchr/testify/mock"

// MockTestLog is an autogenerated mock type for the TestLog type
type MockTestLog struct {
	mock.Mock
}

// Errorf provides a mock function with given fields: format, args
func (_m *MockTestLog) Errorf(format string, args ...any) {
	var _ca []any
	_ca = append(_ca, format)
	_ca = append(_ca, args...)
	_m.Called(_ca...)
}

// Helper provides a mock function with given fields:
func (_m *MockTestLog) Helper() {
	_m.Called()
}
