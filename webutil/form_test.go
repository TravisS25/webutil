package webutil

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pkgerrors "github.com/pkg/errors"
)

func TestValidateRequiredRuleUnitTest(t *testing.T) {
	var err error

	rule := &validateRequiredRule{err: errors.New(REQUIRED_TXT)}

	if err = rule.Validate(nil); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != REQUIRED_TXT {
			t.Errorf("should have returned required error\n")
		}
	}

	if err = rule.Validate(""); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != REQUIRED_TXT {
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
		if err.Error() != REQUIRED_TXT {
			t.Errorf("should have required errors\n")
		}
	}

	if err = rule.Validate([]string{"hi", "there"}); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	hi := "hi"

	if err = rule.Validate([]any{&hi, "there", 1}); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestValidateDateRuleUnitTest(t *testing.T) {
	var err error
	var rule *validateDateRule

	futureDateStr := time.Now().AddDate(0, 0, 1).Format(FORM_DATE_TIME_LAYOUT)
	pastDateStr := time.Now().AddDate(0, 0, -1).Format(FORM_DATE_TIME_LAYOUT)

	rule = &validateDateRule{
		timezone: "invalid",
	}

	if err = rule.Validate(nil); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = rule.Validate(""); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if err = rule.Validate("invalid"); err == nil {
		t.Errorf("should have error\n")
	}

	// --------------------------------------------------------------------

	rule = &validateDateRule{
		canBeFuture: true,
		canBePast:   true,
	}

	if err = rule.Validate(pastDateStr); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	// --------------------------------------------------------------------

	rule = &validateDateRule{
		canBeFuture: true,
		canBePast:   false,
	}

	if err = rule.Validate(pastDateStr); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != INVALID_PAST_DATE_TXT {
			t.Errorf("should have ErrInvalidPastDateValidator error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	// --------------------------------------------------------------------

	rule = &validateDateRule{
		canBeFuture: false,
		canBePast:   true,
	}

	if err = rule.Validate(futureDateStr); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != INVALID_FUTURE_DATE_TXT {
			t.Errorf("should have ErrInvalidFutureDateValidator error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	// --------------------------------------------------------------------

	rule = &validateDateRule{
		canBeFuture: false,
		canBePast:   false,
	}

	if err = rule.Validate(futureDateStr); err == nil {
		t.Errorf("should have error\n")
	} else {
		if pkgerrors.Cause(err).Error() != errFutureAndPastDateInternal.Error() {
			t.Errorf("should have errFutureAndPastDateInternal error\n")
			t.Errorf("err: %s\n", err.Error())
		}
	}

	// --------------------------------------------------------------------

	thenUTCTime := time.Now().UTC().Add(-time.Hour * 1).Format(FORM_DATE_TIME_LAYOUT)

	rule = &validateDateRule{
		compareTime: true,
		canBeFuture: false,
		canBePast:   true,
	}

	if err = rule.Validate(thenUTCTime); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// -------------------------------------------------------------------------

	rule = &validateDateRule{
		timezone:    "America/New_York",
		compareTime: true,
		canBeFuture: false,
		canBePast:   true,
	}

	if err = rule.Validate(thenUTCTime); err == nil {
		t.Errorf("should have error")
	} else if err.Error() != INVALID_FUTURE_DATE_TXT {
		t.Errorf("should have future date error; got %s\n", err.Error())
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

// func TestFormCurrency(t *testing.T) {
// 	type foo struct {
// 		Price FormCurrency `json:"price"`
// 	}

// 	var err error
// 	var form foo

// 	if err = json.Unmarshal([]byte(`{"price": 103.67}`), &form); err != nil {
// 		t.Fatalf(err.Error())
// 	}

// 	if err = form.Price.Validate(); err != nil {
// 		t.Fatalf(err.Error())
// 	}

// 	form.Price.CurrencyRegex = USDCurrencyRegex

// 	if err = form.Price.Validate(); err != nil {
// 		t.Fatalf(err.Error())
// 	}

// 	var tmp decimal.Decimal

// 	tmp = decimal.NewFromFloat(100)
// 	form.Price.Max = &tmp

// 	if err = form.Price.Validate(); err == nil {
// 		t.Fatalf("should have error\n")
// 	} else {
// 		if err.Error() != "Can't be greater than 100" {
// 			t.Fatalf(`err should be: "Can't be greater than 100"; got "%s"`, err.Error())
// 		}
// 	}

// 	form.Price.MaxError = fmt.Errorf("foo")

// 	if err = form.Price.Validate(); err == nil {
// 		t.Fatalf("should have error\n")
// 	} else {
// 		if err.Error() != "foo" {
// 			t.Fatalf(`err should be: "foo"; got "%s"`, err.Error())
// 		}
// 	}

// 	form.Price.Decimal = decimal.NewFromFloat(-1)
// 	tmp = decimal.NewFromFloat(0)
// 	form.Price.Min = &tmp

// 	if err = form.Price.Validate(); err == nil {
// 		t.Fatalf("should have error\n")
// 	} else {
// 		if err.Error() != "Can't be less than 0" {
// 			t.Fatalf(`err should be: "Can't be less than 0"; got "%s"`, err.Error())
// 		}
// 	}

// 	form.Price.MinError = fmt.Errorf("bar")

// 	if err = form.Price.Validate(); err == nil {
// 		t.Fatalf("should have error\n")
// 	} else {
// 		if err.Error() != "bar" {
// 			t.Fatalf(`err should be: "bar"; got "%s"`, err.Error())
// 		}
// 	}

// 	form.Price.Decimal = decimal.NewFromFloat(100.345)

// 	if err = form.Price.Validate(); err == nil {
// 		t.Fatalf("should have error\n")
// 	} else {
// 		if err.Error() != INVALID_FORMAT_TXT {
// 			t.Fatalf(`err should be: "%s"; got "%s"`, INVALID_FORMAT_TXT, err.Error())
// 		}
// 	}
// }
