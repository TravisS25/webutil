package webutil

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/TravisS25/webutil/webutilcfg"
	"github.com/shopspring/decimal"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	validation "github.com/go-ozzo/ozzo-validation"
)

//////////////////////////////////////////////////////////////////
//---------------------- VALIDATOR TYPES -----------------------
//////////////////////////////////////////////////////////////////

const (
	validateArgsType = iota + 1
	validateUniquenessType
	validateExistsType
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
	ColorRegex = regexp.MustCompile("^#[0-9a-f]{6}$")

	// RequiredStringRegex is regex expression used for forms to validate that a field
	// has at least one character that is NOT a space
	RequiredStringRegex = regexp.MustCompile(`[^\s\\]`)

	// FormDateRegex is regex expression used for forms to validate correct format of date
	FormDateRegex = regexp.MustCompile("^[0-9]{1,2}/[0-9]{1,2}/[0-9]{4}$")

	// FormDateTimeRegex is regex expression used for forms to validate correct format of
	// date and time
	FormDateTimeRegex = regexp.MustCompile("^[0-9]{1,2}/[0-9]{1,2}/[0-9]{4} [0-9]{1,2}:[0-9]{2} (?i)(AM|PM)$")

	// USDCurrencyRegex represents usd currency format to validate for form
	USDCurrencyRegex = regexp.MustCompile("^(?:-)?[0-9]+(?:\\.[0-9]{1,2})?$")
)

//////////////////////////////////////////////////////////////////
//----------------------- PRE-BUILT RULES ----------------------
//////////////////////////////////////////////////////////////////

var (
	// RequiredRule makes field required and does NOT allow just spaces
	RequiredRule = &validateRequiredRule{err: errors.New(RequiredTxt)}

	// DefaultPathRegex is the standard setting returned from request
	DefaultPathRegex = func(r *http.Request) (string, error) {
		return mux.CurrentRoute(r).GetPathTemplate()
	}

	// DefaultVars returns mux vars from request
	DefaultVars = func(r *http.Request) map[string]string {
		return mux.Vars(r)
	}
)

//////////////////////////////////////////////////////////////////
//----------------------- CUSTOM ERRORS -------------------------
//////////////////////////////////////////////////////////////////

var (
	// errFutureAndPastDateInternal returns for form field if
	// user sets that a date can not be both a future or past date
	errFutureAndPastDateInternal = errors.New("webutil: both 'canBeFuture and 'canBePast' can not be false")

	// errInvalidStringInternal returns for form field if
	// data type for date field is not "string" or "*string"
	errInvalidStringInternal = errors.New("webutil: input must be string or *string")

	// errInvalidFormSelectionInternal will be returned if a query does not
	// return two columns when trying to query for FormSelections
	errInvalidFormSelectionInternal = errors.New("webutil: query should only return 2 columns")

	// errInvalidValidationTypeInternal will be returned from validation functions i.e.
	// ValidateIDs, ValidateUniquness, etc. if the value that is being validated
	// is not a primitive or slice or primitive types
	errInvalidValidationTypeInternal = errors.New("webutil: form validation type must be primitive or slice of primitives")

	// errInvalidCacheTypeInternal will be returned if results from cache is not a string or
	// number type or an array of string and/or numbers
	errInvalidCacheTypeInternal = errors.New("webutil: cache result types must be string or number or array of string and number")

	// errNilValidation represents when we return a nil error in validation rules
	// because the validated value is empty so we can't continue validating
	errNilValidation = errors.New("webutil: nil validation error")

	// errInvalidValidateValue represents error where a value passed to form validation can't be a struct
	// only primative types
	errInvalidValidateValue = errors.New("webutil: validate value is invalid type")

	// errInvalidInstanceValue represents error if instance value passed to validator is not primitive type
	errInvalidInstanceValue = errors.New("webutil: instance value is invalid type")

	errIncompatableTypes = errors.New("webutil: instance value and passed are different types")
)

//////////////////////////////////////////////////////////////////
//-------------------------- TYPES ----------------------------
//////////////////////////////////////////////////////////////////

// PathRegex is a work around for the fact that injecting and retrieving a route into
// mux is quite complex without spinning up an entire server
type PathRegex func(r *http.Request) (string, error)

// RecoverForm should implement ability to recover from form error
type RecoverForm func(error) error

//////////////////////////////////////////////////////////////////
//---------------------- CONFIG STRUCTS ------------------------
//////////////////////////////////////////////////////////////////

// FormValidationConfig is config struct used in the initialization
// of *FormValidation
type FormValidationConfig struct {
	Cache      CacheStore
	PathRegex  PathRegex
	SQLBindVar int
}

// CacheValidate is config struct used in form validators
type CacheValidate struct {
	// Key is key used in cache to retrieve value
	// Key can be string format such as "email-%s"
	// in which case the KeyArgs property and/or current validated
	// value (if property KeyIdx > -1) will be used
	// to dynamically fill in to be used as key for cache
	Key string

	// KeyArgs is used against the Key property to dynamically
	// change the key for cache as the validator functions
	// will use fmt.Sprint(key, args...) to format the key
	KeyArgs []interface{}

	// KeyIdx is used if current value being validated
	// in a validator function should be used as apart of
	// the key
	//
	// If it is, it will be added to given index of the KeyArgs property
	// If KeyIdx is below 0, validated value won't be added to key
	KeyIdx int

	// PropertyName is the name of field that we are searching
	// for to compare against the current validated value
	// If the results from cache is a primitive type or an array
	// of primitive types, then this property will be ignored
	PropertyName string

	// IgnoreCacheNil will ignore cache nil error and still query db
	IgnoreCacheNil bool

	// IgnoreInvalidCacheResults will ignore if we can't find
	// any results in cache and query db anyways
	IgnoreInvalidCacheResults bool

	// IgnoreValidatedTypes will convert current value being
	// validated and cache to string
	// If value is slice, it will convert all values within slice
	// to string
	IgnoreTypes bool
}

//////////////////////////////////////////////////////////////////
//------------------------- STRUCTS ---------------------------
//////////////////////////////////////////////////////////////////

// FormCurrency is struct used to deal with form fields that involve currency
//
// The reason for having this is that floats do not give exact returns all the
// time so FormCurrency (which embeds the github.com/shopspring/decimal library)
// is for manipulating currency with exact returns
//
// Another reason for this struct is that it does not play nice with the ozzo
// validators such a "Min" so we much implement that logic here
type FormCurrency struct {
	decimal.Decimal

	// CurrencyRegex is used for form validation on validating decimal.Decimal
	// is in the right format
	// This will be used in FormCurrency#Validate function
	//
	// Default: USDCurrencyRegex
	CurrencyRegex *regexp.Regexp `json:"-"`

	// AllowNegative allows decimal number to be negative
	// This will be used in FormCurrency#Validate function
	//
	// Default: false
	// AllowNegative bool `json:"-"`

	// Min is the lowest number decimal allowed
	//
	// Default: nil (no limit)
	Min *decimal.Decimal `json:"-"`

	// MinError is a custom error message a user can set if
	// decimal is lower than Min
	MinError error `json:"-"`

	// Max is the highest number decimal allowed
	//
	// Default: nil (no limit)
	Max *decimal.Decimal `json:"-"`

	// MaxError is a custom error message a user can set if
	// decimal is higher than Max
	MaxError error `json:"-"`
}

func (f *FormCurrency) UnmarshalJSON(decimalBytes []byte) error {
	return f.Decimal.UnmarshalJSON(decimalBytes)
}

func (f FormCurrency) MarshalJSON() ([]byte, error) {
	return f.Decimal.MarshalJSON()
}

func (f FormCurrency) Validate() error {
	var currencyRegexp *regexp.Regexp

	if f.CurrencyRegex == nil {
		currencyRegexp = USDCurrencyRegex
	} else {
		currencyRegexp = f.CurrencyRegex
	}

	if !currencyRegexp.MatchString(f.Decimal.String()) {
		fmt.Printf("format: %s\n", f.Decimal.String())
		return fmt.Errorf(InvalidFormatTxt)
	}

	if f.Min != nil && f.Decimal.LessThan(*f.Min) {
		if f.MinError != nil {
			return f.MinError
		}

		val, _ := f.Min.Float64()
		return fmt.Errorf("Can't be less than %v", val)
	}

	if f.Max != nil && f.Decimal.GreaterThan(*f.Max) {
		if f.MaxError != nil {
			return f.MaxError
		}

		val, _ := f.Max.Float64()
		return fmt.Errorf("Can't be greater than %v", val)
	}

	return nil
}

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
	return json.Marshal(strconv.FormatInt(int64(i), webutilcfg.IntBase))
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

		value, err := strconv.ParseInt(s, webutilcfg.IntBase, webutilcfg.IntBitSize)
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
func (i Int64) Int64Val() int64 {
	return int64(i)
}

func (i Int64) Str() string {
	return strconv.FormatInt(int64(i), webutilcfg.IntBase)
}

func (i *Int64) Scan(value interface{}) error {
	switch value.(type) {
	case string:
		num, err := strconv.ParseInt(value.(string), 10, 64)

		if err != nil {
			return err
		}

		*i = Int64(num)
	case int64:
		*i = Int64(value.(int64))
	default:
		return errors.New("webutil: Invalid data type for Int64")
	}

	return nil
}

func (i Int64) Value() (driver.Value, error) {
	return int64(i), nil
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
// Will raise errFutureAndPastDateInternal error which will be wrapped
// in validation.InternalError if both bool parameters are false
func (f *FormValidation) ValidateDate(
	layout,
	timezone string,
	compareTime,
	canBeFuture,
	canBePast bool,
) *validateDateRule {
	return &validateDateRule{
		layout:      layout,
		compareTime: compareTime,
		timezone:    timezone,
		canBeFuture: canBeFuture,
		canBePast:   canBePast,
	}
}

// ValidateArgs determines whether validated field(s) exists in
// database or cache if set
// The same number of validated fields must be returned from
// database or cache in order to be true
//
// If validated field is single val, then only one value
// must be returned from database or cache to be true
//
// If validated field is slice, then number of results
// that must come from database or cache should be
// length of slice
func (f *FormValidation) ValidateArgs(
	cacheValidate *CacheValidate,
	placeHolderIdx int,
	query string,
	args ...interface{},
) *validateArgsRule {
	return &validateArgsRule{
		validator: &validator{
			querier:        f.entity,
			cache:          f.config.Cache,
			cacheValidate:  cacheValidate,
			placeHolderIdx: placeHolderIdx,
			bindVar:        f.config.SQLBindVar,
			query:          query,
			args:           args,
			err:            errors.New(InvalidTxt),
		},
	}
}

// ValidateUniqueness determines whether validated field is unique
// within database or cache if set
func (f *FormValidation) ValidateUniqueness(
	cacheValidate *CacheValidate,
	instanceValue interface{},
	placeHolderIdx int,
	query string,
	args ...interface{},
) *validateUniquenessRule {
	return &validateUniquenessRule{
		instanceValue: instanceValue,
		validator: &validator{
			querier:        f.entity,
			cache:          f.config.Cache,
			cacheValidate:  cacheValidate,
			placeHolderIdx: placeHolderIdx,
			bindVar:        f.config.SQLBindVar,
			query:          query,
			args:           args,
			err:            errors.New(AlreadyExistsTxt),
		},
	}
}

// ValidateExists determines whether validated field exists
// within database or cache if set
// Only has to find one record to be true
func (f *FormValidation) ValidateExists(
	cacheValidate *CacheValidate,
	placeHolderIdx int,
	query string,
	args ...interface{},
) *validateExistsRule {
	return &validateExistsRule{
		validator: &validator{
			querier:        f.entity,
			cacheValidate:  cacheValidate,
			placeHolderIdx: placeHolderIdx,
			bindVar:        f.config.SQLBindVar,
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
	//entityRecover  EntityRecover
	querier        Querier
	cache          CacheStore
	args           []interface{}
	query          string
	bindVar        int
	placeHolderIdx int
	err            error
	recoverDB      RecoverDB
	// recoverDB      RecoverError
	cacheValidate *CacheValidate
}

type validateRequiredRule struct {
	err error
}

func (v *validateRequiredRule) Validate(value interface{}) error {
	var err error

	checkValue := func(val string) error {
		val = strings.TrimSpace(val)

		if len(val) == 0 {
			return v.err
		}

		return nil
	}

	if isNilValue(value) {
		return v.err
	}

	switch value.(type) {
	case string:
		err = checkValue(value.(string))
	case *string:
		temp := value.(*string)
		err = checkValue(*temp)
	default:
		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(value)

			for i := 0; i < s.Len(); i++ {
				sliceVal := s.Index(i).Interface()

				switch sliceVal.(type) {
				case string:
					err = checkValue(sliceVal.(string))
				case *string:
					temp := sliceVal.(*string)
					err = checkValue(*temp)
				}

				if err != nil {
					return v.err
				}
			}
		}

		return validation.Required.Validate(value)
	}

	return err
}

func (v *validateRequiredRule) Error(message string) *validateRequiredRule {
	return &validateRequiredRule{
		err: errors.New(message),
	}
}

type validateDateRule struct {
	layout      string
	timezone    string
	compareTime bool
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
		if v.compareTime {
			currentTime, err = GetCurrentLocalDateTimeInUTC(v.timezone)
		} else {
			currentTime, err = GetCurrentLocalDateInUTC(v.timezone)
		}

		if err != nil {
			return validation.NewInternalError(err)
		}
	} else {
		if v.compareTime {
			currentTime, err = GetCurrentLocalDateTimeInUTC("UTC")
		} else {
			currentTime, err = GetCurrentLocalDateInUTC("UTC")
		}
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
		err = validation.NewInternalError(errFutureAndPastDateInternal)
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
	return validatorRules(v.validator, value, validateExistsType)
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
	var validateVal, instanceVal interface{}
	var err error

	validateValOf := reflect.ValueOf(value)
	instanceValOf := reflect.ValueOf(v.instanceValue)
	compareValues := true

	if !validateValOf.IsValid() || !instanceValOf.IsValid() {
		compareValues = false
	}
	if validateValOf.Kind() == reflect.Ptr && validateValOf.IsNil() {
		compareValues = false
	}
	if instanceValOf.Kind() == reflect.Ptr && instanceValOf.IsNil() {
		compareValues = false
	}

	if compareValues {
		findVal := func(kind reflect.Kind, isPointer bool, forValidateVal bool) error {
			kindList := []reflect.Kind{
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64, reflect.String,
			}

			found := false

			for _, v := range kindList {
				if kind == v {
					found = true
				}
			}

			if found {
				if forValidateVal {
					if isPointer {
						validateVal = validateValOf.Elem().Interface()
						//fmt.Printf("here foo 1: %v\n", reflect.ValueOf(validateVal).Kind())
					} else {
						validateVal = validateValOf.Interface()
						//fmt.Printf("here foo 2: %v\n", reflect.ValueOf(validateVal).Kind())
					}
				} else {
					if isPointer {
						instanceVal = instanceValOf.Elem().Interface()
						//fmt.Printf("here foo 3: %v\n", reflect.ValueOf(instanceVal).Kind())
					} else {
						instanceVal = instanceValOf.Interface()
						//fmt.Printf("here foo 4: %v\n", reflect.ValueOf(instanceVal).Kind())
					}
				}
			} else {
				if forValidateVal {
					return errInvalidValidateValue
				}

				return errInvalidInstanceValue
			}

			return nil
		}

		var validateKind, instanceKind reflect.Kind

		if validateValOf.Kind() == reflect.Ptr {
			validateKind = validateValOf.Elem().Kind()
			//fmt.Printf("validate kind 1: %v\n", validateKind)

			if err = findVal(validateKind, true, true); err != nil {
				return err
			}
		} else {
			validateKind = validateValOf.Kind()
			//fmt.Printf("validate kind 2: %v\n", validateKind)

			if err = findVal(validateKind, false, true); err != nil {
				return err
			}
		}

		if instanceValOf.Kind() == reflect.Ptr {
			instanceKind = instanceValOf.Elem().Kind()
			//fmt.Printf("instance kind 1: %v\n", instanceKind)

			if err = findVal(instanceKind, true, false); err != nil {
				return err
			}
		} else {
			instanceKind = instanceValOf.Kind()
			//fmt.Printf("instance kind 2: %v\n", instanceKind)

			if err = findVal(instanceKind, false, false); err != nil {
				return err
			}
		}

		if validateKind != instanceKind {
			return errIncompatableTypes
		}

		// fmt.Printf("validate val type: %s\n", reflect.TypeOf(validateVal).Kind())
		// fmt.Printf("instance val type: %s\n", reflect.TypeOf(instanceVal).Kind())

		if instanceVal == validateVal {
			return nil
		}
	}

	return validatorRules(v.validator, value, validateUniquenessType)
}

func (v *validateUniquenessRule) Error(message string) *validateUniquenessRule {
	v.err = errors.New(message)
	return v
}

type validateArgsRule struct {
	*validator
}

func (v *validateArgsRule) Validate(value interface{}) error {
	return validatorRules(v.validator, value, validateArgsType)
}

func (v *validateArgsRule) Error(message string) *validateArgsRule {
	v.err = errors.New(message)
	return v
}

//////////////////////////////////////////////////////////////////
//----------------------- FUNCTIONS -------------------------
//////////////////////////////////////////////////////////////////

// HasFormErrors determines what type of form error is passed and
// sends appropriate error message to client
//
// If err is caused by server from form validator, then we try to recover
// and retry the form if user has set this
// If that fails, we write server error back to client and log if set
func HasFormErrors(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	retryForm func() error,
	clientStatus int,
	conf ServerErrorConfig,
) bool {
	hasError := false
	//hasFormError := false

	if err != nil {
		logConf := LogConfig{CauseErr: err}
		SetHTTPResponseDefaults(&conf.ServerErrorResponse, 500, []byte(serverErrTxt))

		var valErr validation.Errors

		hasFormErrFn := func() bool {
			if errors.Is(err, ErrBodyRequired) {
				hasError = true
				w.WriteHeader(clientStatus)
				w.Write([]byte(bodyRequiredTxt))
			} else if errors.Is(err, ErrInvalidJSON) {
				hasError = true
				w.WriteHeader(clientStatus)
				w.Write([]byte(invalidJSONTxt))
			} else if errors.As(err, &valErr) {
				hasError = true
				jsonString, _ := json.Marshal(errors.Cause(err).(validation.Errors))
				w.WriteHeader(clientStatus)
				w.Write(jsonString)
			}

			return hasError
		}

		serverErrFn := func() {
			w.WriteHeader(*conf.ServerErrorResponse.HTTPStatus)
			w.Write(conf.ServerErrorResponse.HTTPResponse)
			hasError = true
		}

		if !hasFormErrFn() {
			if conf.RecoverForm != nil {
				if formErr := conf.RecoverForm(err); formErr != nil {
					logConf.RecoverFormErr = formErr
					serverErrFn()
				} else {
					if retryErr := retryForm(); retryErr != nil {
						logConf.RetryFormErr = retryErr

						if !hasFormErrFn() {
							serverErrFn()
						}
					}
				}
			} else {
				serverErrFn()
			}
		}

		if conf.Logger != nil {
			conf.Logger(r, logConf)
		}
	}

	return hasError
}

func FormHasErrors(
	w http.ResponseWriter,
	err error,
	clientStatus int,
	serverResp HTTPResponseConfig,
) bool {
	if err != nil {
		SetHTTPResponseDefaults(&serverResp, http.StatusInternalServerError, []byte(serverErrTxt))

		hasFormError := false

		var valErr validation.Errors

		if errors.Is(err, ErrBodyRequired) {
			hasFormError = true
			w.WriteHeader(clientStatus)
			w.Write([]byte(bodyRequiredTxt))
		} else if errors.Is(err, ErrInvalidJSON) {
			hasFormError = true
			w.WriteHeader(clientStatus)
			w.Write([]byte(invalidJSONTxt))
		} else if errors.As(err, &valErr) {
			hasFormError = true
			jsonString, _ := json.Marshal(errors.Cause(err).(validation.Errors))
			w.WriteHeader(clientStatus)
			w.Write(jsonString)
		}

		if !hasFormError {
			w.WriteHeader(*serverResp.HTTPStatus)
			w.Write(serverResp.HTTPResponse)
		}

		return true
	}

	return false
}

func FormHasErrorsL(
	w http.ResponseWriter,
	err error,
	logFunc func(err error),
	clientStatus int,
	serverResp HTTPResponseConfig,
) bool {
	if err != nil {
		SetHTTPResponseDefaults(&serverResp, http.StatusInternalServerError, []byte(serverErrTxt))

		hasFormError := false

		var valErr validation.Errors

		if errors.Is(err, ErrBodyRequired) {
			hasFormError = true
			w.WriteHeader(clientStatus)
			w.Write([]byte(bodyRequiredTxt))
		} else if errors.Is(err, ErrInvalidJSON) {
			hasFormError = true
			w.WriteHeader(clientStatus)
			w.Write([]byte(invalidJSONTxt))
		} else if errors.As(err, &valErr) {
			hasFormError = true
			jsonString, _ := json.Marshal(errors.Cause(err).(validation.Errors))
			w.WriteHeader(clientStatus)
			w.Write(jsonString)
		}

		if !hasFormError {
			w.WriteHeader(*serverResp.HTTPStatus)
			w.Write(serverResp.HTTPResponse)
		}

		if logFunc != nil {
			logFunc(err)
		}

		return true
	}

	return false
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
			fmt.Printf("%s", err.Error())
			return ErrInvalidJSON
		}
	} else {
		if !canSkip {
			return ErrBodyRequired
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
// string or *string type and if not, returns errInvalidStringInternal
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
		err = validation.NewInternalError(errInvalidStringInternal)
	}

	return dateValue, err
}

func queryValidatorRows(v *validator, q string, args ...interface{}) (*sqlx.Rows, error) {
	var rows *sqlx.Rows
	var err error

	if rows, err = v.querier.Queryx(q, args...); err != nil {
		if v.recoverDB != nil {
			if db, err := v.recoverDB(err); err == nil {
				v.querier = db
				//v.entityRecover.SetEntity(db)

				return v.querier.Queryx(q, args...)
			}
		}
	}

	return rows, err
}

func cacheResults(v *validator, value interface{}, queryFunc func() error, validatorType int) error {
	var err error
	var singleVal interface{}
	var searchVals []interface{}
	var cacheBytes []byte

	if isNilValue(value) {
		return errNilValidation
	}

	// First determine whether the interface is slice or not
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(value)

		for k := 0; k < s.Len(); k++ {
			i := s.Index(k).Interface()

			if v.cacheValidate.IgnoreTypes {
				t, err := convertToStr(i)

				if err != nil {
					return validation.NewInternalError(errInvalidValidationTypeInternal)
				}

				searchVals = append(searchVals, t)
			} else {
				searchVals = append(searchVals, i)
			}
		}
	default:
		if v.cacheValidate.IgnoreTypes {
			singleVal, err = convertToStr(value)

			if err != nil {
				return validation.NewInternalError(errInvalidValidationTypeInternal)
			}
		} else {
			singleVal = value
		}
	}

	var cacheArgs []interface{}

	// If KeyIdx is greater than -1, then add current
	// validated value to cache args slice to be used
	// as apart of the cache key along with passed args
	// Else just use passed args
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

	// If error occurs, check what type and perform
	// action depending on config settings passed
	if err != nil {
		if err == ErrCacheNil {
			if v.cacheValidate.IgnoreCacheNil {
				return queryFunc()
			}

			switch validatorType {
			case validateArgsType, validateExistsType:
				return v.err
			case validateUniquenessType:
				return nil
			}
		} else {
			return queryFunc()
		}
	} else {
		var cacheHolding interface{}
		var val interface{}
		var ok bool

		counter := 0

		if err = json.Unmarshal(cacheBytes, &cacheHolding); err != nil {
			return validation.NewInternalError(err)
		}

		counterFunc := func(val interface{}) {
			if singleVal != nil {
				if val == singleVal {
					counter++
				}
			} else {
				for _, t := range searchVals {
					if val == t {
						counter++
					}
				}
			}
		}

		// Determine whether results from cache is a json object,
		// json array or single value
		switch reflect.TypeOf(cacheHolding).Kind() {
		case reflect.Slice:
			cacheSlice := cacheHolding.([]interface{})

			for _, f := range cacheSlice {
				// Determine type of each entity in array
				// Generally people are going to save the same type
				// within an array but here we are giving flexibility
				// to have different types
				switch f.(type) {
				case map[string]interface{}:
					cacheMap := f.(map[string]interface{})

					if val, ok = cacheMap[v.cacheValidate.PropertyName]; ok {
						if v.cacheValidate.IgnoreTypes {
							if val, err = convertToStr(val); err != nil {
								return validation.NewInternalError(errInvalidCacheTypeInternal)
							}
						}
					}
				case float64, string:
					if v.cacheValidate.IgnoreTypes {
						if val, err = convertToStr(f); err != nil {
							return validation.NewInternalError(errInvalidCacheTypeInternal)
						}
					} else {
						val = f
					}
				default:
					return validation.NewInternalError(errInvalidCacheTypeInternal)
				}

				counterFunc(val)
			}
		case reflect.Map:
			cacheMap := cacheHolding.(map[string]interface{})

			if val, ok = cacheMap[v.cacheValidate.PropertyName]; ok {
				if v.cacheValidate.IgnoreTypes {
					if val, err = convertToStr(val); err != nil {
						return validation.NewInternalError(errInvalidCacheTypeInternal)
					}
				}

				counterFunc(val)
			}
		default:
			if v.cacheValidate.IgnoreTypes {
				if val, err = convertToStr(cacheHolding); err != nil {
					return validation.NewInternalError(errInvalidCacheTypeInternal)
				}
			} else {
				val = cacheHolding
			}

			counterFunc(val)
		}

		var expectedLen int

		if singleVal != nil {
			expectedLen = 1
		} else {
			expectedLen = len(searchVals)
		}

		switch validatorType {
		case validateArgsType:
			if counter != expectedLen {
				if v.cacheValidate.IgnoreInvalidCacheResults {
					return queryFunc()
				}

				return v.err
			}
		case validateExistsType:
			if counter == 0 {
				if v.cacheValidate.IgnoreInvalidCacheResults {
					return queryFunc()
				}

				return v.err
			}
		case validateUniquenessType:
			if counter > 0 {
				if v.cacheValidate.IgnoreInvalidCacheResults {
					return queryFunc()
				}

				return v.err
			}
		}

		return nil
	}

	return nil
}

// func queryResults(v *validator, value interface{}) (string, []interface{}, int, error) {
// 	var err error
// 	var searchVals []interface{}
// 	var singleVal interface{}
// 	var expectedLen int
// 	var q string

// 	if isNilValue(value) {
// 		return "", nil, 0, errNilValidation
// 	}

// 	isSlice := false

// 	// First determine whether the interface is slice or not
// 	switch reflect.TypeOf(value).Kind() {
// 	case reflect.Slice:
// 		isSlice = true
// 		s := reflect.ValueOf(value)
// 		expectedLen = s.Len()

// 		for k := 0; k < s.Len(); k++ {
// 			i := s.Index(k).Interface()
// 			searchVals = append(searchVals, i)
// 		}
// 	default:
// 		expectedLen = 1
// 		singleVal = value
// 	}

// 	// If type is slice and is empty, simply return nil as we will get an error
// 	// when trying to query with empty slice
// 	if isSlice && len(searchVals) == 0 {
// 		return "", nil, 0, nil
// 	}

// 	var args []interface{}

// 	if v.placeHolderIdx > -1 {
// 		args = make([]interface{}, 0, len(v.args)+1)
// 		args = append(args, v.args...)

// 		if isSlice {
// 			args = InsertAt(args, searchVals, v.placeHolderIdx)
// 		} else {
// 			args = InsertAt(args, singleVal, v.placeHolderIdx)
// 		}
// 	} else {
// 		args = make([]interface{}, 0, len(v.args))
// 		args = append(args, v.args...)
// 	}

// 	q, args, err = InQueryRebind(v.bindVar, v.query, args...)

// 	if err != nil {
// 		messageErr := fmt.Errorf("err: %s\n query: %s\n args:%v\n", err, q, args)
// 		return "", nil, 0, errors.WithStack(validation.NewInternalError(messageErr))
// 	}

// 	return q, args, expectedLen, nil
// }

func convertToStr(typ interface{}) (interface{}, error) {
	var t interface{}

	switch typ.(type) {
	case int:
		t = strconv.Itoa(typ.(int))
	case *int:
		i := typ.(*int)
		t = strconv.Itoa(*i)
	case int64:
		t = strconv.FormatInt(typ.(int64), webutilcfg.IntBase)
	case *int64:
		i := typ.(*int64)
		t = strconv.FormatInt(*i, webutilcfg.IntBase)
	case float64:
		t = strconv.FormatFloat(typ.(float64), 'f', -1, webutilcfg.IntBitSize)
	case *float64:
		i := typ.(*float64)
		t = strconv.FormatFloat(*i, 'f', -1, webutilcfg.IntBitSize)
	case string:
		t = typ.(string)
	case *string:
		i := typ.(*string)
		t = *i
	default:
		return nil, errors.New("error")
	}

	return t, nil
}

func queryResults(v *validator, value interface{}) (string, []interface{}, int, error) {
	var err error
	var searchVals []interface{}
	var singleVal interface{}
	var expectedLen int
	var q string

	if isNilValue(value) {
		return "", nil, 0, errNilValidation
	}

	isSlice := false

	// First determine whether the interface is slice or not
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		isSlice = true
		s := reflect.ValueOf(value)
		expectedLen = s.Len()

		for k := 0; k < s.Len(); k++ {
			i := s.Index(k).Interface()
			searchVals = append(searchVals, i)
		}
	default:
		expectedLen = 1
		singleVal = value
	}

	// If type is slice and is empty, simply return nil as we will get an error
	// when trying to query with empty slice
	if isSlice && len(searchVals) == 0 {
		return "", nil, 0, nil
	}

	var args []interface{}

	if v.placeHolderIdx > -1 {
		args = make([]interface{}, 0, len(v.args)+1)
		args = append(args, v.args...)

		if isSlice {
			args = InsertAt(args, searchVals, v.placeHolderIdx)
		} else {
			args = InsertAt(args, singleVal, v.placeHolderIdx)
		}
	} else {
		args = make([]interface{}, 0, len(v.args))
		args = append(args, v.args...)
	}

	q, args, err = InQueryRebind(v.bindVar, v.query, args...)

	if err != nil {
		messageErr := fmt.Errorf("err: %s\n query: %s\n args:%v\n", err, q, args)
		return "", nil, 0, errors.WithStack(validation.NewInternalError(messageErr))
	}

	return q, args, expectedLen, nil
}

func validatorRules(v *validator, value interface{}, validateType int) error {
	var err error
	var expectedLen int
	var tmpVal interface{}
	var args []interface{}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(value)
		expectedLen = s.Len()

		var searchVals []interface{}

		for k := 0; k < s.Len(); k++ {
			i := s.Index(k).Interface()
			searchVals = append(searchVals, i)
		}

		// If type is slice and is empty, simply return nil as we will get an error
		// when trying to query with empty slice
		if len(searchVals) == 0 {
			return nil
		}

		tmpVal = searchVals
	default:
		tmpVal = value
		expectedLen = 1
	}

	args = make([]interface{}, 0, len(v.args)+1)
	args = append(args, v.args...)

	if v.placeHolderIdx > -1 {
		args = InsertAt(args, tmpVal, v.placeHolderIdx)
	}

	rows, err := v.querier.QueryxRebind(v.bindVar, v.query, args...)

	if err != nil {
		msg := fmt.Errorf("err: %s\n query:%s\n args:%v\n", err, v.query, args)
		return errors.WithStack(validation.NewInternalError(msg))
	}

	counter := 0

	for rows.Next() {
		counter++
	}

	switch validateType {
	case validateArgsType:
		if counter != expectedLen {
			return v.err
		}
	case validateUniquenessType:
		if counter > 0 {
			return v.err
		}
	case validateExistsType:
		if counter == 0 {
			return v.err
		}
	}

	return nil
}
