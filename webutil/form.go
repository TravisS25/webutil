package webutil

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-jet/jet/v2/qrm"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

//////////////////////////////////////////////////////////////////
//----------------------- CUSTOM ERRORS -------------------------
//////////////////////////////////////////////////////////////////

var (
	// errFutureAndPastDateInternal returns for form field if
	// user sets that a date can not be both a future or past date
	errFutureAndPastDateInternal = errors.New("webutil: both 'canBeFuture and 'canBePast' can not be false")
)

//////////////////////////////////////////////////////////////////
//-------------------------- TYPES ----------------------------
//////////////////////////////////////////////////////////////////

// PathRegex is a work around for the fact that injecting and retrieving a route into
// mux is quite complex without spinning up an entire server
type PathRegex func(r *http.Request) (string, error)

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

func NewFormCurrency(d decimal.Decimal) *FormCurrency {
	return &FormCurrency{
		Decimal: d,
	}
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
		return fmt.Errorf(INVALID_FORMAT_TXT)
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

// Int64 is mainly used for slice of id values from forms
type Int64 int64

// MarshalJSON takes int64 value and returns byte value
func (i Int64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(i), INT_BASE))
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

		value, err := strconv.ParseInt(s, INT_BASE, INT_BIT_SIZE)
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
func (i Int64) Raw() int64 {
	return int64(i)
}

func (i Int64) String() string {
	return strconv.FormatInt(int64(i), INT_BASE)
}

func (i *Int64) Scan(value interface{}) error {
	switch val := value.(type) {
	case string:
		num, err := strconv.ParseInt(val, 10, 64)

		if err != nil {
			return err
		}

		*i = Int64(num)
	case int64:
		*i = Int64(val)
	default:
		return errors.New("webutil: Invalid data type for Int64")
	}

	return nil
}

func (i Int64) Value() (driver.Value, error) {
	if i == 0 {
		return nil, nil
	}

	return int64(i), nil
}

type validateRequiredRule struct {
	err error
}

func (v *validateRequiredRule) Validate(value interface{}) error {
	var err error

	checkStrValue := func(val string) error {
		val = strings.TrimSpace(val)

		if len(val) == 0 {
			return v.err
		}

		return nil
	}

	checkUUIDValue := func(val uuid.UUID) error {
		if val.String() == EMPTY_UUID {
			return v.err
		}

		return nil
	}

	checkTimeValue := func(val time.Time) error {
		if val.String() == EMPTY_TIME {
			return v.err
		}

		return nil
	}

	if isNilValue(value) {
		return v.err
	}

	switch val := value.(type) {
	case string:
		err = checkStrValue(val)
	case *string:
		err = checkStrValue(*val)
	case uuid.UUID:
		err = checkUUIDValue(val)
	case *uuid.UUID:
		err = checkUUIDValue(*val)
	case time.Time:
		err = checkTimeValue(val)
	case *time.Time:
		err = checkTimeValue(*val)
	default:
		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(value)

			for i := 0; i < s.Len(); i++ {
				sliceVal := s.Index(i).Interface()

				switch sVal := sliceVal.(type) {
				case string:
					err = checkStrValue(sVal)
				case *string:
					err = checkStrValue(*sVal)
				case uuid.UUID:
					err = checkUUIDValue(sVal)
				case *uuid.UUID:
					err = checkUUIDValue(*sVal)
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

// FormValidationConfig is config struct used in the initialization
// of *FormValidation
type FormValidationConfig struct {
	PathRegex  PathRegex
	SQLBindVar int
}

// FormValidation is the main struct that other structs will
// embed to validate json data
type FormValidation struct {
	queryable qrm.Queryable
	config    FormValidationConfig
}

// NewFormValidation returns *FormValidation instance
func NewFormValidation(queryable qrm.Queryable, config FormValidationConfig) *FormValidation {
	return &FormValidation{
		queryable: queryable,
		config:    config,
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
	timezone string,
	compareTime,
	canBeFuture,
	canBePast bool,
) *validateDateRule {
	return &validateDateRule{
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
	placeHolderIdx int,
	query string,
	args ...interface{},
) *validateArgsRule {
	return &validateArgsRule{
		validator: &validator{
			queryable:      f.queryable,
			placeHolderIdx: placeHolderIdx,
			bindVar:        f.config.SQLBindVar,
			query:          query,
			args:           args,
			err:            errors.New(INVALID_TXT),
		},
	}
}

// ValidateUniqueness determines whether validated field is unique
// within database or cache if set
func (f *FormValidation) ValidateUniqueness(
	instanceValue interface{},
	placeHolderIdx int,
	query string,
	args ...interface{},
) *validateUniquenessRule {
	return &validateUniquenessRule{
		instanceValue: instanceValue,
		validator: &validator{
			queryable:      f.queryable,
			placeHolderIdx: placeHolderIdx,
			bindVar:        f.config.SQLBindVar,
			query:          query,
			args:           args,
			err:            errors.New(ALREADY_EXISTS_TXT),
		},
	}
}

// ValidateExists determines whether validated field exists
// within database or cache if set
// Only has to find one record to be true
func (f *FormValidation) ValidateExists(
	placeHolderIdx int,
	query string,
	args ...interface{},
) *validateExistsRule {
	return &validateExistsRule{
		validator: &validator{
			queryable:      f.queryable,
			placeHolderIdx: placeHolderIdx,
			bindVar:        f.config.SQLBindVar,
			query:          query,
			args:           args,
			err:            errors.New(DOES_NOT_EXIST_TXT),
		},
	}
}

// GetEntity returns Entity
func (f *FormValidation) GetQueryable() qrm.Queryable {
	return f.queryable
}

// GetConfig return FormValidationConfig
func (f *FormValidation) GetConfig() FormValidationConfig {
	return f.config
}

// SetEntity sets Entity
func (f *FormValidation) SetQueryable(queryable qrm.Queryable) {
	f.queryable = queryable
}

// SetConfig sets FormValidationConfig
func (f *FormValidation) SetConfig(config FormValidationConfig) {
	f.config = config
}

type validator struct {
	queryable      qrm.Queryable
	args           []interface{}
	query          string
	bindVar        int
	placeHolderIdx int
	err            error
}

type validateDateRule struct {
	timezone    string
	compareTime bool
	canBeFuture bool
	canBePast   bool
	err         error
}

func (v *validateDateRule) Validate(value interface{}) error {
	var currentTime, dateTime time.Time
	var err error

	if isNilValue(value) {
		return nil
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

		if err != nil {
			return validation.NewInternalError(err)
		}
	}

	switch val := value.(type) {
	case time.Time:
		dateTime = val
	case *time.Time:
		dateTime = *val
	default:
		return errors.New("Must be time.Time or *time.Time type")
	}

	if v.canBeFuture && v.canBePast {
		err = nil
	} else if v.canBeFuture {
		if dateTime.Before(currentTime) {
			err = errors.New(INVALID_PAST_DATE_TXT)
		}
	} else if v.canBePast {
		if dateTime.After(currentTime) {
			err = errors.New(INVALID_FUTURE_DATE_TXT)
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
	if v.instanceValue != nil {
		if reflect.TypeOf(value) != reflect.TypeOf(v.instanceValue) {
			return fmt.Errorf(
				"webutil: can't compare form value type of '%s' and instance value type of '%s'",
				reflect.TypeOf(value),
				reflect.TypeOf(v.instanceValue),
			)
		}

		if reflect.DeepEqual(value, v.instanceValue) {
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

func validatorRules(v *validator, value interface{}, validateType int) error {
	if isNilValue(value) {
		return nil
	}

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

	query, args, err := InQueryRebind(v.bindVar, v.query, args...)
	if err != nil {
		return errors.WithStack(validation.NewInternalError(err))
	}

	rows, err := v.queryable.QueryContext(context.Background(), query, args...)
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

// isNilValue determines if passed value is truley nil
func isNilValue(value interface{}) bool {
	_, isNil := validation.Indirect(value)
	if validation.IsEmpty(value) || isNil {
		return true
	}

	return false
}

// //////////////////////////////////////////////////////////////////
// //----------------------- FUNCTIONS -------------------------
// //////////////////////////////////////////////////////////////////

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
