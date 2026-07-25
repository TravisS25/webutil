package webutil

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
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
	//CurrencyRegex *regexp.Regexp `json:"-"`

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

func (i *Int64) Scan(value any) error {
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

func (i *Int64) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}

	return int64(*i), nil
}

func (i *Int64) UnmarshalText(text []byte) error {
	val, err := strconv.ParseInt(string(text), INT_BASE, INT_BIT_SIZE)
	if err != nil {
		return fmt.Errorf("failed to parse Int64: %w", err)
	}
	*i = Int64(val)
	return nil
}

// //////////////////////////////////////////////////////////////////
// //----------------------- FUNCTIONS -------------------------
// //////////////////////////////////////////////////////////////////

// CheckBodyAndDecode takes request and decodes the json body from the request
// to the passed struct
//
// The excludeMethods parameter allows user to pass certain http methods
// that skip decoding the request body if nil else will throw ErrBodyRequired error
func CheckBodyAndDecode(req *http.Request, form any, excludeMethods ...string) error {
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
			return fmt.Errorf("%w: %s", err, ErrInvalidJSON)
		}
	} else {
		if !canSkip {
			return ErrBodyRequired
		}
	}

	return nil
}
