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
	mock "github.com/stretchr/testify/mock"
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
		if pkgerrors.Cause(err).Error() != errFutureAndPastDateInternal.Error() {
			t.Errorf("should have errFutureAndPastDateInternal error\n")
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
		if pkgerrors.Cause(err).Error() != errFutureAndPastDateInternal.Error() {
			t.Errorf("should have errFutureAndPastDateInternal error\n")
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
				RecoverDB: func(err error) (*sqlx.DB, error) {
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

	newDB := &sqlx.DB{
		DB: db,
	}

	if _, err = GetFormSelections(
		rr,
		config,
		newDB,
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
		newDB,
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
		newDB,
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
		newDB,
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
		newDB,
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
		newDB,
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
		newDB,
		sqlx.DOLLAR,
		"",
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestValidatorRulesUnitTest(t *testing.T) {
	var err error

	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	queryErr := errors.New("query error")
	validateVal := 1

	newDB := &sqlx.DB{
		DB: db,
	}

	formValidator := &FormValidation{
		config: FormValidationConfig{},
	}
	cacheValidate := &CacheValidate{
		Key: "key",
	}
	validator := &validator{
		querier:        newDB,
		bindVar:        sqlx.DOLLAR,
		placeHolderIdx: -1,
		query:          "select id from foo where id = ?;",
		entityRecover:  formValidator,
		err:            errors.New(InvalidTxt),
	}
	singleRow := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mockCache := &MockCacheStore{}

	type foo struct {
		ID int64 `json:"id,string"`
	}

	cacheSetup := func(cacheVal interface{}, err error) {
		mockCache = &MockCacheStore{}
		mockCache.On("Get", mock.Anything, mock.Anything).Return(cacheVal, err)
	}
	dbSetup := func() {
		singleRow = sqlmock.NewRows([]string{"id"}).AddRow(1)
		mockDB.ExpectQuery("select").WillReturnRows(singleRow)
	}

	reset := func(cacheVal interface{}, err error) {
		cacheSetup(cacheVal, err)
		dbSetup()
		validator.err = errors.New(InvalidTxt)
	}

	//validator := &validateArgsRule{validator: validator}

	// -----------------------------------------------------------

	// Testing that if nil is passed, no errors occur
	if err = validatorRules(validator, nil, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	// -----------------------------------------------------------

	// Testing that empty slice will return no errors
	if err = validatorRules(validator, []interface{}{}, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	// -----------------------------------------------------------

	// Testing we have no cache and that db returns error
	mockDB.ExpectQuery("select").WillReturnError(queryErr)
	if err = validatorRules(validator, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != queryErr.Error() {
			t.Errorf("should have query error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	// -----------------------------------------------------------

	// Testing we get db error and recover but also fail
	// on second attempt
	mockDB.ExpectQuery("select").WillReturnError(queryErr)
	mockDB.ExpectQuery("select").WillReturnError(queryErr)
	validator.recoverDB = func(err error) (*sqlx.DB, error) {
		return &sqlx.DB{DB: db}, nil
	}

	if err = validatorRules(validator, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != queryErr.Error() {
			t.Errorf("should have query error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	// -----------------------------------------------------------

	// Testing a successful call to db with no errors
	mockDB.ExpectQuery("select").WillReturnRows(singleRow)

	if err = validatorRules(validator, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	// -----------------------------------------------------------

	// Testing that if we get multiple results for single
	// validation value, we get error
	multiRow := sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)
	mockDB.ExpectQuery("select").WillReturnRows(multiRow)

	if err = validatorRules(validator, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("should have '%s' error\n", InvalidTxt)
			t.Errorf("err: %s\n", err.Error())
		}
	}

	// -----------------------------------------------------------

	// Assign cache
	validator.cacheValidate = cacheValidate

	// -----------------------------------------------------------

	// Testing that if we get ErrCacheNil from cache and use
	// IgnoreCacheNil option, we query db and get no error
	reset(nil, ErrCacheNil)

	//mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, ErrCacheNil)
	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.IgnoreCacheNil = true

	if err = validatorRules(validator, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that if we get err from cache, we resort to db
	// and not get error
	reset(nil, errors.New("error"))

	//mockCache.On("Get", mock.Anything, mock.Anything).Return(nil, ErrCacheNil)
	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.IgnoreCacheNil = true

	if err = validatorRules(validator, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that if we don't ignore types, we will get error
	// even if the values in type are equal
	f := foo{ID: 1}
	valObjectBytes, err := json.Marshal(&f)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	cacheSetup(valObjectBytes, nil)

	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.KeyIdx = -1
	validator.placeHolderIdx = -1
	validator.cacheValidate.PropertyName = "id"

	if err = validatorRules(validator, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("should have '%s' error\n", InvalidTxt)
		}
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that if we ignore types, we will not get error
	cacheSetup(valObjectBytes, nil)
	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.PropertyName = "id"
	validator.validateConf.IgnoreTypes = true
	validator.placeHolderIdx = 0

	if err = validatorRules(validator, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that when we pass slice values for validation and
	// receive array from cache and they equal each other, we
	// don't get error
	validateSliceVal := []int{1}
	fSlice := []foo{
		{
			ID: 1,
		},
	}

	valObjectSliceBytes, err := json.Marshal(&fSlice)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	cacheSetup(valObjectSliceBytes, nil)

	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.PropertyName = "id"

	if err = validatorRules(validator, validateSliceVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that if we have wrong property type and
	// set IgnoreInvalidCacheResults to false, we get error
	cacheSetup(valObjectBytes, nil)
	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.PropertyName = "invalid"

	if err = validatorRules(validator, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("should have err: %s\n", err.Error())
		}
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name but IgnoreInvalidCacheResults is set for
	// validateArgsType that we resort to db and not get error
	reset(valObjectSliceBytes, nil)

	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.PropertyName = "invalid"
	validator.cacheValidate.IgnoreInvalidCacheResults = true

	if err = validatorRules(validator, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name but IgnoreInvalidCacheResults is set false
	// so we get error
	reset(valObjectSliceBytes, nil)

	validator.cache = mockCache
	validator.err = errors.New(DoesNotExistTxt)
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.PropertyName = "invalid"
	validator.cacheValidate.IgnoreInvalidCacheResults = false

	if err = validatorRules(validator, validateVal, validateExistsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != DoesNotExistTxt {
			t.Errorf("should have '%s' error\n", DoesNotExistTxt)
		}
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name but IgnoreInvalidCacheResults is set
	// for validateExistsType that we resort to db and not get error
	reset(valObjectSliceBytes, nil)

	validator.cache = mockCache
	validator.err = errors.New(DoesNotExistTxt)
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.PropertyName = "invalid"
	validator.cacheValidate.IgnoreInvalidCacheResults = true

	if err = validatorRules(validator, validateVal, validateExistsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing when we get array cache results and have valid
	// property name but IgnoreInvalidCacheResults is set false
	// we get error
	cacheSetup(valObjectSliceBytes, nil)
	emptyRow := sqlmock.NewRows([]string{"id"})
	mockDB.ExpectQuery("select").WillReturnRows(emptyRow)

	validator.cache = mockCache
	validator.err = errors.New(AlreadyExistsTxt)
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.PropertyName = "id"

	if err = validatorRules(validator, validateVal, validateUniquenessType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name but IgnoreInvalidCacheResults is set
	// for validateUniquenessType that we resort to db and not get error
	reset(valObjectSliceBytes, nil)

	validator.cache = mockCache
	validator.err = errors.New(AlreadyExistsTxt)
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.PropertyName = "invalid"
	validator.cacheValidate.IgnoreInvalidCacheResults = true

	if err = validatorRules(validator, validateVal, validateUniquenessType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that if we get ErrCacheNil error and IgnoreCacheNil is false,
	// we get error
	reset(nil, ErrCacheNil)

	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.cacheValidate.IgnoreCacheNil = false

	if err = validatorRules(validator, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("error should be '%s'", InvalidTxt)
		}
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that array cache results that is not objects
	// with IgnoreTypes set will not return error
	singleSliceVals := []int{1}
	singleSliceBytes, err := json.Marshal(singleSliceVals)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	reset(singleSliceBytes, nil)

	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.validateConf.IgnoreTypes = true

	if err = validatorRules(validator, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that array cache results that is not objects
	// with IgnoreTypes set false will return error
	reset(singleSliceBytes, nil)

	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.validateConf.IgnoreTypes = false

	if err = validatorRules(validator, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("should have '%s' error\n", err.Error())
		}
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that bool is not a proper type and results in error
	boolSlice := []bool{true}
	boolSliceBytes, err := json.Marshal(boolSlice)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	reset(boolSliceBytes, nil)

	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate

	if err = validatorRules(validator, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != errInvalidCacheTypeInternal.Error() {
			t.Errorf("err type should be 'errInvalidCacheTypeInternal'")
		}
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that cache result that is not array or object is
	// valid and does not return error
	singleValBytes, err := json.Marshal(1)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	reset(singleValBytes, nil)

	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.validateConf.IgnoreTypes = false

	if err = validatorRules(validator, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %v\n", err.Error())
	}

	mockCache.AssertExpectations(t)

	// -----------------------------------------------------------

	// Testing that cache result that is not array or object with
	// IgnoreTypes set is valid and does not return error
	reset(singleValBytes, nil)

	validator.cache = mockCache
	*validator.cacheValidate = *cacheValidate
	validator.validateConf.IgnoreTypes = true

	if err = validatorRules(validator, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	mockCache.AssertExpectations(t)
}
