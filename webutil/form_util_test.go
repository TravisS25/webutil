package webutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/jmoiron/sqlx"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	testifymock "github.com/stretchr/testify/mock"
)

func TestValidateRequiredRuleUnitTest(t *testing.T) {
	var err error

	rule := &validateRequiredRule{err: errors.New(RequiredTxt)}

	if err = rule.Validate(nil); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != RequiredTxt {
			t.Errorf("should have returned required error\n")
		}
	}

	if err = rule.Validate(""); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != RequiredTxt {
			t.Errorf("should have ErrRequiredValidator error\n")
		}
	}

	strVal := "val"

	if err = rule.Validate(strVal); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = rule.Validate(&strVal); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestValidateDateRuleUnitTest(t *testing.T) {
	var err error

	futureDateStr := time.Now().AddDate(0, 0, 1).Format(DateTimeLayout)
	pastDateStr := time.Now().AddDate(0, 0, -1).Format(DateTimeLayout)
	rule := &validateDateRule{timezone: "invalid"}

	if err = rule.Validate(nil); err != nil {
		t.Errorf("should not have error\n")
	}

	if err = rule.Validate(""); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = rule.Validate("invalid"); err == nil {
		t.Errorf("should have error\n")
	}

	rule.timezone = ""

	if err = rule.Validate("invalid"); err == nil {
		t.Errorf("should have error\n")
	} else {
		if pkgerrors.Cause(err).Error() != InvalidFormatTxt {
			t.Errorf("should have ErrInvalidFormatValidator error\n")
		}
	}

	rule.layout = DateTimeLayout

	if err = rule.Validate(pastDateStr); err == nil {
		t.Errorf("should have error\n")
	} else {
		if pkgerrors.Cause(err).Error() != ErrFutureAndPastDateInternal.Error() {
			t.Errorf("should have ErrFutureAndPastDateInternal error\n")
			t.Errorf("got err: %s\n", pkgerrors.Cause(err))
		}
	}

	rule.canBeFuture = true
	rule.canBePast = true

	if err = rule.Validate(pastDateStr); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	rule.canBePast = false

	if err = rule.Validate(pastDateStr); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidPastDateTxt {
			t.Errorf("should have ErrInvalidPastDateValidator error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	rule.canBePast = true
	rule.canBeFuture = false

	if err = rule.Validate(futureDateStr); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidFutureDateTxt {
			t.Errorf("should have ErrInvalidFutureDateValidator error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	rule.canBePast = false

	if err = rule.Validate(futureDateStr); err == nil {
		t.Errorf("should have error\n")
	} else {
		if pkgerrors.Cause(err).Error() != ErrFutureAndPastDateInternal.Error() {
			t.Errorf("should have ErrFutureAndPastDateInternal error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}
}

func TestCheckIfExistsUnitTest(t *testing.T) {
	var err error

	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	mockCacheStore1 := &MockCacheStore{}
	defer mockCacheStore1.AssertExpectations(t)
	mockCacheStore1.On("Get", testifymock.Anything).Return([]byte("foo"), nil)

	//mockCacheStore.EXPECT().Get(gomock.Any()).Return([]byte("foo"), nil)

	rule := &validator{
		querier: db,
		cache:   mockCacheStore1,
		cacheValidateKey: &CacheValidateKey{
			IgnoreCacheNil: true,
		},
		args: []interface{}{1},
		err:  errors.New(DoesNotExistTxt),
	}

	if err = checkIfExists(rule, "foo", false); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != DoesNotExistTxt {
			t.Errorf("should have ErrDoesNotExistValidator error\n")
		}
	}

	mockCacheStore2 := &MockCacheStore{}
	defer mockCacheStore2.AssertExpectations(t)
	rule.cache = mockCacheStore2
	mockCacheStore2.On("Get", testifymock.Anything).Return(nil, ErrCacheNil)
	mockDB.ExpectQuery("select").WillReturnError(ErrServer)

	if err = checkIfExists(rule, "foo", false); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != ErrServer.Error() {
			t.Errorf("should have ErrServer error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	mockCacheStore3 := &MockCacheStore{}
	defer mockCacheStore3.AssertExpectations(t)
	rule.cache = mockCacheStore3
	rows := sqlmock.NewRows([]string{"value", "text"})
	mockCacheStore3.On("Get", testifymock.Anything).Return(nil, ErrCacheNil)
	mockDB.ExpectQuery("select").WillReturnRows(rows)

	if err = checkIfExists(rule, "foo", true); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != DoesNotExistTxt {
			t.Errorf("should have ErrServer error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	mockCacheStore4 := &MockCacheStore{}
	defer mockCacheStore4.AssertExpectations(t)
	rule.cache = mockCacheStore4
	mockCacheStore4.On("Get", testifymock.Anything).Return(nil, ErrCacheNil)
	mockDB.ExpectQuery("select").WillReturnError(ErrServer)

	if err = checkIfExists(rule, "foo", true); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != ErrServer.Error() {
			t.Errorf("should have ErrServer error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	mockCacheStore5 := &MockCacheStore{}
	defer mockCacheStore5.AssertExpectations(t)
	rule.cache = mockCacheStore5
	mockCacheStore5.On("Get", testifymock.Anything).Return(nil, ErrServer)
	mockDB.ExpectQuery("select").WillReturnError(ErrServer)

	if err = checkIfExists(rule, "foo", true); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != ErrServer.Error() {
			t.Errorf("should have ErrServer error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	mockCacheStore6 := &MockCacheStore{}
	defer mockCacheStore6.AssertExpectations(t)
	rule.cache = mockCacheStore6
	// rule.recoverDB = func(err error) error {
	// 	return ErrServer
	// }
	mockCacheStore6.On("Get", testifymock.Anything).Return(nil, ErrServer)
	mockDB.ExpectQuery("select").WillReturnError(ErrServer)

	if err = checkIfExists(rule, "foo", true); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != ErrServer.Error() {
			t.Errorf("should have ErrServer error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}
}

func TestHasFormErrorsUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()

	if !HasFormErrors(rr, ErrBodyRequired, ServerErrorConfig{}) {
		t.Errorf("should have form error\n")
	}
	if rr.Result().StatusCode != http.StatusNotAcceptable {
		t.Errorf("returned status should be 406\n")
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(rr.Result().Body)
	rr.Result().Body.Close()

	if buf.String() != ErrBodyRequired.Error() {
		t.Errorf("error response should be %s\n", ErrBodyRequired.Error())
	}
	if rr.Result().StatusCode != http.StatusNotAcceptable {
		t.Errorf("returned status should be 406\n")
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	HasFormErrors(rr, ErrInvalidJSON, ServerErrorConfig{})
	buf.ReadFrom(rr.Result().Body)
	rr.Result().Body.Close()

	if buf.String() != ErrInvalidJSON.Error() {
		t.Errorf("error response should be %s\n", ErrInvalidJSON.Error())
	}
	if rr.Result().StatusCode != http.StatusNotAcceptable {
		t.Errorf("returned status should be 406\n")
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	vErr := validation.Errors{
		"id": errors.New("field error"),
	}

	HasFormErrors(rr, vErr, ServerErrorConfig{})
	buf.ReadFrom(rr.Result().Body)
	rr.Result().Body.Close()

	vBytes, err := json.Marshal(vErr)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	if buf.String() != string(vBytes) {
		t.Errorf("error response should be %s\n", string(vBytes))
		t.Errorf("err: %s\n", string(vBytes))
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	HasFormErrors(rr, errors.New("errors"), ServerErrorConfig{})
	buf.ReadFrom(rr.Result().Body)

	if rr.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("returned status should be 500\n")
	}
	if buf.String() != ErrServer.Error() {
		t.Errorf("error response should be %s\n", ErrServer.Error())
		t.Errorf("err: %s\n", buf.String())
	}
}

func TestCheckBodyAndDecodeUnitTest(t *testing.T) {
	var err error

	type foo struct {
		ID   int64
		Name string
	}

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	exMethods := []string{http.MethodPost}

	f := foo{}

	if err = CheckBodyAndDecode(req, f); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != ErrBodyRequired.Error() {
			t.Errorf("should have ErrBodyRequired error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	if err = CheckBodyAndDecode(req, f, exMethods...); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	buf := &bytes.Buffer{}
	req = httptest.NewRequest(http.MethodPost, "/url", buf)

	if err = CheckBodyAndDecode(req, errors.New("error")); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != ErrInvalidJSON.Error() {
			t.Errorf("should have ErrInvalidJSON error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}
}

func TestGetFormSelectionsUnitTest(t *testing.T) {
	var err error

	rr := httptest.NewRecorder()
	buf := &bytes.Buffer{}
	config := ServerErrorCacheConfig{
		ServerErrorConfig: ServerErrorConfig{
			RecoverConfig: RecoverConfig{
				RecoverDB: func(err error) (*DB, error) {
					return nil, ErrServer
				},
			},
		},
	}
	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	mockDB.ExpectQuery("select").WillReturnError(ErrServer)

	if _, err = GetFormSelections(
		rr,
		config,
		db,
		sqlx.DOLLAR,
		"",
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		buf.ReadFrom(rr.Result().Body)
		if rr.Result().StatusCode != http.StatusInternalServerError {
			t.Errorf("should have 500 error\n")
		}
		if buf.String() != ErrServer.Error() {
			t.Errorf("should have ErrServer error\n")
		}
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	config.ServerErrorConfig.RecoverDB = nil
	mockDB.ExpectQuery("select").WillReturnError(ErrServer)

	if _, err = GetFormSelections(
		rr,
		config,
		db,
		sqlx.DOLLAR,
		"",
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		buf.ReadFrom(rr.Result().Body)
		if rr.Result().StatusCode != http.StatusInternalServerError {
			t.Errorf("should have 500 error\n")
		}
		if buf.String() != ErrServer.Error() {
			t.Errorf("should have ErrServer error\n")
		}
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	rows := sqlmock.NewRows([]string{"value", "text"}).
		AddRow(1, "foo").
		AddRow(2, "bar")
	mockDB.ExpectQuery("").WillReturnRows(rows)

	if _, err = GetFormSelections(
		rr,
		config,
		db,
		sqlx.DOLLAR,
		"",
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	mockCacheStore1 := &MockCacheStore{}
	defer mockCacheStore1.AssertExpectations(t)
	config.CacheConfig.Cache = mockCacheStore1
	mockCacheStore1.On("Get", testifymock.Anything).Return(nil, ErrServer)
	mockDB.ExpectQuery("").WillReturnRows(rows)

	if _, err = GetFormSelections(
		rr,
		config,
		db,
		sqlx.DOLLAR,
		"",
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	mockCacheStore2 := &MockCacheStore{}
	defer mockCacheStore2.AssertExpectations(t)
	config.CacheConfig.Cache = mockCacheStore2
	config.CacheConfig.IgnoreCacheNil = true
	mockCacheStore2.On("Get", testifymock.Anything).Return(nil, ErrCacheNil)
	mockDB.ExpectQuery("").WillReturnRows(rows)

	if _, err = GetFormSelections(
		rr,
		config,
		db,
		sqlx.DOLLAR,
		"",
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	mockCacheStore3 := &MockCacheStore{}
	defer mockCacheStore3.AssertExpectations(t)
	config.CacheConfig.Cache = mockCacheStore3
	config.CacheConfig.IgnoreCacheNil = false
	mockCacheStore3.On("Get", testifymock.Anything).Return(nil, ErrCacheNil)

	if _, err = GetFormSelections(
		rr,
		config,
		db,
		sqlx.DOLLAR,
		"",
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if rr.Result().StatusCode != http.StatusInternalServerError {
			t.Errorf("should have StatusInternalServerError error\n")
		}
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	form := []FormSelection{
		{
			Value: 1, Text: "foo",
		},
	}

	formBytes, err := json.Marshal(&form)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	mockCacheStore4 := &MockCacheStore{}
	defer mockCacheStore4.AssertExpectations(t)
	config.CacheConfig.Cache = mockCacheStore4
	mockCacheStore4.On("Get", testifymock.Anything).Return(formBytes, nil)

	if _, err = GetFormSelections(
		rr,
		config,
		db,
		sqlx.DOLLAR,
		"",
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestValidateIDsRuleUnitTest(t *testing.T) {
	var err error

	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	queryErr := errors.New("query error")
	validateVal := []interface{}{"1"}

	mockCache1 := &MockCacheStore{}
	defer mockCache1.AssertExpectations(t)

	mockDB.ExpectQuery("select").WillReturnError(queryErr)

	formValidator := &FormValidation{}
	validator := &validator{
		querier:        db,
		cache:          mockCache1,
		bindVar:        sqlx.DOLLAR,
		placeHolderIdx: -1,
		entityRecover:  formValidator,
		err:            errors.New(InvalidTxt),
	}

	idValidator := &validateIDsRule{validator: validator}

	if err = idValidator.Validate(nil); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = idValidator.Validate([]interface{}{}); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = idValidator.Validate(validateVal); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != queryErr.Error() {
			t.Errorf("should have query error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	mockDB.ExpectQuery("select").WillReturnError(queryErr)
	mockDB.ExpectQuery("select").WillReturnError(queryErr)
	idValidator.recoverDB = func(err error) (*DB, error) {
		return &DB{DB: &sqlx.DB{DB: db}}, nil
	}

	if err = idValidator.Validate(validateVal); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != queryErr.Error() {
			t.Errorf("should have query error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mockDB.ExpectQuery("select").WillReturnRows(rows)

	if err = idValidator.Validate(validateVal); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	rows = sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)
	mockDB.ExpectQuery("select").WillReturnRows(rows)

	if err = idValidator.Validate(validateVal); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("should have invalid error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	mockCache1.On("Get", mock.Anything, mock.Anything).Return(nil, ErrCacheNil)
	idValidator.cacheValidateKey = &CacheValidateKey{
		Key:            "key",
		IgnoreCacheNil: true,
	}

	rows = sqlmock.NewRows([]string{"id"}).AddRow(1)
	mockDB.ExpectQuery("select").WillReturnRows(rows)

	if err = idValidator.Validate(validateVal); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	valBytes, err := json.Marshal(&validateVal)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	mockCache2 := &MockCacheStore{}
	defer mockCache2.AssertExpectations(t)
	mockCache2.On("Get", mock.Anything, mock.Anything).Return(valBytes, nil)

	idValidator.cache = mockCache2

	if err = idValidator.Validate(validateVal); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	idValidator.cache = nil
	idValidator.query = "select id from something where id = ?"
	idValidator.placeHolderIdx = 0
	rows = sqlmock.NewRows([]string{"id"}).AddRow(1)
	mockDB.ExpectQuery("select").WillReturnRows(rows)

	if err = idValidator.Validate(1); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}
