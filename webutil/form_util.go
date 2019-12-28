package webutil

//go:generate mockgen -source=form_util.go -destination=../webutilmock/form_util_mock.go -package=webutilmock
//go:generate mockgen -source=form_util.go -destination=form_util_mock_test.go -package=webutil

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	validation "github.com/go-ozzo/ozzo-validation"
)

//////////////////////////////////////////////////////////////////
//----------------------- GLOBAL VARS -------------------------
//////////////////////////////////////////////////////////////////

var (
	// EmailRegex is regex expression used for forms to validate email
	EmailRegex = regexp.MustCompile("^.+@[a-zA-Z0-9.]+$")

	// ZipRegex is regex expression used for forms to validate zip code
	ZipRegex = regexp.MustCompile("^[0-9]{5}$")

	// PhoneNumberRegex is regex expression used for forms to validate phone number
	PhoneNumberRegex = regexp.MustCompile("^\\([0-9]{3}\\)-[0-9]{3}-[0-9]{4}$")

	// ColorRegex is regex expression used for forms to validate color format is correct
	ColorRegex = regexp.MustCompile("^#[0-9a-z]{6}$")

	// RequiredStringRegex is regex expression used for forms to validate that a field
	// has at least one character that is NOT a space
	RequiredStringRegex = regexp.MustCompile(`[^\s\\]`)

	// FormDateRegex is regex expression used for forms to validate correct format of date
	FormDateRegex = regexp.MustCompile("^[0-9]{1,2}/[0-9]{1,2}/[0-9]{4}$")

	// FormDateTimeRegex is regex expression used for forms to validate correct format of
	// date and time
	FormDateTimeRegex = regexp.MustCompile("^[0-9]{1,2}/[0-9]{1,2}/[0-9]{4} [0-9]{1,2}:[0-9]{2} (?i)(AM|PM)$")
)

var (
	// RequiredRule makes field required and does NOT allow just spaces
	RequiredRule = &validateRequiredRule{err: ErrRequiredValidator}
)

//////////////////////////////////////////////////////////////////
//----------------------- CUSTOM ERRORS -------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrRequiredValidator returns for form field if field is empty
	ErrRequiredValidator = errors.New("required")

	// ErrAlreadyExistsValidator returns for form field if an entry
	// in database or cache already exists
	ErrAlreadyExistsValidator = errors.New("already exists")

	// ErrDoesNotExistValidator returns for form field if an entry
	// in database or cache does not exists
	ErrDoesNotExistValidator = errors.New("does not exist")

	// ErrInvalidValidator returns for form field if value for
	// field is generally invalid
	ErrInvalidValidator = errors.New("invalid")

	// ErrInvalidFormatValidator returns for form field if value for
	// field is formatted incorrectly
	ErrInvalidFormatValidator = errors.New("invalid format")

	// ErrInvalidFutureDateValidator returns for form field if value for
	// field is not allowed to be passed the current date/time
	ErrInvalidFutureDateValidator = errors.New("date can not be after current date/time")

	// ErrInvalidPastDateValidator returns for form field if value for
	// field is not allowed to be before the current date/time
	ErrInvalidPastDateValidator = errors.New("date can not be before current date/time")

	// ErrCanNotBeNegativeValidator returns for form field if value
	// is negative, generally used for things like currency
	ErrCanNotBeNegativeValidator = errors.New("can not be negative")

	// ErrFutureAndPastDateInternal returns for form field if
	// user sets that a date can not be both a future or past date
	ErrFutureAndPastDateInternal = errors.New("both 'canBeFuture and 'canBePast' can not be false")

	// ErrInvalidStringInternal returns for form field if
	// data type for date field is not "string" or "*string"
	ErrInvalidStringInternal = errors.New("input must be string or *string")
)

//////////////////////////////////////////////////////////////////
//-------------------------- TYPES ----------------------------
//////////////////////////////////////////////////////////////////

// PathRegex is a work around for the fact that injecting and retrieving a route into
// mux is quite complex without spinning up an entire server
type PathRegex func(r *http.Request) (string, error)

//////////////////////////////////////////////////////////////////
//----------------------- INTERFACES --------------------------
//////////////////////////////////////////////////////////////////

// RequestValidator should implement validating fields sent from
// request and return form or error is one occurs validating
type RequestValidator interface {
	Validate(req *http.Request, instance interface{}) (interface{}, error)
}

//////////////////////////////////////////////////////////////////
//----------------------- CONFIG STRUCTS -----------------------
//////////////////////////////////////////////////////////////////

// FormSelectionConfig is config struct used for GetFormSelections function
type FormSelectionConfig struct {
	CacheConfig         CacheConfig
	RecoverDB           RecoverDB
	ServerErrorResponse HTTPResponseConfig
}

// FormErrorConfig is config struct used for HasFormErrors function
type FormErrorConfig struct {
	InvalidHTTPStatus   *int
	ServerErrorResponse HTTPResponseConfig
}

//////////////////////////////////////////////////////////////////
//------------------------- STRUCTS ---------------------------
//////////////////////////////////////////////////////////////////

type Boolean struct {
	value bool
}

func (c *Boolean) UnmarshalJSON(data []byte) error {
	asString := string(data)
	convertedBool, err := strconv.ParseBool(asString)

	if err != nil {
		c.value = false
	} else {
		c.value = convertedBool
	}

	return nil
}

func (c Boolean) Value() bool {
	return c.value
}

type Int64 int64

func (i Int64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(i), 10))
}

func (i *Int64) UnmarshalJSON(b []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if s == "" {
			i = nil
			return nil
		}

		value, err := strconv.ParseInt(s, 10, IntBitSize)
		if err != nil {
			return err
		}
		*i = Int64(value)
		return nil
	}

	// Fallback to number
	return json.Unmarshal(b, (*int64)(i))
}

func (i Int64) Value() int64 {
	return int64(i)
}

// FormSelection is generic struct used for html forms
type FormSelection struct {
	Text  interface{} `json:"text"`
	Value interface{} `json:"value"`
}

// FormValidation is the main struct that other structs will
// embed to validate json data.  It is also the struct that
// implements SetQuerier and SetCache of Form interface
type FormValidation struct {
	entity Entity
	cache  CacheStore
}

func NewFormValidation(entity Entity) *FormValidation {
	return &FormValidation{
		entity: entity,
		//cache:  cache,
	}
}

// IsValid returns *validRule based on isValid parameter
// Basically IsValid is a wrapper for the passed bool
// to return valid rule to then apply custom error message
// for the Error function
func (f *FormValidation) IsValid(isValid bool) *validRule {
	return &validRule{isValid: isValid, err: errors.New("Not Valid")}
}

// ValidateDate verifies whether a date string matches the passed in
// layout format
// If a user wishes, they can also verify whether the given date is
// allowed to be a past or future date of the current time
// The timezone parameter converts given time to compare to current
// time if you choose to
// If no timezone is passed, UTC is used by default
// If user does not want to compare time, both bool parameters
// should be true
// Will raise validation.InternalError if both bool parameters are false
func (f *FormValidation) ValidateDate(
	layout,
	timezone string,
	canBeFuture,
	canBePast bool,
) *validateDateRule {
	return &validateDateRule{
		layout:      layout,
		timezone:    timezone,
		canBeFuture: canBeFuture,
		canBePast:   canBePast,
	}
}

// ValidateIDs takes a list of arguments and queries against the querier type given and returns an validateIDsRule instance
// to indicate whether the ids are valid or not
// If the only placeholder parameter within your query is the ids validating against, then the args paramter of ValidateIDs
// can be nil
// Note of caution, the ids we are validating against should be the first placeholder parameters within the query passed
//
// If the ids passed happen to be type formutil#Int64, it will extract the values so it can be used against the query properly
//
// The cacheConfig parameter can be nil if you do not need/have a cache backend
func (f *FormValidation) ValidateIDs(
	cacheConfig *CacheConfig,
	placeHolderPosition,
	bindVar int,
	query string,
	args ...interface{},
) *validateIDsRule {
	return &validateIDsRule{
		validator: validator{
			querier:             f.entity,
			cacheConfig:         cacheConfig,
			placeHolderPosition: placeHolderPosition,
			bindVar:             bindVar,
			query:               query,
			args:                args,
			err:                 ErrInvalidValidator,
		},
	}
}

// ValidateUniqueness determines whether passed field is unique
// within database or cache if set
func (f *FormValidation) ValidateUniqueness(
	cacheConfig *CacheConfig,
	instanceValue interface{},
	placeHolderPosition,
	bindVar int,
	query string,
	args ...interface{},
) *validateUniquenessRule {
	return &validateUniquenessRule{
		instanceValue: instanceValue,
		validator: validator{
			querier:             f.entity,
			cacheConfig:         cacheConfig,
			placeHolderPosition: placeHolderPosition,
			bindVar:             bindVar,
			query:               query,
			args:                args,
			err:                 ErrAlreadyExistsValidator,
		},
	}
}

// ValidateExists determines whether passed field exists
// within database or cache if set
func (f *FormValidation) ValidateExists(
	querier Querier,
	cacheConfig *CacheConfig,
	placeHolderPosition,
	bindVar int,
	query string,
	args ...interface{},
) *validateExistsRule {
	return &validateExistsRule{
		validator: validator{
			querier:             querier,
			cacheConfig:         cacheConfig,
			placeHolderPosition: placeHolderPosition,
			bindVar:             bindVar,
			query:               query,
			args:                args,
			err:                 ErrDoesNotExistValidator,
		},
	}
}

// GetEntity returns Entity
func (f *FormValidation) GetEntity() Entity {
	return f.entity
}

// GetCache returns CacheStore
func (f *FormValidation) GetCache() CacheStore {
	return f.cache
}

// SetEntity sets Entity
func (f *FormValidation) SetEntity(entity Entity) {
	f.entity = entity
}

// SetCache sets CacheStore
func (f *FormValidation) SetCache(cache CacheStore) {
	f.cache = cache
}

type validator struct {
	querier             Querier
	cacheConfig         *CacheConfig
	args                []interface{}
	query               string
	bindVar             int
	placeHolderPosition int
	err                 error
}

type validateRequiredRule struct {
	err error
}

func (v *validateRequiredRule) Validate(value interface{}) error {
	var val string
	var err error

	if isNilValue(value) {
		return ErrRequiredValidator
	}

	if val, err = getStringFromValue(value); err != nil {
		return err
	}

	val = strings.TrimSpace(val)

	if len(val) == 0 {
		return validation.NewInternalError(ErrRequiredValidator)
	}

	return nil
}

func (v *validateRequiredRule) Error(message string) *validateRequiredRule {
	v.err = errors.New(message)
	return v
}

type validateDateRule struct {
	layout      string
	timezone    string
	canBeFuture bool
	canBePast   bool
	err         error
}

func (v *validateDateRule) Validate(value interface{}) error {
	var currentTime, dateTime time.Time
	var err error
	var dateValue string

	if isNilValue(value) {
		return nil
	}

	if dateValue, err = getStringFromValue(value); err != nil {
		return err
	}

	if v.timezone != "" {
		currentTime, err = GetCurrentLocalDateInUTC(v.timezone)

		if err != nil {
			return validation.NewInternalError(err)
		}
	} else {
		currentTime = time.Now().UTC()
	}

	if dateTime, err = time.Parse(v.layout, dateValue); err != nil {
		return ErrInvalidFormatValidator
	}

	if v.canBeFuture && v.canBePast {
		err = nil
	} else if v.canBeFuture {
		if dateTime.Before(currentTime) {
			err = ErrInvalidPastDateValidator
		}
	} else if v.canBePast {
		if dateTime.After(currentTime) {
			err = ErrInvalidFutureDateValidator
		}
	} else {
		err = validation.NewInternalError(ErrFutureAndPastDateInternal)
	}

	if err != nil {
		return err
	}

	return nil
}

func (v *validateDateRule) Error(message string) *validateDateRule {
	v.err = errors.New(message)
	return v
}

type validRule struct {
	isValid bool
	err     error
}

func (v *validRule) Validate(value interface{}) error {
	if !v.isValid {
		return v.err
	}

	return nil
}

// Error sets the error message for the rule.
func (v *validRule) Error(message string) *validRule {
	v.err = errors.New(message)
	return v
}

type validateExistsRule struct {
	validator
}

func (v *validateExistsRule) Validate(value interface{}) error {
	return checkCacheIfExists(v.validator, value, true)
}

func (v *validateExistsRule) Error(message string) *validateExistsRule {
	v.err = errors.New(message)
	return v
}

type validateUniquenessRule struct {
	validator
	instanceValue interface{}
}

func (v *validateUniquenessRule) Validate(value interface{}) error {
	// If value and instance value are the same return nil
	// as this indicates that the form value hasn't changed
	if v.instanceValue == value {
		return nil
	}

	return checkCacheIfExists(v.validator, value, false)
}

func (v *validateUniquenessRule) Error(message string) *validateUniquenessRule {
	v.err = errors.New(message)
	return v
}

type validateIDsRule struct {
	validator
}

func (v *validateIDsRule) Validate(value interface{}) error {
	var err error
	var ids []interface{}
	var expectedLen int
	var singleVal interface{}

	isSlice := false

	if isNilValue(value) {
		return nil
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		isSlice = true
		s := reflect.ValueOf(value)

		for i := 0; i < s.Len(); i++ {
			ids = append(ids, s.Index(i))
		}

		expectedLen = len(ids)
	default:
		expectedLen = 1
		singleVal = value
	}

	// If type is slice and is empty, simply return nil as we will get an error
	// when trying to query with empty slice
	if len(ids) == 0 {
		return nil
	}

	args := make([]interface{}, 0, len(v.args))

	if len(v.args) != 0 {
		args = append(args, v.args...)
	}

	if v.placeHolderPosition > 0 {
		if isSlice {
			args = InsertAt(args, ids, v.placeHolderPosition-1)
		} else {
			args = InsertAt(args, singleVal, v.placeHolderPosition-1)
		}
	}

	q, arguments, err := InQueryRebind(v.bindVar, v.query, args...)

	if err != nil {
		messageErr := fmt.Errorf(err.Error()+"\n query: %s\n args:%v\n", q, args)
		return validation.NewInternalError(messageErr)
	}

	queryFunc := func() error {
		rower, err := v.querier.Query(q, arguments...)

		if err != nil {
			errS := fmt.Errorf("query: %s  err: %s", q, err.Error())
			return validation.NewInternalError(errS)
		}

		counter := 0
		for rower.Next() {
			counter++
		}

		if expectedLen != counter {
			return v.err
		}

		return nil
	}

	if v.cacheConfig != nil {
		var validID bool
		var singleID bool
		var cacheBytes []byte

		if !isSlice {
			singleID = true
			validID, err = v.cacheConfig.Cache.HasKey(v.cacheConfig.Key)
		} else {
			cacheBytes, err = v.cacheConfig.Cache.Get(v.cacheConfig.Key)
		}

		if err != nil {
			if err == ErrCacheNil {
				if v.cacheConfig.IgnoreCacheNil {
					err = queryFunc()
				}
			} else {
				err = queryFunc()
			}
		} else {
			if singleID {
				if !validID {
					err = v.err
				}
			} else {
				var cacheIDs []interface{}
				err = json.Unmarshal(cacheBytes, &cacheIDs)

				if err != nil {
					return validation.NewInternalError(err)
				}

				count := 0

				for _, v := range ids {
					for _, t := range cacheIDs {
						if v == t {
							count++
						}
					}
				}

				if count != len(ids) {
					err = v.err
				}
			}
		}
	} else {
		err = queryFunc()
	}

	return err
}

func (v *validateIDsRule) Error(message string) *validateIDsRule {
	v.err = errors.New(message)
	return v
}

//////////////////////////////////////////////////////////////////
//----------------------- FUNCTIONS -------------------------
//////////////////////////////////////////////////////////////////

func StandardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// HasFormErrors determines what type of form error is passed and
// sends appropriate error message to client
func HasFormErrors(
	w http.ResponseWriter,
	err error,
	db DBInterface,
	config FormErrorConfig,
) bool {
	if err != nil {
		invalidStatus := 406
		serverError := 500

		if config.InvalidHTTPStatus == nil {
			config.InvalidHTTPStatus = &invalidStatus
		}
		if config.ServerErrorResponse.HTTPStatus == nil {
			config.ServerErrorResponse.HTTPStatus = &serverError
		}
		if config.ServerErrorResponse.HTTPResponse == nil {
			config.ServerErrorResponse.HTTPResponse = []byte(ErrServer.Error())
		}

		switch err {
		case ErrBodyRequired:
			w.WriteHeader(*config.InvalidHTTPStatus)
			w.Write([]byte(ErrBodyRequired.Error()))
		case ErrInvalidJSON:
			w.WriteHeader(*config.InvalidHTTPStatus)
			w.Write([]byte(ErrInvalidJSON.Error()))
		default:
			if payload, ok := err.(validation.Errors); ok {
				w.WriteHeader(*config.InvalidHTTPStatus)
				jsonString, _ := json.Marshal(payload)
				w.Write(jsonString)
			} else {
				w.WriteHeader(*config.ServerErrorResponse.HTTPStatus)
				w.Write(config.ServerErrorResponse.HTTPResponse)
			}
		}

		return true
	}

	return false
}

// GetFormSelections takes query with arguments and returns slice of
// FormSelection of result
func GetFormSelections(
	w http.ResponseWriter,
	config FormSelectionConfig,
	db Querier,
	bindVar int,
	query string,
	args ...interface{},
) ([]FormSelection, error) {
	var err error

	defaultStatus := 500

	if config.ServerErrorResponse.HTTPStatus == nil {
		config.ServerErrorResponse.HTTPStatus = &defaultStatus
	}
	if config.ServerErrorResponse.HTTPResponse == nil {
		config.ServerErrorResponse.HTTPResponse = []byte(ErrServer.Error())
	}

	getFormSelectionsFromDB := func() ([]FormSelection, error) {
		query, args, err = InQueryRebind(bindVar, query, args...)

		if HasServerError(w, err, "") {
			return nil, err
		}

		rower, err := db.Query(query, args...)

		if err = config.RecoverDB(err); err != nil {
			w.WriteHeader(*config.ServerErrorResponse.HTTPStatus)
			w.Write(config.ServerErrorResponse.HTTPResponse)
			return nil, err
		}

		forms := make([]FormSelection, 0)

		for rower.Next() {
			form := FormSelection{}
			err = rower.Scan(
				&form.Value,
				&form.Text,
			)

			forms = append(forms, form)
		}

		return forms, nil
	}

	if config.CacheConfig.Cache == nil {
		return getFormSelectionsFromDB()
	}

	jsonBytes, err := config.CacheConfig.Cache.Get(config.CacheConfig.Key)

	if err != nil {
		if err != ErrCacheNil {
			return getFormSelectionsFromDB()
		}
		if config.CacheConfig.IgnoreCacheNil {
			return getFormSelectionsFromDB()
		}

		w.WriteHeader(*config.ServerErrorResponse.HTTPStatus)
		w.Write(config.ServerErrorResponse.HTTPResponse)
		return nil, err
	}

	forms := make([]FormSelection, 0)
	err = json.Unmarshal(jsonBytes, &forms)

	if HasServerError(w, err, "") {
		return nil, err
	}

	return forms, nil
}

// CheckBodyAndDecode takes request and decodes the json body from the request
// to the passed struct
//
// The excludeMethods parameter allows user to pass certain http methods
// that skip decoding the request body if nil
func CheckBodyAndDecode(req *http.Request, form interface{}, excludeMethods ...string) error {
	canSkip := false

	for _, v := range excludeMethods {
		if req.Method == v {
			canSkip = true
			break
		}
	}

	if req.Body != nil {
		dec := json.NewDecoder(req.Body)
		err := dec.Decode(&form)

		if err != nil {
			fmt.Printf(err.Error())
			return ErrInvalidJSON
		}
	} else {
		if !canSkip {
			return ErrBodyRequired
		}
	}

	return nil
}

// checkCacheIfExists determines whether the passed query returns any
// results and returns error depending on the wantExists parameter
func checkCacheIfExists(v validator, value interface{}, wantExists bool) error {
	var filler interface{}
	var err error
	var q string

	if isNilValue(value) {
		return nil
	}

	dbCall := func() error {
		queryArgs := make([]interface{}, 0)

		if len(v.args) != 0 {
			queryArgs = append(queryArgs, v.args...)
		}

		v.args = InsertAt(v.args, value, v.placeHolderPosition)
		q, queryArgs, err = InQueryRebind(v.bindVar, v.query, queryArgs...)

		if err != nil {
			return validation.NewInternalError(err)
		}

		row := v.querier.QueryRow(q, queryArgs...)

		if err = row.Scan(&filler); err != nil {
			if wantExists {
				if err == sql.ErrNoRows {
					return v.err
				}

				return validation.NewInternalError(err)
			}

			if err != sql.ErrNoRows {
				return validation.NewInternalError(err)
			}
		}

		return nil
	}

	if v.cacheConfig != nil {
		exists, err := v.cacheConfig.Cache.HasKey(v.cacheConfig.Key)

		if err != nil {
			if err != ErrCacheNil {
				if err = dbCall(); err != nil {
					return err
				}
			} else {
				if v.cacheConfig.IgnoreCacheNil {
					if err = dbCall(); err != nil {
						return err
					}
				}
			}
		} else {
			if wantExists {
				if !exists {
					return v.err
				}
			} else {
				if exists {
					return v.err
				}
			}
		}
	}

	return nil
}

// isNilValue determines if passed value is truley nil
func isNilValue(value interface{}) bool {
	_, isNil := validation.Indirect(value)
	if validation.IsEmpty(value) || isNil {
		return true
	}

	return false
}

// getStringFromValue determines whether passed value is a
// string or *string type and if not, returns ErrInvalidStringInternal
func getStringFromValue(value interface{}) (string, error) {
	var dateValue string
	var err error

	switch value.(type) {
	case string:
		dateValue = value.(string)
	case *string:
		temp := value.(*string)

		if temp == nil {
			return "", nil
		}

		dateValue = *temp
	default:
		err = validation.NewInternalError(ErrInvalidStringInternal)
	}

	return dateValue, err
}
