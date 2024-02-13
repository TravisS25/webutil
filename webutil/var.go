package webutil

import (
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

//////////////////////////////////////////////////////////////////
//------------------------ ERROR TYPES -------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrEmptyConfigList is error returned when trying to recover
	// from database error and there is no backup configs set up
	ErrEmptyConfigList = errors.New("webutil: empty config list")

	// ErrNoConnection is error returned when there is no
	// connection to database available
	ErrNoConnection = errors.New("webutil: connection could not be established")

	// ErrInvalidDBType is error returned when trying to pass an invalid
	// database type string to function
	ErrInvalidDBType = errors.New("webutil: invalid database type")

	// ErrInvalidSort is error returned if client tries
	// to pass filter parameter that is not sortable
	ErrInvalidSort = errors.New("webutil: invalid sort")

	// ErrInvalidArray is error returned if client tries
	// to pass array parameter that is invalid array type
	ErrInvalidArray = errors.New("webutil: invalid array for field")

	// ErrInvalidValue is error returned if client tries
	// to pass filter parameter that had invalid field
	// value for certain field
	ErrInvalidValue = errors.New("webutil: invalid field value")
)

//////////////////////////////////////////////////////////////////
//-------------------------- REGEX -----------------------------
//////////////////////////////////////////////////////////////////

var (
	// EmailRegex is regex expression used for forms to validate email
	EmailRegex = regexp.MustCompile("^.+@[a-zA-Z0-9.]+$")

	// ZipRegex is regex expression used for forms to validate zip code
	ZipRegex = regexp.MustCompile("^[0-9]{5}$")

	// PhoneNumberRegex is regex expression used for forms to validate phone number
	PhoneNumberRegex = regexp.MustCompile(`^\([0-9]{3}\)-[0-9]{3}-[0-9]{4}$`)

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
	USDCurrencyRegex = regexp.MustCompile(`^(?:-)?[0-9]+(?:\\.[0-9]{1,2})?$`)
)

//////////////////////////////////////////////////////////////////
//----------------------- PRE-BUILT RULES ----------------------
//////////////////////////////////////////////////////////////////

var (
	// RequiredRule makes field required and does NOT allow just spaces
	RequiredRule = &validateRequiredRule{err: errors.New(REQUIRED_TXT)}
)

//////////////////////////////////////////////////////////////////
//--------------------------- DEFAULTS -------------------------
//////////////////////////////////////////////////////////////////

var (
	// DefaultPathRegex is the standard setting returned from request
	DefaultPathRegex = func(r *http.Request) (string, error) {
		return mux.CurrentRoute(r).GetPathTemplate()
	}

	// DefaultVars returns mux vars from request
	DefaultVars = func(r *http.Request) map[string]string {
		return mux.Vars(r)
	}
)
