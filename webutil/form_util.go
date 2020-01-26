package webutil

//go:generate mockgen -source=form_util.go -destination=../webutilmock/form_util_mock.go -package=webutilmock
//go:generate mockgen -source=form_util.go -destination=form_util_mock_test.go -package=webutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	validation "github.com/go-ozzo/ozzo-validation"
)

//////////////////////////////////////////////////////////////////
//--------------------------- CONSTS -------------------------
//////////////////////////////////////////////////////////////////

const (
	// RequiredTxt is string const error when field is required
	RequiredTxt = "Required"

	// AlreadyExistsTxt is string const error when field already exists
	// in database or cache
	AlreadyExistsTxt = "Already exists"

	// DoesNotExistTxt is string const error when field does not exist
	// in database or cache
	DoesNotExistTxt = "Does not exist"

	// InvalidTxt is string const error when field is invalid
	InvalidTxt = "Invalid"

	// InvalidFormatTxt is string const error when field has invalid format
	InvalidFormatTxt = "Invalid format"

	// InvalidFutureDateTxt is sring const when field is not allowed
	// to be in the future
	InvalidFutureDateTxt = "Date can't be after current date/time"

	// InvalidPastDateTxt is sring const when field is not allowed
	// to be in the past
	InvalidPastDateTxt = "Date can't be before current date/time"

	// CantBeNegativeTxt is sring const when field can't be negative
	CantBeNegativeTxt = "Can't be negative"
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

//////////////////////////////////////////////////////////////////
//----------------------- PRE-BUILT RULES ----------------------
//////////////////////////////////////////////////////////////////

var (
	// RequiredRule makes field required and does NOT allow just spaces
	RequiredRule = &validateRequiredRule{err: errors.New(RequiredTxt)}
)

//////////////////////////////////////////////////////////////////
//----------------------- CUSTOM ERRORS -------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrFutureAndPastDateInternal returns for form field if
	// user sets that a date can not be both a future or past date
	ErrFutureAndPastDateInternal = errors.New("webutil: both 'canBeFuture and 'canBePast' can not be false")

	// ErrInvalidStringInternal returns for form field if
	// data type for date field is not "string" or "*string"
	ErrInvalidStringInternal = errors.New("webutil: input must be string or *string")

	// ErrInvalidFormSelectionInternal will be returned if a query does not
	// return two columns when trying to query for Formselections
	ErrInvalidFormSelectionInternal = errors.New("webutil: query should only return 2 columns")
)

//////////////////////////////////////////////////////////////////
//-------------------------- TYPES ----------------------------
//////////////////////////////////////////////////////////////////

// PathRegex is a work around for the fact that injecting and retrieving a route into
// mux is quite complex without spinning up an entire server
type PathRegex func(r *http.Request) (string, error)

//////////////////////////////////////////////////////////////////
//---------------------- CONFIG STRUCTS ------------------------
//////////////////////////////////////////////////////////////////

// FormValidationConfig is config struct used in the initialization
// of *FormValidation
type FormValidationConfig struct {
	Cache     CacheStore
	RecoverDB RecoverDB
	PathRegex PathRegex
}

// CacheValidate is config struct used in form validators
type CacheValidate struct {
	// Key is key used in cache to retreive value
	// Key can be string format such as "email-%s"
	// in which case the KeyArgs property and/or current validated
	// value (if property KeyIdx > 0) will be used
	// to dynamically fill in to be used as key for cache
	Key string

	// KeyArgs is used against the Key property dynamically
	// to change the key for cache as the validator functions
	// will use fmt.Sprint(key, args...) to format the key
	KeyArgs []interface{}

	// KeyIdx is used if current value being validated
	// in a validator function should be used as apart of
	// the key
	//
	// If it is, it will be added to given index of the KeyArgs property
	// If KeyIdx is below 0, validated value won't be added to key
	KeyIdx int

	// IgnoreCacheNil will ignore cache nil error and still query db
	IgnoreCacheNil bool

	// ValueFieldName is the name of field that we are searching
	// for to compare our Value property with that returned from cache
	ValueFieldName string

	// IgnoreValueFieldName is flag to indicate whether we ignore
	// that the value in property ValueFieldName could not be found
	// in unmarshaled map from cache and continue to query db
	IgnoreValueFieldName bool
}

//////////////////////////////////////////////////////////////////
//----------------------- INTERFACES --------------------------
//////////////////////////////////////////////////////////////////

// RequestValidator should implement validating fields sent from
// request and return form or error if one occurs
type RequestValidator interface {
	Validate(req *http.Request, instance interface{}) (interface{}, error)
}

//////////////////////////////////////////////////////////////////
//------------------------- STRUCTS ---------------------------
//////////////////////////////////////////////////////////////////

// Boolean is used to be able interpret string bool values
type Boolean struct {
	value bool
}

// UnmarshalJSON takes in value and interprets it as a
// string and parses it to determine its value
// This is used so if user sends invalid bool value
// we don't panic when trying to process
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

// Value returns given bool value
func (c Boolean) Value() bool {
	return c.value
}

// Int64 is mainly used for slice of id values from forms
type Int64 int64

// MarshalJSON takes int64 value and returns byte value
func (i Int64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(i), IntBase))
}

// UnmarshalJSON first tries to convert given bytes
// to string because values are being passed from a
// web frontend client using javascript, javascript is not
// capable of int64 so it must send in string format
// and we convert it from string to int64
//
// Else fall back to trying to unmarshal an int64
func (i *Int64) UnmarshalJSON(b []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if s == "" {
			i = nil
			return nil
		}

		value, err := strconv.ParseInt(s, IntBase, IntBitSize)
		if err != nil {
			return err
		}
		*i = Int64(value)
		return nil
	}

	// Fallback to number
	return json.Unmarshal(b, (*int64)(i))
}

// Value returns given int64 value
func (i Int64) Value() int64 {
	return int64(i)
}

// FormSelection is generic struct used for html forms
type FormSelection struct {
	Text  interface{} `json:"text"`
	Value interface{} `json:"value"`
}

// FormValidation is the main struct that other structs will
// embed to validate json data
type FormValidation struct {
	entity Entity
	config FormValidationConfig
	//cache  CacheStore
}

// NewFormValidation returns *FormValidation instance
func NewFormValidation(entity Entity, config FormValidationConfig) *FormValidation {
	return &FormValidation{
		entity: entity,
		config: config,
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
//
// If a user wishes, they can also verify whether the given date is
// allowed to be a past or future date of the current time
//
// The timezone parameter converts given time to compare to current
// time if you choose to
// If no timezone is passed, UTC is used by default
// If user does not want to compare time, both bool parameters
// should be true
//
// Will raise ErrFutureAndPastDateInternal error which will be wrapped
// in validation.InternalError if both bool parameters are false
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

// ValidateIDs takes a list of arguments and queries against the querier type
// given and returns an validateIDsRule instance to indicate whether the ids
// are valid or not
//
// If the only placeholder parameter within your query is the ids validating against,
// then the args parameter of ValidateIDs can be nil
//
// Note of caution, the ids we are validating against should be the first placeholder
// parameters within the query passed
//
// If the ids passed happen to be type formutil#Int64, it will extract the values so it
// can be used against the query properly
func (f *FormValidation) ValidateIDs(
	cacheValidate *CacheValidate,
	recoverDB RecoverDB,
	placeHolderIdx,
	bindVar int,
	query string,
	args ...interface{},
) *validateIDsRule {
	return &validateIDsRule{
		validator: &validator{
			querier:        f.entity,
			cache:          f.config.Cache,
			entityRecover:  f,
			recoverDB:      recoverDB,
			cacheValidate:  cacheValidate,
			placeHolderIdx: placeHolderIdx,
			bindVar:        bindVar,
			query:          query,
			args:           args,
			err:            errors.New(InvalidTxt),
		},
	}
}

// ValidateUniqueness determines whether passed field is unique
// within database or cache if set
func (f *FormValidation) ValidateUniqueness(
	cacheValidate *CacheValidate,
	recoverDB RecoverDB,
	instanceValue interface{},
	placeHolderIdx,
	bindVar int,
	query string,
	args ...interface{},
) *validateUniquenessRule {
	return &validateUniquenessRule{
		instanceValue: instanceValue,
		validator: &validator{
			querier:        f.entity,
			cache:          f.config.Cache,
			entityRecover:  f,
			recoverDB:      recoverDB,
			cacheValidate:  cacheValidate,
			placeHolderIdx: placeHolderIdx,
			bindVar:        bindVar,
			query:          query,
			args:           args,
			err:            errors.New(AlreadyExistsTxt),
		},
	}
}

// ValidateExists determines whether passed field exists
// within database or cache if set
func (f *FormValidation) ValidateExists(
	cacheValidate *CacheValidate,
	recoverDB RecoverDB,
	placeHolderIdx,
	bindVar int,
	query string,
	args ...interface{},
) *validateExistsRule {
	return &validateExistsRule{
		validator: &validator{
			querier:        f.entity,
			cache:          f.config.Cache,
			entityRecover:  f,
			recoverDB:      recoverDB,
			cacheValidate:  cacheValidate,
			placeHolderIdx: placeHolderIdx,
			bindVar:        bindVar,
			query:          query,
			args:           args,
			err:            errors.New(DoesNotExistTxt),
		},
	}
}

// GetEntity returns Entity
func (f *FormValidation) GetEntity() Entity {
	return f.entity
}

// GetCache returns CacheStore
func (f *FormValidation) GetCache() CacheStore {
	return f.config.Cache
}

// GetConfig return FormValidationConfig
func (f *FormValidation) GetConfig() FormValidationConfig {
	return f.config
}

// SetEntity sets Entity
func (f *FormValidation) SetEntity(entity Entity) {
	f.entity = entity
}

// SetCache sets CacheStore
func (f *FormValidation) SetCache(cache CacheStore) {
	f.config.Cache = cache
}

// SetConfig sets FormValidationConfig
func (f *FormValidation) SetConfig(config FormValidationConfig) {
	f.config = config
}

type validator struct {
	entityRecover  EntityRecover
	querier        Querier
	cache          CacheStore
	args           []interface{}
	query          string
	bindVar        int
	placeHolderIdx int
	err            error
	recoverDB      RecoverDB
	cacheValidate  *CacheValidate
}

type validateRequiredRule struct {
	err error
}

func (v *validateRequiredRule) Validate(value interface{}) error {
	var err error

	checkValue := func(val string) error {
		val = strings.TrimSpace(val)

		if len(val) == 0 {
			return errors.New(RequiredTxt)
		}

		return nil
	}

	if isNilValue(value) {
		return errors.New(RequiredTxt)
	}

	switch value.(type) {
	case string:
		err = checkValue(value.(string))
	case *string:
		temp := value.(*string)
		err = checkValue(*temp)
	}

	return err
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
		return errors.New(InvalidFormatTxt)
	}

	if v.canBeFuture && v.canBePast {
		err = nil
	} else if v.canBeFuture {
		if dateTime.Before(currentTime) {
			err = errors.New(InvalidPastDateTxt)
		}
	} else if v.canBePast {
		if dateTime.After(currentTime) {
			err = errors.New(InvalidFutureDateTxt)
		}
	} else {
		err = validation.NewInternalError(ErrFutureAndPastDateInternal)
	}

	return err
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
	*validator
}

func (v *validateExistsRule) Validate(value interface{}) error {
	return checkIfExists(v.validator, value, true)
}

func (v *validateExistsRule) Error(message string) *validateExistsRule {
	v.err = errors.New(message)
	return v
}

type validateUniquenessRule struct {
	*validator
	instanceValue interface{}
}

func (v *validateUniquenessRule) Validate(value interface{}) error {
	// If value and instance value are the same return nil
	// as this indicates that the form value hasn't changed
	if v.instanceValue == value {
		return nil
	}

	return checkIfExists(v.validator, value, false)
}

func (v *validateUniquenessRule) Error(message string) *validateUniquenessRule {
	v.err = errors.New(message)
	return v
}

type validateIDsRule struct {
	*validator
}

func (v *validateIDsRule) Validate(value interface{}) error {
	var err error
	var searchVals []interface{}
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

		for k := 0; k < s.Len(); k++ {
			t := s.Index(k).Interface()
			searchVals = append(searchVals, applyTypeToInterface(t))
		}

		expectedLen = len(searchVals)
	default:
		expectedLen = 1
		singleVal = value
	}

	// If type is slice and is empty, simply return nil as we will get an error
	// when trying to query with empty slice
	if isSlice && len(searchVals) == 0 {
		return nil
	}

	var args []interface{}

	if v.placeHolderIdx > -1 {
		args = make([]interface{}, 0, len(args)+1)
		args = append(args, v.args...)

		if isSlice {
			args = InsertAt(args, searchVals, v.placeHolderIdx)
		} else {
			args = InsertAt(args, singleVal, v.placeHolderIdx)
		}
	} else {
		args = make([]interface{}, 0, len(args))
		args = append(args, v.args...)
	}

	q, arguments, err := InQueryRebind(v.bindVar, v.query, args...)

	if err != nil {
		messageErr := fmt.Errorf(err.Error()+"\n query: %s\n args:%v\n", q, args)
		return validation.NewInternalError(messageErr)
	}

	queryFunc := func() error {
		var rows *sqlx.Rows
		isValidQuery := false

		if rows, err = v.querier.Queryx(q, arguments...); err != nil {
			if v.recoverDB != nil {
				if db, err := v.recoverDB(err); err == nil {
					v.querier = db
					v.entityRecover.SetEntity(db)

					if rows, err = v.querier.Queryx(q, arguments...); err != nil {
						return validation.NewInternalError(err)
					}

					isValidQuery = true
				}
			}
		} else {
			isValidQuery = true
		}

		if !isValidQuery {
			return validation.NewInternalError(err)
		}

		counter := 0

		for rows.Next() {
			counter++
		}

		if expectedLen != counter {
			return v.err
		}

		return nil
	}

	if v.cache != nil && v.cacheValidate != nil {
		var singleID bool
		var cacheBytes []byte

		if !isSlice {
			singleID = true
		}

		var cacheArgs []interface{}

		if v.cacheValidate.KeyIdx > -1 {
			cacheArgs = make([]interface{}, 0, len(v.cacheValidate.KeyArgs)+1)
			cacheArgs = append(cacheArgs, v.cacheValidate.KeyArgs...)
			cacheArgs = InsertAt(v.cacheValidate.KeyArgs, value, v.cacheValidate.KeyIdx)
		} else {
			cacheArgs = make([]interface{}, 0, len(cacheArgs))
			cacheArgs = append(cacheArgs, v.cacheValidate.KeyArgs...)
		}

		key := fmt.Sprintf(v.cacheValidate.Key, cacheArgs...)
		cacheBytes, err = v.cache.Get(key)

		if err != nil {
			if err == ErrCacheNil {
				if v.cacheValidate.IgnoreCacheNil {
					err = queryFunc()
				}
			} else {
				err = queryFunc()
			}
		} else {
			if singleID {
				var cacheMap map[string]interface{}

				if err = json.Unmarshal(cacheBytes, &cacheMap); err != nil {
					return validation.NewInternalError(err)
				}

				if val, ok := cacheMap[v.cacheValidate.ValueFieldName]; ok {
					switch val.(type) {
					case float64:
						val = strconv.FormatFloat(val.(float64), 'e', -1, IntBitSize)
					}

					value = applyTypeToInterface(value)

					if val != value {
						return v.err
					}
				} else {
					if v.cacheValidate.IgnoreValueFieldName {
						err = queryFunc()
					} else {
						return v.err
					}
				}
			} else {
				var cacheSlice []map[string]interface{}
				var valSlice []interface{}

				count := 0

				if err = json.Unmarshal(cacheBytes, &cacheSlice); err != nil {
					return validation.NewInternalError(err)
				}

				validField := true

				for _, t := range cacheSlice {
					if val, ok := t[v.cacheValidate.ValueFieldName]; ok {
						switch val.(type) {
						case float64:
							str := strconv.FormatFloat(val.(float64), 'e', -1, IntBitSize)
							valSlice = append(valSlice, str)
						default:
							valSlice = append(valSlice, val)
						}
					} else {
						if v.cacheValidate.IgnoreValueFieldName {
							err = queryFunc()
							validField = false
							break
						} else {
							return v.err
						}
					}
				}

				if validField {
					for _, i := range searchVals {
						for _, t := range valSlice {
							if i == t {
								count++
							}
						}
					}

					if count != len(searchVals) {
						err = v.err
					}
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

// HasFormErrors determines what type of form error is passed and
// sends appropriate error message to client
//
// ServerErrorResponse and ClientErrorResponse have set default
// if user does not
//
// RecoverDB is optional and can be set but RetryDB is not used
func HasFormErrors(
	w http.ResponseWriter,
	err error,
	config ServerErrorConfig,
) bool {
	if err != nil {
		SetHTTPResponseDefaults(&config.ServerErrorResponse, 500, []byte(ErrServer.Error()))
		SetHTTPResponseDefaults(&config.ClientErrorResponse, 406, nil)

		switch err {
		case ErrBodyRequired:
			w.WriteHeader(*config.ClientErrorResponse.HTTPStatus)
			w.Write([]byte(ErrBodyRequired.Error()))
		case ErrInvalidJSON:
			w.WriteHeader(*config.ClientErrorResponse.HTTPStatus)
			w.Write([]byte(ErrInvalidJSON.Error()))
		default:
			if payload, ok := err.(validation.Errors); ok {
				w.WriteHeader(*config.ClientErrorResponse.HTTPStatus)
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
// The query given should only return two columns, value and
//
// ServerErrorResponse and ClientErrorResponse have set default
// if user does not
//
// RecoverDB is optional and can be set but RetryDB is not used
func GetFormSelections(
	w http.ResponseWriter,
	config ServerErrorCacheConfig,
	db Querier,
	bindVar int,
	query string,
	args ...interface{},
) ([]FormSelection, error) {
	var err error
	var newDB *sqlx.DB

	SetHTTPResponseDefaults(&config.ServerErrorResponse, 500, []byte(ErrServer.Error()))

	getFormSelectionsFromDB := func() ([]FormSelection, error) {
		if query, args, err = InQueryRebind(bindVar, query, args...); err != nil {
			http.Error(w, ErrServer.Error(), http.StatusInternalServerError)
			return nil, err
		}

		rower, err := db.Queryx(query, args...)

		if err != nil {
			canRecover := false

			if config.RecoverDB != nil {
				if newDB, err = config.RecoverDB(err); err == nil {
					config.DBInterfaceRecover.SetDBInterface(newDB)
					rower, err = db.Queryx(query, args...)

					if err == nil {
						canRecover = true
					}
				}
			}

			if !canRecover {
				w.WriteHeader(*config.ServerErrorResponse.HTTPStatus)
				w.Write(config.ServerErrorResponse.HTTPResponse)
				return nil, err
			}
		}

		cols, err := rower.Columns()

		if err != nil {
			return nil, err
		}

		if len(cols) != 2 {
			return nil, ErrInvalidFormSelectionInternal
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

	jsonBytes, err := config.CacheConfig.Cache.Get(config.Key)

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

	if err = json.Unmarshal(jsonBytes, &forms); err != nil {
		http.Error(w, ErrServer.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return forms, nil
}

// CheckBodyAndDecode takes request and decodes the json body from the request
// to the passed struct
//
// The excludeMethods parameter allows user to pass certain http methods
// that skip decoding the request body if nil else will throw ErrBodyRequired error
func CheckBodyAndDecode(req *http.Request, form interface{}, excludeMethods ...string) error {
	canSkip := false

	for _, v := range excludeMethods {
		if req.Method == v {
			canSkip = true
			break
		}
	}

	if req.Body != nil && req.Body != http.NoBody {
		dec := json.NewDecoder(req.Body)

		if err := dec.Decode(&form); err != nil {
			return ErrInvalidJSON
		}
	} else {
		if !canSkip {
			return ErrBodyRequired
		}
	}

	return nil
}

// checkIfExists determines whether the passed query returns any
// results and returns error depending on the wantExists parameter
func checkIfExists(v *validator, value interface{}, wantExists bool) error {
	var err error
	var q string

	if isNilValue(value) {
		return nil
	}

	dbCall := func() error {
		var rows *sqlx.Rows
		queryArgs := make([]interface{}, 0, len(v.args)+1)

		if len(v.args) != 0 {
			queryArgs = append(queryArgs, v.args...)
		}

		queryArgs = InsertAt(queryArgs, value, v.placeHolderIdx)
		if q, queryArgs, err = InQueryRebind(v.bindVar, v.query, queryArgs...); err != nil {
			return validation.NewInternalError(err)
		}

		count := 0

		if rows, err = v.querier.Queryx(q, queryArgs...); err != nil {
			if v.recoverDB == nil {
				return validation.NewInternalError(err)
			}
			if db, err := v.recoverDB(err); err != nil {
				v.entityRecover.SetEntity(db)

				if rows, err = v.querier.Queryx(q, queryArgs...); err != nil {
					return validation.NewInternalError(err)
				}
			}
		}

		for rows.Next() {
			count++
		}

		if count == 0 {
			if wantExists {
				return v.err
			}
		} else {
			if !wantExists {
				return v.err
			}
		}

		return nil
	}

	if v.cache != nil && v.cacheValidate != nil {
		cacheArgs := make([]interface{}, 0)

		if v.cacheValidate.KeyIdx > -1 {
			cacheArgs = InsertAt(v.cacheValidate.KeyArgs, value, v.cacheValidate.KeyIdx)
		}

		key := fmt.Sprintf(v.cacheValidate.Key, cacheArgs...)
		_, err = v.cache.Get(key)

		if err != nil {
			if err != ErrCacheNil {
				if err = dbCall(); err != nil {
					return err
				}
			} else {
				if v.cacheValidate.IgnoreCacheNil {
					if err = dbCall(); err != nil {
						return err
					}
				}
			}
		} else {
			if !wantExists {
				return v.err
			}
		}
	} else {
		if err = dbCall(); err != nil {
			return err
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

func applyTypeToInterface(typ interface{}) interface{} {
	var t interface{}

	switch typ.(type) {
	case int:
		t = strconv.Itoa(typ.(int))
	case *int:
		i := typ.(*int)
		t = strconv.Itoa(*i)
	case int64:
		t = strconv.FormatInt(typ.(int64), IntBase)
	case *int64:
		i := typ.(*int64)
		t = strconv.FormatInt(*i, IntBase)
	case float64:
		t = strconv.FormatFloat(typ.(float64), 'e', -1, IntBitSize)
	case *float64:
		i := typ.(*float64)
		t = strconv.FormatFloat(*i, 'e', -1, IntBitSize)
	}

	return t
}
