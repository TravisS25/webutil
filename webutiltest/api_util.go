package webutiltest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/TravisS25/webutil/webutilcfg"
)

const (
	IDParam = "{id:[0-9]+}"
)

// ValidateObjectSlice takes in a slice of maps, with a mapkey to then test against the expectedMap keys
// which should be the value expected and the value of expectedMap should be unique name of value
func ValidateObjectSlice(t TestLog, data []interface{}, mapKey string, expectedMap map[interface{}]string) error {
	t.Helper()

	var errStr string

	if data == nil || expectedMap == nil {
		errStr = "'expectedMap' or 'data' parameters can not be nil"
		t.Errorf(errStr)
		return errors.New(errStr)
	}

	unexpectedVals := make([]interface{}, 0)
	newExpectedMap := make(map[interface{}]string)

	for k, v := range expectedMap {
		newExpectedMap[k] = v
	}

	for _, val := range data {
		entry, ok := val.(map[string]interface{})

		if !ok {
			errStr = "values are not type 'map[string]interface{}' within 'data' parameter"
			t.Errorf(errStr)
			return errors.New(errStr)
		}

		var keyVal interface{}

		if keyVal, ok = entry[mapKey]; !ok {
			errStr = "passed 'mapkey' parameter value not found in object"
			t.Errorf(errStr)
			return errors.New(errStr)
		}

		if _, ok = newExpectedMap[keyVal]; ok {
			delete(newExpectedMap, keyVal)
		} else {
			unexpectedVals = append(unexpectedVals, val)
		}
	}

	if len(newExpectedMap) != 0 {
		vals := make([]string, 0)

		for _, v := range newExpectedMap {
			vals = append(vals, v)
		}

		errStr += fmt.Sprintf("expected entries not found: %v\n\n", vals)
	}

	if len(unexpectedVals) > 0 {
		errStr += fmt.Sprintf("unexpected entries found: %v\n\n", unexpectedVals)
	}

	if errStr != "" {
		t.Errorf(errStr)
		return errors.New(errStr)
	}

	return nil
}

func loginUser(url string, loginForm interface{}) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		buf.ReadFrom(res.Body)
		errorMessage := fmt.Sprintf("status code: %d\n  response: %s\n", res.StatusCode, buf.String())
		return nil, errors.New(errorMessage)
	}

	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(&loginForm); err != nil {
		return &http.Response{}, errors.WithStack(err)
	}

	token := res.Header.Get(webutilcfg.TokenHeader)
	csrf := res.Header.Get(webutilcfg.SetCookieHeader)
	req, err = http.NewRequest(http.MethodPost, url, &buffer)

	if err != nil {
		return nil, err
	}

	req.Header.Set(webutilcfg.TokenHeader, token)
	req.Header.Set(webutilcfg.CookieHeader, csrf)
	res, err = client.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		buf.ReadFrom(res.Body)
		errorMessage := fmt.Sprintf("status code: %d\n  response: %s\n", res.StatusCode, buf.String())
		return nil, errors.New(errorMessage)
	}

	return res, nil
}

// LoginUser takes email and password along with login url and form information
// to use to make a POST request to login url and if successful, returns user cookie
func LoginUser(client HTTPClient, url string, loginForm interface{}) (string, error) {
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
		return "", errors.New(
			fmt.Sprintf("status code: %d\n  response: %s\n", res.StatusCode, buf.String()),
		)
	}

	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(&loginForm); err != nil {
		return "", errors.WithStack(err)
	}

	token := res.Header.Get(webutilcfg.TokenHeader)
	csrf := res.Header.Get(webutilcfg.SetCookieHeader)
	req, err = http.NewRequest(http.MethodPost, url, &buffer)

	if err != nil {
		return "", err
	}

	req.Header.Set(webutilcfg.TokenHeader, token)
	req.Header.Set(webutilcfg.CookieHeader, csrf)
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
