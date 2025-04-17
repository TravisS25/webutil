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
	// ErrBodyRequired is used for when a post/put request does not contain a body in request
	ErrBodyRequired = errors.New("webutil: " + bodyRequiredTxt)

	// ErrInvalidJSON is used when there is an error unmarshalling a struct
	ErrInvalidJSON = errors.New("webutil: " + invalidJSONTxt)
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
	USDCurrencyRegex = regexp.MustCompile(`^[0-9]+.[0-9]{2}$`)
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
