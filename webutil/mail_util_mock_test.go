// Code generated by MockGen. DO NOT EDIT.
// Source: mail_util.go

// Package webutil is a generated GoMock package.
package webutil

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockSendMessage is a mock of SendMessage interface
type MockSendMessage struct {
	ctrl     *gomock.Controller
	recorder *MockSendMessageMockRecorder
}

// MockSendMessageMockRecorder is the mock recorder for MockSendMessage
type MockSendMessageMockRecorder struct {
	mock *MockSendMessage
}

// NewMockSendMessage creates a new mock instance
func NewMockSendMessage(ctrl *gomock.Controller) *MockSendMessage {
	mock := &MockSendMessage{ctrl: ctrl}
	mock.recorder = &MockSendMessageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSendMessage) EXPECT() *MockSendMessageMockRecorder {
	return m.recorder
}

// Send mocks base method
func (m *MockSendMessage) Send(msg *Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockSendMessageMockRecorder) Send(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockSendMessage)(nil).Send), msg)
}
