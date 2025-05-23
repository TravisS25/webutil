package webutiltest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sanity-io/litter"
	"github.com/stretchr/objx"

	"github.com/TravisS25/webutil/webutil"
)

// ValidateObjectSlice takes in a slice of maps, with a mapkey to then test against the expectedMap keys
// which should be the value expected and the value of expectedMap should be unique name of value
func ValidateObjectSlice(t TestLog, data []any, mapKey string, expectedMap map[any]string) error {
	t.Helper()

	var errStr string

	if data == nil || expectedMap == nil {
		errStr = "'expectedMap' or 'data' parameters can not be nil"
		t.Errorf(errStr)
		return errors.New(errStr)
	}

	unexpectedVals := make([]any, 0)
	newExpectedMap := make(map[any]string)

	for k, v := range expectedMap {
		newExpectedMap[k] = v
	}

	for _, val := range data {
		entry, ok := val.(map[string]any)

		if !ok {
			errStr = "values are not type 'map[string]any' within 'data' parameter"
			t.Errorf(errStr)
			return errors.New(errStr)
		}

		objEntry := objx.New(entry)

		if !objEntry.Has(mapKey) {
			errStr = "passed 'mapkey' parameter value not found in object"
			t.Errorf(errStr)
			return errors.New(errStr)
		}

		if _, ok = newExpectedMap[objEntry.Get(mapKey).Data()]; ok {
			delete(newExpectedMap, objEntry.Get(mapKey).Data())
		} else {
			unexpectedVals = append(unexpectedVals, val)
		}
	}

	if len(newExpectedMap) != 0 {
		vals := make([]string, 0)

		for _, v := range newExpectedMap {
			vals = append(vals, v)
		}

		errStr += fmt.Sprintf("\n\nexpected entries not found: %v\n\n", vals)
	}

	if len(unexpectedVals) > 0 {
		errStr += fmt.Sprintf("unexpected entries found: %s\n\n", litter.Sdump(unexpectedVals))
	}

	if errStr != "" {
		t.Errorf(errStr)
		return errors.New(errStr)
	}

	return nil
}

// LoginUser takes email and password along with login url and form information
// to use to make a POST request to login url and if successful, returns user cookie
func LoginUser(client HTTPClient, url string, loginForm any) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return "", errors.WithStack(err)
	}

	res, err := client.Do(req)

	if err != nil {
		return "", errors.WithStack(err)
	}

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		buf.ReadFrom(res.Body)
		return "", fmt.Errorf("status code: %d\n  response: %s", res.StatusCode, buf.String())
	}

	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(&loginForm); err != nil {
		return "", errors.WithStack(err)
	}

	token := res.Header.Get(webutil.TOKEN_HEADER)
	csrf := res.Header.Get(webutil.SET_COOKIE_HEADER)
	req, err = http.NewRequest(http.MethodPost, url, &buffer)

	if err != nil {
		return "", err
	}

	req.Header.Set(webutil.TOKEN_HEADER, token)
	req.Header.Set(webutil.COOKIE_HEADER, csrf)
	res, err = client.Do(req)

	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		buf.ReadFrom(res.Body)
		errorMessage := fmt.Sprintf("status code: %d\n  response: %s\n", res.StatusCode, buf.String())
		return "", errors.New(errorMessage)
	}

	if len(res.Cookies()) > 0 {
		return res.Cookies()[0].Value, nil
	}

	return "", fmt.Errorf("webutiltest: no cookie value returned")
}
