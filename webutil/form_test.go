package webutil

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
