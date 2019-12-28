// Code generated by MockGen. DO NOT EDIT.
// Source: db_util.go

// Package webutilmock is a generated GoMock package.
package webutilmock

import (
	sql "database/sql"
	webutil "github.com/TravisS25/webutil/webutil"
	gomock "github.com/golang/mock/gomock"
	sqlx "github.com/jmoiron/sqlx"
	reflect "reflect"
)

// MockQuerier is a mock of Querier interface
type MockQuerier struct {
	ctrl     *gomock.Controller
	recorder *MockQuerierMockRecorder
}

// MockQuerierMockRecorder is the mock recorder for MockQuerier
type MockQuerierMockRecorder struct {
	mock *MockQuerier
}

// NewMockQuerier creates a new mock instance
func NewMockQuerier(ctrl *gomock.Controller) *MockQuerier {
	mock := &MockQuerier{ctrl: ctrl}
	mock.recorder = &MockQuerierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockQuerier) EXPECT() *MockQuerierMockRecorder {
	return m.recorder
}

// QueryRow mocks base method
func (m *MockQuerier) QueryRow(query string, args ...interface{}) *sqlx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(*sqlx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow
func (mr *MockQuerierMockRecorder) QueryRow(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockQuerier)(nil).QueryRow), varargs...)
}

// Query mocks base method
func (m *MockQuerier) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sqlx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockQuerierMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockQuerier)(nil).Query), varargs...)
}

// MockScanner is a mock of Scanner interface
type MockScanner struct {
	ctrl     *gomock.Controller
	recorder *MockScannerMockRecorder
}

// MockScannerMockRecorder is the mock recorder for MockScanner
type MockScannerMockRecorder struct {
	mock *MockScanner
}

// NewMockScanner creates a new mock instance
func NewMockScanner(ctrl *gomock.Controller) *MockScanner {
	mock := &MockScanner{ctrl: ctrl}
	mock.recorder = &MockScannerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockScanner) EXPECT() *MockScannerMockRecorder {
	return m.recorder
}

// Scan mocks base method
func (m *MockScanner) Scan(dest ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range dest {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Scan", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan
func (mr *MockScannerMockRecorder) Scan(dest ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*MockScanner)(nil).Scan), dest...)
}

// MockRower is a mock of Rower interface
type MockRower struct {
	ctrl     *gomock.Controller
	recorder *MockRowerMockRecorder
}

// MockRowerMockRecorder is the mock recorder for MockRower
type MockRowerMockRecorder struct {
	mock *MockRower
}

// NewMockRower creates a new mock instance
func NewMockRower(ctrl *gomock.Controller) *MockRower {
	mock := &MockRower{ctrl: ctrl}
	mock.recorder = &MockRowerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRower) EXPECT() *MockRowerMockRecorder {
	return m.recorder
}

// Scan mocks base method
func (m *MockRower) Scan(dest ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range dest {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Scan", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan
func (mr *MockRowerMockRecorder) Scan(dest ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*MockRower)(nil).Scan), dest...)
}

// Next mocks base method
func (m *MockRower) Next() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Next")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Next indicates an expected call of Next
func (mr *MockRowerMockRecorder) Next() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Next", reflect.TypeOf((*MockRower)(nil).Next))
}

// Columns mocks base method
func (m *MockRower) Columns() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Columns")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Columns indicates an expected call of Columns
func (mr *MockRowerMockRecorder) Columns() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Columns", reflect.TypeOf((*MockRower)(nil).Columns))
}

// MockTx is a mock of Tx interface
type MockTx struct {
	ctrl     *gomock.Controller
	recorder *MockTxMockRecorder
}

// MockTxMockRecorder is the mock recorder for MockTx
type MockTxMockRecorder struct {
	mock *MockTx
}

// NewMockTx creates a new mock instance
func NewMockTx(ctrl *gomock.Controller) *MockTx {
	mock := &MockTx{ctrl: ctrl}
	mock.recorder = &MockTxMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTx) EXPECT() *MockTxMockRecorder {
	return m.recorder
}

// QueryRow mocks base method
func (m *MockTx) QueryRow(query string, args ...interface{}) *sqlx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(*sqlx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow
func (mr *MockTxMockRecorder) QueryRow(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockTx)(nil).QueryRow), varargs...)
}

// Query mocks base method
func (m *MockTx) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sqlx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockTxMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockTx)(nil).Query), varargs...)
}

// Exec mocks base method
func (m *MockTx) Exec(arg0 string, arg1 ...interface{}) (sql.Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockTxMockRecorder) Exec(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockTx)(nil).Exec), varargs...)
}

// Get mocks base method
func (m *MockTx) Get(dest interface{}, query string, args ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{dest, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Get indicates an expected call of Get
func (mr *MockTxMockRecorder) Get(dest, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{dest, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockTx)(nil).Get), varargs...)
}

// Select mocks base method
func (m *MockTx) Select(dest interface{}, query string, args ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{dest, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Select", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Select indicates an expected call of Select
func (mr *MockTxMockRecorder) Select(dest, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{dest, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Select", reflect.TypeOf((*MockTx)(nil).Select), varargs...)
}

// Commit mocks base method
func (m *MockTx) Commit() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit
func (mr *MockTxMockRecorder) Commit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockTx)(nil).Commit))
}

// Rollback mocks base method
func (m *MockTx) Rollback() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rollback")
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback
func (mr *MockTxMockRecorder) Rollback() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockTx)(nil).Rollback))
}

// MockTransaction is a mock of Transaction interface
type MockTransaction struct {
	ctrl     *gomock.Controller
	recorder *MockTransactionMockRecorder
}

// MockTransactionMockRecorder is the mock recorder for MockTransaction
type MockTransactionMockRecorder struct {
	mock *MockTransaction
}

// NewMockTransaction creates a new mock instance
func NewMockTransaction(ctrl *gomock.Controller) *MockTransaction {
	mock := &MockTransaction{ctrl: ctrl}
	mock.recorder = &MockTransactionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTransaction) EXPECT() *MockTransactionMockRecorder {
	return m.recorder
}

// Begin mocks base method
func (m *MockTransaction) Begin() (*sqlx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin")
	ret0, _ := ret[0].(*sqlx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin
func (mr *MockTransactionMockRecorder) Begin() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*MockTransaction)(nil).Begin))
}

// Commit mocks base method
func (m *MockTransaction) Commit(tx *sqlx.Tx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", tx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit
func (mr *MockTransactionMockRecorder) Commit(tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockTransaction)(nil).Commit), tx)
}

// MockQueryTransaction is a mock of QueryTransaction interface
type MockQueryTransaction struct {
	ctrl     *gomock.Controller
	recorder *MockQueryTransactionMockRecorder
}

// MockQueryTransactionMockRecorder is the mock recorder for MockQueryTransaction
type MockQueryTransactionMockRecorder struct {
	mock *MockQueryTransaction
}

// NewMockQueryTransaction creates a new mock instance
func NewMockQueryTransaction(ctrl *gomock.Controller) *MockQueryTransaction {
	mock := &MockQueryTransaction{ctrl: ctrl}
	mock.recorder = &MockQueryTransactionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockQueryTransaction) EXPECT() *MockQueryTransactionMockRecorder {
	return m.recorder
}

// Begin mocks base method
func (m *MockQueryTransaction) Begin() (*sqlx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin")
	ret0, _ := ret[0].(*sqlx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin
func (mr *MockQueryTransactionMockRecorder) Begin() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*MockQueryTransaction)(nil).Begin))
}

// Commit mocks base method
func (m *MockQueryTransaction) Commit(tx *sqlx.Tx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", tx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit
func (mr *MockQueryTransactionMockRecorder) Commit(tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockQueryTransaction)(nil).Commit), tx)
}

// QueryRow mocks base method
func (m *MockQueryTransaction) QueryRow(query string, args ...interface{}) *sqlx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(*sqlx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow
func (mr *MockQueryTransactionMockRecorder) QueryRow(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockQueryTransaction)(nil).QueryRow), varargs...)
}

// Query mocks base method
func (m *MockQueryTransaction) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sqlx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockQueryTransactionMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockQueryTransaction)(nil).Query), varargs...)
}

// MockXODB is a mock of XODB interface
type MockXODB struct {
	ctrl     *gomock.Controller
	recorder *MockXODBMockRecorder
}

// MockXODBMockRecorder is the mock recorder for MockXODB
type MockXODBMockRecorder struct {
	mock *MockXODB
}

// NewMockXODB creates a new mock instance
func NewMockXODB(ctrl *gomock.Controller) *MockXODB {
	mock := &MockXODB{ctrl: ctrl}
	mock.recorder = &MockXODBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockXODB) EXPECT() *MockXODBMockRecorder {
	return m.recorder
}

// QueryRow mocks base method
func (m *MockXODB) QueryRow(query string, args ...interface{}) *sqlx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(*sqlx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow
func (mr *MockXODBMockRecorder) QueryRow(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockXODB)(nil).QueryRow), varargs...)
}

// Query mocks base method
func (m *MockXODB) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sqlx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockXODBMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockXODB)(nil).Query), varargs...)
}

// Exec mocks base method
func (m *MockXODB) Exec(arg0 string, arg1 ...interface{}) (sql.Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockXODBMockRecorder) Exec(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockXODB)(nil).Exec), varargs...)
}

// MockSqlxDB is a mock of SqlxDB interface
type MockSqlxDB struct {
	ctrl     *gomock.Controller
	recorder *MockSqlxDBMockRecorder
}

// MockSqlxDBMockRecorder is the mock recorder for MockSqlxDB
type MockSqlxDBMockRecorder struct {
	mock *MockSqlxDB
}

// NewMockSqlxDB creates a new mock instance
func NewMockSqlxDB(ctrl *gomock.Controller) *MockSqlxDB {
	mock := &MockSqlxDB{ctrl: ctrl}
	mock.recorder = &MockSqlxDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSqlxDB) EXPECT() *MockSqlxDBMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockSqlxDB) Get(dest interface{}, query string, args ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{dest, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Get indicates an expected call of Get
func (mr *MockSqlxDBMockRecorder) Get(dest, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{dest, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSqlxDB)(nil).Get), varargs...)
}

// Select mocks base method
func (m *MockSqlxDB) Select(dest interface{}, query string, args ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{dest, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Select", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Select indicates an expected call of Select
func (mr *MockSqlxDBMockRecorder) Select(dest, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{dest, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Select", reflect.TypeOf((*MockSqlxDB)(nil).Select), varargs...)
}

// MockEntity is a mock of Entity interface
type MockEntity struct {
	ctrl     *gomock.Controller
	recorder *MockEntityMockRecorder
}

// MockEntityMockRecorder is the mock recorder for MockEntity
type MockEntityMockRecorder struct {
	mock *MockEntity
}

// NewMockEntity creates a new mock instance
func NewMockEntity(ctrl *gomock.Controller) *MockEntity {
	mock := &MockEntity{ctrl: ctrl}
	mock.recorder = &MockEntityMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEntity) EXPECT() *MockEntityMockRecorder {
	return m.recorder
}

// QueryRow mocks base method
func (m *MockEntity) QueryRow(query string, args ...interface{}) *sqlx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(*sqlx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow
func (mr *MockEntityMockRecorder) QueryRow(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockEntity)(nil).QueryRow), varargs...)
}

// Query mocks base method
func (m *MockEntity) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sqlx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockEntityMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockEntity)(nil).Query), varargs...)
}

// Exec mocks base method
func (m *MockEntity) Exec(arg0 string, arg1 ...interface{}) (sql.Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockEntityMockRecorder) Exec(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockEntity)(nil).Exec), varargs...)
}

// Get mocks base method
func (m *MockEntity) Get(dest interface{}, query string, args ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{dest, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Get indicates an expected call of Get
func (mr *MockEntityMockRecorder) Get(dest, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{dest, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockEntity)(nil).Get), varargs...)
}

// Select mocks base method
func (m *MockEntity) Select(dest interface{}, query string, args ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{dest, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Select", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Select indicates an expected call of Select
func (mr *MockEntityMockRecorder) Select(dest, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{dest, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Select", reflect.TypeOf((*MockEntity)(nil).Select), varargs...)
}

// MockDBInterface is a mock of DBInterface interface
type MockDBInterface struct {
	ctrl     *gomock.Controller
	recorder *MockDBInterfaceMockRecorder
}

// MockDBInterfaceMockRecorder is the mock recorder for MockDBInterface
type MockDBInterfaceMockRecorder struct {
	mock *MockDBInterface
}

// NewMockDBInterface creates a new mock instance
func NewMockDBInterface(ctrl *gomock.Controller) *MockDBInterface {
	mock := &MockDBInterface{ctrl: ctrl}
	mock.recorder = &MockDBInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDBInterface) EXPECT() *MockDBInterfaceMockRecorder {
	return m.recorder
}

// QueryRow mocks base method
func (m *MockDBInterface) QueryRow(query string, args ...interface{}) *sqlx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(*sqlx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow
func (mr *MockDBInterfaceMockRecorder) QueryRow(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockDBInterface)(nil).QueryRow), varargs...)
}

// Query mocks base method
func (m *MockDBInterface) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sqlx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockDBInterfaceMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockDBInterface)(nil).Query), varargs...)
}

// Exec mocks base method
func (m *MockDBInterface) Exec(arg0 string, arg1 ...interface{}) (sql.Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockDBInterfaceMockRecorder) Exec(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockDBInterface)(nil).Exec), varargs...)
}

// Get mocks base method
func (m *MockDBInterface) Get(dest interface{}, query string, args ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{dest, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Get indicates an expected call of Get
func (mr *MockDBInterfaceMockRecorder) Get(dest, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{dest, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDBInterface)(nil).Get), varargs...)
}

// Select mocks base method
func (m *MockDBInterface) Select(dest interface{}, query string, args ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{dest, query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Select", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Select indicates an expected call of Select
func (mr *MockDBInterfaceMockRecorder) Select(dest, query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{dest, query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Select", reflect.TypeOf((*MockDBInterface)(nil).Select), varargs...)
}

// Begin mocks base method
func (m *MockDBInterface) Begin() (*sqlx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin")
	ret0, _ := ret[0].(*sqlx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin
func (mr *MockDBInterfaceMockRecorder) Begin() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*MockDBInterface)(nil).Begin))
}

// Commit mocks base method
func (m *MockDBInterface) Commit(tx *sqlx.Tx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", tx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit
func (mr *MockDBInterfaceMockRecorder) Commit(tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockDBInterface)(nil).Commit), tx)
}

// RecoverError mocks base method
func (m *MockDBInterface) RecoverError(err error) (*webutil.DB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecoverError", err)
	ret0, _ := ret[0].(*webutil.DB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RecoverError indicates an expected call of RecoverError
func (mr *MockDBInterfaceMockRecorder) RecoverError(err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecoverError", reflect.TypeOf((*MockDBInterface)(nil).RecoverError), err)
}

// MockRecover is a mock of Recover interface
type MockRecover struct {
	ctrl     *gomock.Controller
	recorder *MockRecoverMockRecorder
}

// MockRecoverMockRecorder is the mock recorder for MockRecover
type MockRecoverMockRecorder struct {
	mock *MockRecover
}

// NewMockRecover creates a new mock instance
func NewMockRecover(ctrl *gomock.Controller) *MockRecover {
	mock := &MockRecover{ctrl: ctrl}
	mock.recorder = &MockRecoverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRecover) EXPECT() *MockRecoverMockRecorder {
	return m.recorder
}

// RecoverError mocks base method
func (m *MockRecover) RecoverError(err error) (*webutil.DB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecoverError", err)
	ret0, _ := ret[0].(*webutil.DB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RecoverError indicates an expected call of RecoverError
func (mr *MockRecoverMockRecorder) RecoverError(err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecoverError", reflect.TypeOf((*MockRecover)(nil).RecoverError), err)
}

// MockRecoverQuerier is a mock of RecoverQuerier interface
type MockRecoverQuerier struct {
	ctrl     *gomock.Controller
	recorder *MockRecoverQuerierMockRecorder
}

// MockRecoverQuerierMockRecorder is the mock recorder for MockRecoverQuerier
type MockRecoverQuerierMockRecorder struct {
	mock *MockRecoverQuerier
}

// NewMockRecoverQuerier creates a new mock instance
func NewMockRecoverQuerier(ctrl *gomock.Controller) *MockRecoverQuerier {
	mock := &MockRecoverQuerier{ctrl: ctrl}
	mock.recorder = &MockRecoverQuerierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRecoverQuerier) EXPECT() *MockRecoverQuerierMockRecorder {
	return m.recorder
}

// QueryRow mocks base method
func (m *MockRecoverQuerier) QueryRow(query string, args ...interface{}) *sqlx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(*sqlx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow
func (mr *MockRecoverQuerierMockRecorder) QueryRow(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockRecoverQuerier)(nil).QueryRow), varargs...)
}

// Query mocks base method
func (m *MockRecoverQuerier) Query(query string, args ...interface{}) (*sqlx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sqlx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockRecoverQuerierMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockRecoverQuerier)(nil).Query), varargs...)
}

// RecoverError mocks base method
func (m *MockRecoverQuerier) RecoverError(err error) (*webutil.DB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecoverError", err)
	ret0, _ := ret[0].(*webutil.DB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RecoverError indicates an expected call of RecoverError
func (mr *MockRecoverQuerierMockRecorder) RecoverError(err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecoverError", reflect.TypeOf((*MockRecoverQuerier)(nil).RecoverError), err)
}
