package webutil

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-jet/jet/v2/qrm"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
)

const (
	EMPTY_UUID = "00000000-0000-0000-0000-000000000000"
	EMPTY_TIME = "0001-01-01 00:00:00 +0000 UTC"
)

// FormValidationConfig is config struct used in the initialization
// of *JetFormValidation
type JetFormValidationConfig struct {
	PathRegex  PathRegex
	SQLBindVar int
}

// JetFormValidation is the main struct that other structs will
// embed to validate json data
type JetFormValidation struct {
	queryable qrm.Queryable
	config    JetFormValidationConfig
}

// NewFormValidation returns *JetFormValidation instance
func NewJetFormValidation(queryable qrm.Queryable, config JetFormValidationConfig) *JetFormValidation {
	return &JetFormValidation{
		queryable: queryable,
		config:    config,
	}
}

// IsValid returns *validRuleJet based on isValid parameter
// Basically IsValid is a wrapper for the passed bool
// to return valid rule to then apply custom error message
// for the Error function
func (f *JetFormValidation) IsValid(isValid bool) *validRuleJet {
	return &validRuleJet{isValid: isValid, err: errors.New("Not Valid")}
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
func (f *JetFormValidation) ValidateDate(
	layout,
	timezone string,
	compareTime,
	canBeFuture,
	canBePast bool,
) *validateDateRuleJet {
	return &validateDateRuleJet{
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
func (f *JetFormValidation) ValidateArgs(
	placeHolderIdx int,
	query string,
	args ...interface{},
) *validateArgsRuleJet {
	return &validateArgsRuleJet{
		validatorJet: &validatorJet{
			queryable:      f.queryable,
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
func (f *JetFormValidation) ValidateUniqueness(
	instanceValue interface{},
	placeHolderIdx int,
	query string,
	args ...interface{},
) *validateUniquenessRuleJet {
	return &validateUniquenessRuleJet{
		instanceValue: instanceValue,
		validatorJet: &validatorJet{
			queryable:      f.queryable,
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
func (f *JetFormValidation) ValidateExists(
	placeHolderIdx int,
	query string,
	args ...interface{},
) *validateExistsRuleJet {
	return &validateExistsRuleJet{
		validatorJet: &validatorJet{
			queryable:      f.queryable,
			placeHolderIdx: placeHolderIdx,
			bindVar:        f.config.SQLBindVar,
			query:          query,
			args:           args,
			err:            errors.New(DoesNotExistTxt),
		},
	}
}

// GetEntity returns Entity
func (f *JetFormValidation) GetQueryable() qrm.Queryable {
	return f.queryable
}

// GetConfig return FormValidationConfig
func (f *JetFormValidation) GetConfig() JetFormValidationConfig {
	return f.config
}

// SetEntity sets Entity
func (f *JetFormValidation) SetQueryable(queryable qrm.Queryable) {
	f.queryable = queryable
}

// SetConfig sets FormValidationConfig
func (f *JetFormValidation) SetConfig(config JetFormValidationConfig) {
	f.config = config
}

type validatorJet struct {
	queryable      qrm.Queryable
	args           []interface{}
	query          string
	bindVar        int
	placeHolderIdx int
	err            error
}

type validateDateRuleJet struct {
	layout      string
	timezone    string
	compareTime bool
	canBeFuture bool
	canBePast   bool
	err         error
}

func (v *validateDateRuleJet) Validate(value interface{}) error {
	var currentTime, dateTime time.Time
	var err error
	var dateValue string

	if isNilValue(value) {
		return nil
	}

	if dateValue, err = getStringFromValue(value, v.layout); err != nil {
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

		if err != nil {
			return validation.NewInternalError(err)
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

func (v *validateDateRuleJet) Error(message string) *validateDateRuleJet {
	v.err = errors.New(message)
	return v
}

type validRuleJet struct {
	isValid bool
	err     error
}

func (v *validRuleJet) Validate(value interface{}) error {
	if !v.isValid {
		return v.err
	}

	return nil
}

// Error sets the error message for the rule.
func (v *validRuleJet) Error(message string) *validRuleJet {
	v.err = errors.New(message)
	return v
}

type validateExistsRuleJet struct {
	*validatorJet
}

func (v *validateExistsRuleJet) Validate(value interface{}) error {
	return validatorRulesJet(v.validatorJet, value, validateExistsType)
}

func (v *validateExistsRuleJet) Error(message string) *validateExistsRuleJet {
	v.err = errors.New(message)
	return v
}

type validateUniquenessRuleJet struct {
	*validatorJet
	instanceValue interface{}
}

func (v *validateUniquenessRuleJet) Validate(value interface{}) error {
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

	return validatorRulesJet(v.validatorJet, value, validateUniquenessType)
}

func (v *validateUniquenessRuleJet) Error(message string) *validateUniquenessRuleJet {
	v.err = errors.New(message)
	return v
}

type validateArgsRuleJet struct {
	*validatorJet
}

func (v *validateArgsRuleJet) Validate(value interface{}) error {
	return validatorRulesJet(v.validatorJet, value, validateArgsType)
}

func (v *validateArgsRuleJet) Error(message string) *validateArgsRuleJet {
	v.err = errors.New(message)
	return v
}

func validatorRulesJet(v *validatorJet, value interface{}, validateType int) error {
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

	query, args, err := JetInQueryRebind(v.bindVar, v.query, args...)
	if err != nil {
		return err
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
