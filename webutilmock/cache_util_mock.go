// Code generated by MockGen. DO NOT EDIT.
// Source: cache_util.go

// Package webutilmock is a generated GoMock package.
package webutilmock

import (
	redis "github.com/go-redis/redis"
	gomock "github.com/golang/mock/gomock"
	sessions "github.com/gorilla/sessions"
	http "net/http"
	reflect "reflect"
	time "time"
)

// MockRedisSessionStore is a mock of RedisSessionStore interface
type MockRedisSessionStore struct {
	ctrl     *gomock.Controller
	recorder *MockRedisSessionStoreMockRecorder
}

// MockRedisSessionStoreMockRecorder is the mock recorder for MockRedisSessionStore
type MockRedisSessionStoreMockRecorder struct {
	mock *MockRedisSessionStore
}

// NewMockRedisSessionStore creates a new mock instance
func NewMockRedisSessionStore(ctrl *gomock.Controller) *MockRedisSessionStore {
	mock := &MockRedisSessionStore{ctrl: ctrl}
	mock.recorder = &MockRedisSessionStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRedisSessionStore) EXPECT() *MockRedisSessionStoreMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockRedisSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", r, name)
	ret0, _ := ret[0].(*sessions.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockRedisSessionStoreMockRecorder) Get(r, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockRedisSessionStore)(nil).Get), r, name)
}

// New mocks base method
func (m *MockRedisSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New", r, name)
	ret0, _ := ret[0].(*sessions.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// New indicates an expected call of New
func (mr *MockRedisSessionStoreMockRecorder) New(r, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockRedisSessionStore)(nil).New), r, name)
}

// Save mocks base method
func (m *MockRedisSessionStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", r, w, s)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save
func (mr *MockRedisSessionStoreMockRecorder) Save(r, w, s interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockRedisSessionStore)(nil).Save), r, w, s)
}

// Pool mocks base method
func (m *MockRedisSessionStore) Pool() *redis.Conn {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pool")
	ret0, _ := ret[0].(*redis.Conn)
	return ret0
}

// Pool indicates an expected call of Pool
func (mr *MockRedisSessionStoreMockRecorder) Pool() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pool", reflect.TypeOf((*MockRedisSessionStore)(nil).Pool))
}

// MockCacheStore is a mock of CacheStore interface
type MockCacheStore struct {
	ctrl     *gomock.Controller
	recorder *MockCacheStoreMockRecorder
}

// MockCacheStoreMockRecorder is the mock recorder for MockCacheStore
type MockCacheStoreMockRecorder struct {
	mock *MockCacheStore
}

// NewMockCacheStore creates a new mock instance
func NewMockCacheStore(ctrl *gomock.Controller) *MockCacheStore {
	mock := &MockCacheStore{ctrl: ctrl}
	mock.recorder = &MockCacheStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCacheStore) EXPECT() *MockCacheStoreMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockCacheStore) Get(key string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", key)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockCacheStoreMockRecorder) Get(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCacheStore)(nil).Get), key)
}

// Set mocks base method
func (m *MockCacheStore) Set(key string, value interface{}, expiration time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Set", key, value, expiration)
}

// Set indicates an expected call of Set
func (mr *MockCacheStoreMockRecorder) Set(key, value, expiration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCacheStore)(nil).Set), key, value, expiration)
}

// Del mocks base method
func (m *MockCacheStore) Del(keys ...string) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range keys {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Del", varargs...)
}

// Del indicates an expected call of Del
func (mr *MockCacheStoreMockRecorder) Del(keys ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del", reflect.TypeOf((*MockCacheStore)(nil).Del), keys...)
}

// HasKey mocks base method
func (m *MockCacheStore) HasKey(key string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasKey", key)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasKey indicates an expected call of HasKey
func (mr *MockCacheStoreMockRecorder) HasKey(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasKey", reflect.TypeOf((*MockCacheStore)(nil).HasKey), key)
}

// MockSessionStore is a mock of SessionStore interface
type MockSessionStore struct {
	ctrl     *gomock.Controller
	recorder *MockSessionStoreMockRecorder
}

// MockSessionStoreMockRecorder is the mock recorder for MockSessionStore
type MockSessionStoreMockRecorder struct {
	mock *MockSessionStore
}

// NewMockSessionStore creates a new mock instance
func NewMockSessionStore(ctrl *gomock.Controller) *MockSessionStore {
	mock := &MockSessionStore{ctrl: ctrl}
	mock.recorder = &MockSessionStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSessionStore) EXPECT() *MockSessionStoreMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", r, name)
	ret0, _ := ret[0].(*sessions.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockSessionStoreMockRecorder) Get(r, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSessionStore)(nil).Get), r, name)
}

// New mocks base method
func (m *MockSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New", r, name)
	ret0, _ := ret[0].(*sessions.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// New indicates an expected call of New
func (mr *MockSessionStoreMockRecorder) New(r, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockSessionStore)(nil).New), r, name)
}

// Save mocks base method
func (m *MockSessionStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", r, w, s)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save
func (mr *MockSessionStoreMockRecorder) Save(r, w, s interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockSessionStore)(nil).Save), r, w, s)
}

// Ping mocks base method
func (m *MockSessionStore) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping
func (mr *MockSessionStoreMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockSessionStore)(nil).Ping))
}
