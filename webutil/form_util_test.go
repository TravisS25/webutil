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
			t.Errorf("should have required error\n")
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

	if err = rule.Validate([]string{"hi", "there", ""}); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != RequiredTxt {
			t.Errorf("should have required errors\n")
		}
	}

	if err = rule.Validate([]string{"hi", "there"}); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	hi := "hi"

	if err = rule.Validate([]interface{}{&hi, "there", 1}); err != nil {
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
	httpStatus := 406

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/url", nil)

	if !HasFormErrors(rr, req, ErrBodyRequired, nil, httpStatus, ServerErrorConfig{}) {
		t.Errorf("should have form error\n")
	}
	if rr.Result().StatusCode != http.StatusNotAcceptable {
		t.Errorf("returned status should be 406\n")
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(rr.Result().Body)
	rr.Result().Body.Close()

	if buf.String() != bodyRequiredTxt {
		t.Errorf("error response should be %s\n", bodyRequiredTxt)
	}
	if rr.Result().StatusCode != http.StatusNotAcceptable {
		t.Errorf("returned status should be 406\n")
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	HasFormErrors(rr, req, ErrInvalidJSON, nil, httpStatus, ServerErrorConfig{})
	buf.ReadFrom(rr.Result().Body)
	rr.Result().Body.Close()

	if buf.String() != invalidJSONTxt {
		t.Errorf("error response should be %s\n", invalidJSONTxt)
	}
	if rr.Result().StatusCode != http.StatusNotAcceptable {
		t.Errorf("returned status should be 406\n")
	}

	buf.Reset()
	rr = httptest.NewRecorder()
	vErr := validation.Errors{
		"id": errors.New("field error"),
	}

	HasFormErrors(rr, req, vErr, nil, httpStatus, ServerErrorConfig{})
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
	HasFormErrors(rr, req, errors.New("errors"), nil, httpStatus, ServerErrorConfig{})
	buf.ReadFrom(rr.Result().Body)

	if rr.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("returned status should be 500\n")
	}
	if buf.String() != serverErrTxt {
		t.Errorf("error response should be %s\n", serverErrTxt)
		t.Errorf("err: %s\n", buf.String())
	}
}

func TestCheckBodyAndDecodeUnitTest(t *testing.T) {
	var err error

	type idStruct struct {
		ID   int64
		Name string
	}

	req := httptest.NewRequest(http.MethodPost, "/url", nil)
	exMethods := []string{http.MethodPost}

	f := idStruct{}

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

func TestValidatorRulesUnitTest(t *testing.T) {
	var err error
	var row *sqlmock.Rows

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

	validatorS := &validator{
		querier:        newDB,
		bindVar:        sqlx.DOLLAR,
		placeHolderIdx: -1,
		query:          "select id from idStruct where id = ?;",
		entityRecover:  formValidator,
		err:            errors.New(InvalidTxt),
		cacheValidate: &CacheValidate{
			Key: "key",
		},
	}

	mockCache := &MockCacheStore{}

	type idStruct struct {
		ID int64 `json:"id,string"`
	}

	reset := func(cacheVal interface{}, cacheErr error, numOfRows int, dbErr error) {
		mockCache = &MockCacheStore{}
		mockCache.On("Get", mock.Anything, mock.Anything).Return(cacheVal, cacheErr)

		if dbErr != nil {
			mockDB.ExpectQuery("select").WillReturnError(dbErr)
		} else {
			if numOfRows > -1 {
				row = sqlmock.NewRows([]string{"id"})
				for i := 0; i < numOfRows; i++ {
					row.AddRow(i)
				}
				mockDB.ExpectQuery("select").WillReturnRows(row)
			}
		}

		validatorS.err = errors.New(InvalidTxt)
		validatorS.cache = mockCache
		validatorS.cacheValidate = &CacheValidate{Key: "key"}
	}

	assertExpectations := func() error {
		mockCache.AssertExpectations(t)
		return mockDB.ExpectationsWereMet()
	}

	// -----------------------------------------------------------

	// Testing that if nil is passed, no errors occur
	if err = validatorRules(validatorS, nil, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that empty slice will return no errors
	if err = validatorRules(validatorS, []interface{}{}, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = mockDB.ExpectationsWereMet(); err != nil {
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing we have no cache and that db returns error

	//mockDB.ExpectQuery("select").WillReturnError(queryErr)
	reset(nil, ErrCacheNil, -1, queryErr)
	validatorS.cacheValidate.IgnoreCacheNil = true

	if err = validatorRules(validatorS, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != queryErr.Error() {
			t.Errorf("should have query error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}
	// -----------------------------------------------------------

	// Testing we get db error and recover but also fail
	// on second attempt

	// mockDB.ExpectQuery("select").WillReturnError(queryErr)
	// mockDB.ExpectQuery("select").WillReturnError(queryErr)
	reset(nil, ErrCacheNil, 0, queryErr)
	reset(nil, ErrCacheNil, 0, queryErr)
	validatorS.cacheValidate.IgnoreCacheNil = true
	validatorS.recoverDB = func(err error) (*sqlx.DB, error) {
		return &sqlx.DB{DB: db}, nil
	}

	if err = validatorRules(validatorS, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != queryErr.Error() {
			t.Errorf("should have query error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that if we get multiple results for single
	// validation value, we get error

	reset(nil, ErrCacheNil, 2, nil)
	validatorS.cacheValidate.IgnoreCacheNil = true

	if err = validatorRules(validatorS, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("should have '%s' error\n", InvalidTxt)
			t.Errorf("err: %s\n", err.Error())
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that if we get err from cache, we resort to db
	// and not get error
	reset(nil, errors.New("error"), 1, nil)

	if err = validatorRules(validatorS, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that if we don't ignore types, we will get error
	// even if the values in type are equal
	f := idStruct{ID: 1}
	valObjectBytes, err := json.Marshal(&f)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	//cacheSetup(valObjectBytes, nil)
	reset(valObjectBytes, nil, -1, nil)

	validatorS.cacheValidate.KeyIdx = -1
	validatorS.placeHolderIdx = -1
	validatorS.cacheValidate.PropertyName = "id"

	if err = validatorRules(validatorS, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("should have '%s' error\n", InvalidTxt)
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that if we ignore types, we will not get error
	reset(valObjectBytes, nil, -1, nil)
	validatorS.cacheValidate.PropertyName = "id"
	validatorS.cacheValidate.IgnoreTypes = true
	// validatorS.validateConf.IgnoreTypes = true
	validatorS.placeHolderIdx = 0

	if err = validatorRules(validatorS, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that when we pass slice values for validation and
	// receive array from cache and they equal each other, we
	// don't get error with IgnoreTypes set
	validateSliceVal := []int{1}
	fSlice := []idStruct{
		{
			ID: 1,
		},
	}

	valObjectSliceBytes, err := json.Marshal(&fSlice)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	reset(valObjectSliceBytes, nil, -1, nil)
	validatorS.cacheValidate.PropertyName = "id"

	if err = validatorRules(validatorS, validateSliceVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("err should be %s; got %s", InvalidTxt, err.Error())
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that when we pass slice values for validation and
	// receive array from cache and they equal each other, we
	// get error because they are not same type

	reset(valObjectSliceBytes, nil, -1, nil)
	validatorS.cacheValidate.PropertyName = "id"
	validatorS.cacheValidate.IgnoreTypes = true

	if err = validatorRules(validatorS, validateSliceVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that if we have wrong property type and
	// set IgnoreInvalidCacheResults to false, we get error
	reset(valObjectBytes, nil, -1, nil)
	validatorS.cacheValidate.PropertyName = "invalid"
	validatorS.cacheValidate.IgnoreInvalidCacheResults = false

	if err = validatorRules(validatorS, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("should have err: %s\n", err.Error())
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name but IgnoreInvalidCacheResults is set for
	// validateArgsType that we resort to db and not get error
	reset(valObjectSliceBytes, nil, 1, nil)
	validatorS.cacheValidate.PropertyName = "invalid"
	validatorS.cacheValidate.IgnoreInvalidCacheResults = true

	if err = validatorRules(validatorS, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name but IgnoreInvalidCacheResults is set false
	// so we get error
	reset(valObjectSliceBytes, nil, -1, nil)

	validatorS.err = errors.New(DoesNotExistTxt)
	validatorS.cacheValidate.PropertyName = "invalid"
	validatorS.cacheValidate.IgnoreInvalidCacheResults = false

	if err = validatorRules(validatorS, validateVal, validateExistsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != DoesNotExistTxt {
			t.Errorf("should have '%s' error\n", DoesNotExistTxt)
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name but IgnoreInvalidCacheResults is set
	// for validateExistsType that we resort to db and not get error
	reset(valObjectSliceBytes, nil, 1, nil)

	//validatorS.cache = mockCache
	//*validatorS.cacheValidate = *cacheValidate
	validatorS.err = errors.New(DoesNotExistTxt)
	validatorS.cacheValidate.PropertyName = "invalid"
	validatorS.cacheValidate.IgnoreInvalidCacheResults = true

	if err = validatorRules(validatorS, validateVal, validateExistsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name and IgnoreInvalidCacheResults is set false but since
	// this is validateUniquenessType we don't get error since
	// validateUniquenessType is valid when we have zero results
	reset(valObjectSliceBytes, nil, -1, nil)

	validatorS.err = errors.New(AlreadyExistsTxt)
	validatorS.cacheValidate.PropertyName = "invalid"
	validatorS.cacheValidate.IgnoreInvalidCacheResults = false

	if err = validatorRules(validatorS, validateVal, validateUniquenessType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name but IgnoreInvalidCacheResults is set
	// for validateUniquenessType that we resort to db and not get error
	reset(valObjectSliceBytes, nil, 0, nil)

	validatorS.err = errors.New(AlreadyExistsTxt)
	validatorS.cacheValidate.IgnoreTypes = true
	validatorS.cacheValidate.PropertyName = "id"
	validatorS.cacheValidate.IgnoreInvalidCacheResults = true

	if err = validatorRules(validatorS, validateVal, validateUniquenessType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing when we get array cache results and have invalid
	// property name but IgnoreInvalidCacheResults is set
	// for validateUniquenessType that we resort to db and not get error
	reset(nil, ErrCacheNil, -1, nil)

	validatorS.err = errors.New(AlreadyExistsTxt)
	validatorS.cacheValidate.IgnoreCacheNil = false

	if err = validatorRules(validatorS, validateVal, validateUniquenessType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that if we get ErrCacheNil error and IgnoreCacheNil is false,
	// we get error
	reset(nil, ErrCacheNil, -1, nil)

	validatorS.cacheValidate.IgnoreCacheNil = false

	if err = validatorRules(validatorS, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("error should be '%s'", InvalidTxt)
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that array cache results that is not objects
	// with IgnoreTypes set will not return error
	singleSliceVals := []int{1}
	singleSliceBytes, err := json.Marshal(singleSliceVals)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	reset(singleSliceBytes, nil, -1, nil)

	//validatorS.validateConf.IgnoreTypes = true
	validatorS.cacheValidate.IgnoreTypes = true

	if err = validatorRules(validatorS, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that array cache results that is not objects
	// with IgnoreTypes set false will return error
	reset(singleSliceBytes, nil, -1, nil)

	// validatorS.validateConf.IgnoreTypes = false
	validatorS.cacheValidate.IgnoreTypes = false

	if err = validatorRules(validatorS, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("should have '%s' error\n", err.Error())
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that bool is not a proper type and results in error
	boolSlice := []bool{true}
	boolSliceBytes, err := json.Marshal(boolSlice)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	reset(boolSliceBytes, nil, -1, nil)

	if err = validatorRules(validatorS, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != errInvalidCacheTypeInternal.Error() {
			t.Errorf("err type should be 'errInvalidCacheTypeInternal'")
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that cache result that is not array or object is
	// not valid with IgnoreTypes = false so return error
	singleValBytes, err := json.Marshal(1)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	reset(singleValBytes, nil, -1, nil)

	//validatorS.validateConf.IgnoreTypes = false
	validatorS.cacheValidate.IgnoreTypes = false

	if err = validatorRules(validatorS, validateVal, validateArgsType); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != InvalidTxt {
			t.Errorf("err should be %s; got %s", InvalidTxt, err.Error())
		}
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that cache result that is not array or object with
	// IgnoreTypes set is valid and does not return error
	reset(singleValBytes, nil, -1, nil)

	// validatorS.validateConf.IgnoreTypes = true
	validatorS.cacheValidate.IgnoreTypes = true

	if err = validatorRules(validatorS, validateVal, validateArgsType); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = assertExpectations(); err != nil {
		t.Error(err.Error())
	}

	// -----------------------------------------------------------

	// Testing that validateUniquenessRule will return nil error
	// if the instance and value are equal and that is should not
	// even make it to cache or database
	reset(nil, nil, -1, queryErr)

	iv := "foo"
	validateValueStr := "t"
	incompatableValidateVal := 1

	s := idStruct{}

	v2 := &validator{
		querier:        newDB,
		bindVar:        sqlx.DOLLAR,
		placeHolderIdx: -1,
		query:          "select id from idStruct where id = ?;",
		entityRecover:  formValidator,
		err:            errors.New(InvalidTxt),
	}

	vur := &validateUniquenessRule{
		validator:     v2,
		instanceValue: s,
	}

	if err = vur.Validate(nil); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	reset(nil, nil, -1, queryErr)

	if !errors.Is(vur.Validate(s), errInvalidValidateValue) {
		t.Errorf("error should be %s\n", errInvalidValidateValue.Error())
	}

	reset(nil, nil, -1, queryErr)

	err = vur.Validate(validateValueStr)

	if !errors.Is(err, errInvalidInstanceValue) {
		t.Errorf("error should be %s; got: %s\n", errInvalidInstanceValue.Error(), err.Error())
	}

	vur.instanceValue = &s

	reset(nil, nil, -1, queryErr)

	err = vur.Validate(validateValueStr)

	if !errors.Is(err, errInvalidInstanceValue) {
		t.Errorf("error should be %s got: %s\n", errInvalidInstanceValue.Error(), err.Error())
	}

	reset(nil, nil, -1, queryErr)

	vur.instanceValue = iv
	err = vur.Validate(incompatableValidateVal)

	if !errors.Is(err, errIncompatableTypes) {
		t.Errorf("error should be %s got: %s\n", errIncompatableTypes.Error(), err.Error())
	}

	reset(nil, nil, -1, queryErr)

	vur.instanceValue = iv

	if err = vur.Validate(&iv); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

}

func TestFoeBar(t *testing.T) {
	var err error
	iv := "foo"

	v2 := &validator{
		//querier:        newDB,
		bindVar:        sqlx.DOLLAR,
		placeHolderIdx: -1,
		query:          "select id from idStruct where id = ?;",
		//entityRecover:  formValidator,
		err: errors.New(InvalidTxt),
	}

	vur := &validateUniquenessRule{
		validator:     v2,
		instanceValue: iv,
	}

	if err = vur.Validate(&iv); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	t.Errorf("boom")
}
