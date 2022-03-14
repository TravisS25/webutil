package webutiltest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/pkg/errors"

	"github.com/TravisS25/webutil/webutilcfg"
)

const (
	IDParam = "{id:[0-9]+}"
)

func ValidateObjectSlice(t *testing.T, data []interface{}, mapKey string, expectedMap map[interface{}]string) {
	t.Helper()

	unexpectedVals := make([]interface{}, 0)
	nMap := make(map[interface{}]string)

	for k, v := range expectedMap {
		nMap[k] = v
	}

	for _, val := range data {
		entry, ok := val.(map[string]interface{})

		if !ok {
			t.Errorf("data entry is not JSON object")
			return
		}

		var entryVal interface{}

		if entryVal, ok = entry[mapKey]; !ok {
			t.Errorf("passed 'mapkey' parameter value not found in object")
			return
		}

		if _, ok = nMap[entryVal]; ok {
			delete(nMap, entryVal)
		} else {
			unexpectedVals = append(unexpectedVals, entry)
		}
	}

	var errStr string

	if len(nMap) != 0 {
		vals := make([]string, 0)

		for _, v := range nMap {
			vals = append(vals, v)
		}

		errStr += fmt.Sprintf("expected entries not found: %v\n\n", vals)
	}

	if len(unexpectedVals) > 0 {
		errStr += fmt.Sprintf("unexpected entries found: %v\n\n", unexpectedVals)
	}

	if errStr != "" {
		t.Errorf(errStr)
		return
	}
}

// func ValidateObjectSlice(data []interface{}, mapKey string, expectedMap map[interface{}]string) error {
// 	unexpectedVals := make([]interface{}, 0)
// 	nMap := make(map[interface{}]string)

// 	for k, v := range expectedMap {
// 		nMap[k] = v
// 	}

// 	for _, val := range data {
// 		entry, ok := val.(map[string]interface{})

// 		if !ok {
// 			return fmt.Errorf("data entry is not JSON object")
// 		}

// 		var entryVal interface{}

// 		if entryVal, ok = entry[mapKey]; !ok {
// 			return fmt.Errorf("passed 'mapkey' parameter value not found in object")
// 		}

// 		if _, ok = nMap[entryVal]; ok {
// 			delete(nMap, entryVal)
// 		} else {
// 			unexpectedVals = append(unexpectedVals, entry)
// 		}
// 	}

// 	var errStr string

// 	if len(nMap) != 0 {
// 		vals := make([]string, 0)

// 		for _, v := range nMap {
// 			vals = append(vals, v)
// 		}

// 		errStr += fmt.Sprintf("expected entries not found: %v\n\n", vals)
// 	}

// 	if len(unexpectedVals) > 0 {
// 		errStr += fmt.Sprintf("unexpected entries found: %v\n\n", unexpectedVals)
// 	}

// 	if errStr != "" {
// 		return fmt.Errorf(errStr)
// 	}

// 	return nil
// }

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
func LoginUser(url string, loginForm interface{}) (string, error) {
	res, err := loginUser(url, loginForm)

	if err != nil {
		return "", err
	}

	return res.Header.Get(webutilcfg.SetCookieHeader), nil
}

// LoginUserV takes email and password along with login url and form information
// to use to make a POST request to login url and if successful, returns user cookie
// with the value extracted
func LoginUserV(url string, loginForm interface{}) (string, error) {
	res, err := loginUser(url, loginForm)

	if err != nil {
		return "", err
	}

	if len(res.Cookies()) > 0 {
		return res.Cookies()[0].Value, nil
	}

	return "", fmt.Errorf("webutiltest: no cookie value returned")
}
