package webutiltest

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	testifymock "github.com/stretchr/testify/mock"
)

func TestValidateObjectSlice(t *testing.T) {
	var err error

	mockTestLog := &MockTestLog{}
	mockTestLog.On("Helper").Return(nil)
	mockTestLog.On("Errorf", testifymock.Anything).Return(nil)

	if err = ValidateObjectSlice(mockTestLog, nil, "invalid", nil); err == nil {
		t.Errorf("should have error")
	} else if err.Error() != "'expectedMap' or 'data' parameters can not be nil" {
		t.Errorf(
			"error should be %s; got %s\n",
			"values are not type 'map[string]interface{}' within 'data' parameter",
			err.Error(),
		)
	}

	invalidSlice := []interface{}{"1"}

	if err = ValidateObjectSlice(mockTestLog, invalidSlice, "invalid", map[interface{}]string{}); err == nil {
		t.Errorf("should have error")
	} else if err.Error() != "values are not type 'map[string]interface{}' within 'data' parameter" {
		t.Errorf(
			"error should be %s; got %s\n",
			"values are not type 'map[string]interface{}' within 'data' parameter",
			err.Error(),
		)
	}

	idKey := "id"

	mapSlice := make([]interface{}, 0)
	mapSlice = append(mapSlice, map[string]interface{}{
		idKey: "1",
	})

	if err = ValidateObjectSlice(mockTestLog, mapSlice, "invalid", map[interface{}]string{}); err == nil {
		t.Errorf("should have error")
	} else if err.Error() != "passed 'mapkey' parameter value not found in object" {
		t.Errorf(
			"error should be %s; got %s\n",
			"passed 'mapkey' parameter value not found in object",
			err.Error(),
		)
	}

	if err = ValidateObjectSlice(mockTestLog, mapSlice, idKey, map[interface{}]string{
		"1": "message",
	}); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	mapSlice = nil
	mapSlice = make([]interface{}, 0)
	mapSlice = append(
		mapSlice,
		map[string]interface{}{
			idKey: "1",
		},
		map[string]interface{}{
			idKey: "2",
		},
	)

	if err = ValidateObjectSlice(mockTestLog, mapSlice, idKey, map[interface{}]string{
		"1": "message",
	}); err == nil {
		t.Errorf("should have error")
	} else if !strings.Contains(err.Error(), "unexpected entries found") {
		t.Errorf("should have substr 'unexpected entries found'; got %s\n", err.Error())
	}

	mapSlice = nil
	mapSlice = make([]interface{}, 0)
	mapSlice = append(
		mapSlice,
		map[string]interface{}{
			idKey: "1",
		},
	)

	if err = ValidateObjectSlice(mockTestLog, mapSlice, idKey, map[interface{}]string{
		"1": "message",
		"2": "message",
	}); err == nil {
		t.Errorf("should have error")
	} else if !strings.Contains(err.Error(), "expected entries not found") {
		t.Errorf("should have substr 'expected entries not found'; got %s\n", err.Error())
	}
}

func TestLoginUser(t *testing.T) {
	var err error

	loginURL := "/login"

	router := mux.NewRouter()
	router.HandleFunc(loginURL, func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("hittng server")

		if r.Method == http.MethodGet {

		} else {
			http.SetCookie(w, &http.Cookie{
				Name:  "user",
				Value: "user",
			})
		}
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	resErr := errors.New("response error")

	mockClient := &MockHTTPClient{}
	mockClient.On("Do", testifymock.Anything).Return(&http.Response{}, resErr).Once()

	if _, err = LoginUser(mockClient, loginURL, map[string]interface{}{}); err == nil {
		t.Errorf("should have error")
	} else if !errors.Is(err, resErr) {
		t.Errorf("should have err: %s; got %s\n", resErr.Error(), err.Error())
	}

	mockClient.On("Do", testifymock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("error")),
		},
		nil,
	).Once()

	if _, err = LoginUser(mockClient, loginURL, map[string]interface{}{}); err == nil {
		t.Errorf("should have error")
	} else if !strings.Contains(err.Error(), "status code") {
		t.Errorf("should have substr err: 'status code'; got %s\n", err.Error())
	}

	mockClient.On("Do", testifymock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusOK,
		},
		nil,
	).Once()
	mockClient.On("Do", testifymock.Anything).Return(
		&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("error")),
		},
		nil,
	).Once()

	if _, err = LoginUser(mockClient, loginURL, map[string]interface{}{}); err == nil {
		t.Errorf("should have error")
	} else if !strings.Contains(err.Error(), "status code") {
		t.Errorf("should have substr err: 'status code'; got %s\n", err.Error())
	}

	if _, err = LoginUser(http.DefaultClient, ts.URL+loginURL, map[string]interface{}{}); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}
}
