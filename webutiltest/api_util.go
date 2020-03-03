package webutiltest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/TravisS25/webutil/webutil"
)

const (
	IDParam                 = "{id:[0-9]+}"
	ResponseErrorMessage    = "apitesting: Result values: %v;\n expected results: %v\n"
	MapResponseErrorMessage = "apitesting: Key value \"%s\":\n Result values: %v;\n expected results: %v\n"
)

const (
	intMapIDResult = iota + 1
	int64MapIDResult
	intArrayIDResult
	int64ArrayIDResult
	intFilteredIDResult
	int64FilteredIDResult
	intObjectIDResult
	int64ObjectIDResult
)

// TestCase is config struct used in conjunction with
// the RunTestCases function
type TestCase struct {
	// TestName is name of given test
	TestName string
	// Method is http method used for request eg. "get", "post", "put", "delete"
	Method string
	// RequestURL is the url you want to test
	RequestURL string
	// ExpectedStatus is http response code you expect to retrieve from request
	ExpectedStatus int
	// ExpectedBody is the expected response, if any, that given response will have
	ExpectedBody string
	// RunInParallel allows for tests to be run in parallel
	RunInParallel bool
	// ContextValues is used for adding context values with request
	ContextValues map[interface{}]interface{}
	// Header is for adding custom header to request
	Header http.Header
	// Form is json information you wish to post in body of request
	Form interface{}
	//URLValues is form information you wish to post in body of request
	URLValues url.Values
	// FileUploadConfs is used to simulate a request to upload file(s)
	// to server
	FileUploadConfs []FileUploadConfig
	// Handler is the request handler that you which to test
	Handler http.Handler
	// ValidResponse allows user to take in response from api end
	// and determine if the given response is the expected one
	// ValidResponse func(bodyResponse io.Reader) (bool, error)
	ValidateResponse Response
	// PostResponse is used to validate anything a user wishes after api is
	// done executing.  This is mainly intended to be used for querying
	// against a database after POST/PUT/DELETE request to validate that
	// proper things were written to the database.  Could also be used
	// for clean up
	PostResponseValidation func() error
}

type intID struct {
	ID int `json:"id"`
}

type int64ID struct {
	ID int64 `json:"id,string"`
}

type filteredIntID struct {
	Data  []intID `json:"data"`
	Count int     `json:"count"`
}

type filteredInt64ID struct {
	Data  []int64ID `json:"data"`
	Count int       `json:"count"`
}

// FileConfig is used in conjunction with FileUploadConfig to
// set configuration settings to upload file to server
type FileConfig struct {
	// FilePath is file path to file to upload
	FilePath string `json:"filePath" mapstructure:"file_path"`

	// Params is extra parameters to add to file upload request
	Params map[string]string `json:"params" mapstructure:"params"`
}

// FileUploadConfig is config struct used to set up configuration
// to upload file(s) to server
type FileUploadConfig struct {
	// ParamName is name of parameter that is used to get multipart form
	// uploaded from client
	ParamName string       `json:"paramName" mapstructure:"param_name"`
	FileConfs []FileConfig `json:"fileConfs" mapstructure:"file_confs"`
}

type Response struct {
	ExpectedResult       interface{}
	ValidateResponseFunc func(bodyResponse io.Reader, expectedResult interface{}) error
}

func NewRequestWithForm(method, url string, form interface{}) (*http.Request, error) {
	if form != nil {
		var buffer bytes.Buffer
		encoder := json.NewEncoder(&buffer)
		encoder.Encode(&form)
		return http.NewRequest(method, url, &buffer)
	}

	return http.NewRequest(method, url, nil)
}

func RunTestCases(t *testing.T, deferFunc func() error, testCases []TestCase) {
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.TestName, func(v *testing.T) {
			if tc.RunInParallel {
				t.Parallel()
			}
			panicked := true
			defer func() {
				if deferFunc != nil {
					if panicked {
						err := deferFunc()

						if err != nil {
							fmt.Printf(err.Error())
						}
					}
				}
			}()
			var req *http.Request
			var err error

			// If Form and File options are nil, init req without added parameters
			// Else check whether Form or file option is selected.
			// Right now, File option will overide Form option
			if tc.Form == nil && tc.URLValues == nil && tc.FileUploadConfs == nil {
				req, err = http.NewRequest(tc.Method, tc.RequestURL, nil)
			} else {
				if tc.FileUploadConfs != nil {
					req, err = NewFileUploadRequest(tc.FileUploadConfs, http.MethodPost, "/url")

					if err != nil {
						v.Fatal(err)
					}
				} else if tc.URLValues != nil {
					req, err = http.NewRequest(tc.Method, tc.RequestURL, strings.NewReader(tc.URLValues.Encode()))

					if err != nil {
						v.Fatal(err)
					}
				} else {
					var buffer bytes.Buffer
					encoder := json.NewEncoder(&buffer)
					err = encoder.Encode(&tc.Form)

					if err != nil {
						v.Fatal(err)
					}

					req, err = http.NewRequest(tc.Method, tc.RequestURL, &buffer)

					if err != nil {
						v.Fatal(err)
					}
				}
			}

			req.Header = tc.Header

			// If ContextValues is not nil, apply given context values to req
			if tc.ContextValues != nil {
				ctx := req.Context()

				for key, value := range tc.ContextValues {
					ctx = context.WithValue(ctx, key, value)
				}

				req = req.WithContext(ctx)
			}

			// Init recorder that will be written to based on the status
			// we get from created request
			rr := httptest.NewRecorder()
			tc.Handler.ServeHTTP(rr, req)

			// If status is not what was expected, print error
			if status := rr.Code; status != tc.ExpectedStatus {
				v.Errorf("got status %d; want %d\n", status, tc.ExpectedStatus)
				v.Errorf("body response: %s\n", rr.Body.String())
			}

			// If ExpectedBody option was given and does not equal what was
			// returned, print error
			if tc.ExpectedBody != "" {
				if tc.ExpectedBody != rr.Body.String() {
					v.Errorf("got body %s; want %s\n", rr.Body.String(), tc.ExpectedBody)

				}
			}

			if tc.ValidateResponse.ValidateResponseFunc != nil {
				err = tc.ValidateResponse.ValidateResponseFunc(
					rr.Body,
					tc.ValidateResponse.ExpectedResult,
				)

				if err != nil {
					v.Errorf(err.Error() + "\n")
				}
			}

			if tc.PostResponseValidation != nil {
				if err = tc.PostResponseValidation(); err != nil {
					v.Errorf(err.Error() + "\n")
				}
			}

			panicked = false
		})
	}
}

func validateIDResponse(bodyResponse io.Reader, result int, expectedResults interface{}) error {
	foundResult := false

	switch result {
	case intArrayIDResult:
		expectedIDs, ok := expectedResults.([]int)

		if !ok {
			return errors.New("apitesting: Expected result should be []int")
		}

		var responseResults []intID
		err := SetJSONFromResponse(bodyResponse, &responseResults)

		if err != nil {
			return err
		}

		if len(responseResults) != len(expectedIDs) {
			errorMessage := fmt.Sprintf(
				ResponseErrorMessage,
				responseResults,
				expectedIDs,
			)
			return errors.New(errorMessage)
		}

		for _, m := range expectedIDs {
			for _, v := range responseResults {
				if m == v.ID {
					foundResult = true
					break
				}
			}

			if foundResult == false {
				errorMessage := fmt.Sprintf(
					ResponseErrorMessage,
					responseResults,
					expectedIDs,
				)
				return errors.New(errorMessage)
			}

			foundResult = false
		}
		break

	case int64ArrayIDResult:
		expectedIDs, ok := expectedResults.([]int64)

		if !ok {
			return errors.New("apitesting: Expected result should be []int64")
		}

		var responseResults []int64ID
		err := SetJSONFromResponse(bodyResponse, &responseResults)

		if err != nil {
			return err
		}

		if len(responseResults) != len(expectedIDs) {
			errorMessage := fmt.Sprintf(
				ResponseErrorMessage,
				responseResults,
				expectedIDs,
			)
			return errors.New(errorMessage)
		}

		for _, m := range expectedIDs {
			for _, v := range responseResults {
				if m == v.ID {
					foundResult = true
					break
				}
			}

			if foundResult == false {
				errorMessage := fmt.Sprintf(
					ResponseErrorMessage,
					responseResults,
					expectedIDs,
				)
				return errors.New(errorMessage)
			}

			foundResult = false
		}
		break

	case intFilteredIDResult:
		expectedIDs, ok := expectedResults.([]int)

		if !ok {
			return errors.New("apitesting: Expected result should be []int")
		}

		var responseResults filteredIntID
		err := SetJSONFromResponse(bodyResponse, &responseResults)

		if err != nil {
			return err
		}

		if len(responseResults.Data) != len(expectedIDs) {
			errorMessage := fmt.Sprintf(
				ResponseErrorMessage,
				responseResults.Data,
				expectedIDs,
			)
			return errors.New(errorMessage)
		}

		for _, m := range expectedIDs {
			for _, v := range responseResults.Data {
				if m == v.ID {
					foundResult = true
					break
				}
			}

			if foundResult == false {
				errorMessage := fmt.Sprintf(
					ResponseErrorMessage,
					responseResults.Data,
					expectedIDs,
				)
				return errors.New(errorMessage)
			}

			foundResult = false
		}
		break
	case int64FilteredIDResult:
		expectedIDs, ok := expectedResults.([]int64)

		if !ok {
			return errors.New("apitesting: Expected result should be []int64")
		}

		var responseResults filteredInt64ID
		err := SetJSONFromResponse(bodyResponse, &responseResults)

		if err != nil {
			return err
		}

		if len(responseResults.Data) != len(expectedIDs) {
			errorMessage := fmt.Sprintf(
				ResponseErrorMessage,
				responseResults.Data,
				expectedIDs,
			)
			return errors.New(errorMessage)
		}

		for _, m := range expectedIDs {
			for _, v := range responseResults.Data {
				if m == v.ID {
					foundResult = true
					break
				}
			}

			if foundResult == false {
				errorMessage := fmt.Sprintf(
					ResponseErrorMessage,
					responseResults.Data,
					expectedIDs,
				)
				return errors.New(errorMessage)
			}

			foundResult = false
		}
		break
	case intObjectIDResult:
		expectedID, ok := expectedResults.(int)

		if !ok {
			return errors.New("apitesting: Expected result should be int")
		}

		var responseResults intID
		err := SetJSONFromResponse(bodyResponse, &responseResults)

		if err != nil {
			return err
		}

		if responseResults.ID != expectedID {
			errorMessage := fmt.Sprintf(
				ResponseErrorMessage,
				responseResults.ID,
				expectedID,
			)
			return errors.New(errorMessage)
		}
	case int64ObjectIDResult:
		expectedID, ok := expectedResults.(int64)

		if !ok {
			return errors.New("apitesting: Expected result should be int64")
		}

		var responseResults int64ID
		err := SetJSONFromResponse(bodyResponse, &responseResults)

		if err != nil {
			return err
		}

		if responseResults.ID != expectedID {
			errorMessage := fmt.Sprintf(
				ResponseErrorMessage,
				responseResults.ID,
				expectedID,
			)
			return errors.New(errorMessage)
		}
	case intMapIDResult:
		expectedMap, ok := expectedResults.(map[string]interface{})

		if !ok {
			return errors.New("apitesting: Expected result should be map[string]interface{}")
		}

		var responseResults map[string]interface{}
		err := SetJSONFromResponse(bodyResponse, &responseResults)

		if err != nil {
			return err
		}

		if len(responseResults) != len(expectedMap) {
			errorMessage := fmt.Sprintf(
				ResponseErrorMessage,
				responseResults,
				expectedMap,
			)
			return errors.New(errorMessage)
		}

		// Loop through given expected map of values and check whether the key
		// values are within the body response key values
		//
		// If key exists, determine through reflection if value is struct or
		// slice and compare ids to determine if expected map value
		// equals value of body response map
		//
		// If key does not exist, return err
		for k := range expectedMap {
			if responseVal, ok := responseResults[k]; ok {
				// Get json bytes from body response
				buf := bytes.Buffer{}
				buf.ReadFrom(bodyResponse)

				// Determine kind for interface{} value so we can
				// properly convert to typed json
				switch reflect.TypeOf(responseVal).Kind() {
				// If interface{} value is struct, then convert convertedResults
				// and expectedMap to typed json (Int64ID) to compare id
				case reflect.Struct:
					var expectedIntID intID
					var responseIntID intID

					expectedIDBytes, err := json.Marshal(expectedMap[k])

					if err != nil {
						message := fmt.Sprintf("apitesting: %s", err.Error())
						return errors.New(message)
					}

					responseIDBytes, err := json.Marshal(responseVal)

					if err != nil {
						message := fmt.Sprintf("apitesting: %s", err.Error())
						return errors.New(message)
					}

					err = json.Unmarshal(expectedIDBytes, &expectedIntID)

					if err != nil {
						message := fmt.Sprintf("apitesting: %s", err.Error())
						return errors.New(message)
					}

					err = json.Unmarshal(responseIDBytes, &responseIntID)

					if err != nil {
						message := fmt.Sprintf("apitesting: %s", err.Error())
						return errors.New(message)
					}

					if expectedIntID.ID != responseIntID.ID {
						errorMessage := fmt.Sprintf(
							ResponseErrorMessage,
							responseResults,
							expectedMap,
						)
						return errors.New(errorMessage)
					}

				// If interface{} value is slice, then convert body response
				// and expectedMap to typed json (int64MapSliceID) to then
				// loop through and compare ids
				case reflect.Slice:
					var expectedIntIDs []intID
					var responseIntIDs []intID

					expectedIDsBytes, err := json.Marshal(expectedMap[k])

					if err != nil {
						message := fmt.Sprintf("apitesting: %s", err.Error())
						return errors.New(message)
					}

					responseIDsBytes, err := json.Marshal(responseVal)

					if err != nil {
						message := fmt.Sprintf("apitesting: %s", err.Error())
						return errors.New(message)
					}

					err = json.Unmarshal(expectedIDsBytes, &expectedIntIDs)

					if err != nil {
						message := fmt.Sprintf("apitesting: %s", err.Error())
						return errors.New(message)
					}

					err = json.Unmarshal(responseIDsBytes, &responseIntIDs)

					if err != nil {
						message := fmt.Sprintf("apitesting: %s", err.Error())
						return errors.New(message)
					}

					for _, v := range expectedIntIDs {
						containsID := false

						for _, t := range responseIntIDs {
							if t.ID == v.ID {
								containsID = true
								break
							}
						}

						if !containsID {
							message := fmt.Sprintf(
								"apitesting: Slice response does not contain %d", v.ID,
							)
							return errors.New(message)
						}
					}

				// Id interface{} valie is neither struct or slice, then return err
				default:
					return errors.New("apitesting: not valid type")
				}
			} else {
				message := fmt.Sprintf("apitesting: key %s not in results from body", k)
				return errors.New(message)
			}
		}
	case int64MapIDResult:
		expectedMap, ok := expectedResults.(map[string]interface{})

		if !ok {
			return errors.New("apitesting: Expected result should be map[string]interface{}")
		}

		var responseResults map[string]interface{}
		err := SetJSONFromResponse(bodyResponse, &responseResults)

		if err != nil {
			return err
		}

		if len(responseResults) != len(expectedMap) {
			errorMessage := fmt.Sprintf(
				"Map result length %d; expected map length: %d",
				len(responseResults),
				len(expectedMap),
			)
			return errors.New(errorMessage)
		}

		// Loop through given expected map of values and check whether the key
		// values are within the body response key values
		//
		// If key exists, determine through reflection if value is struct or
		// slice and compare ids to determine if expected map value
		// equals value of body response map
		//
		// If key does not exist, return err
		for k := range expectedMap {
			if responseVal, ok := responseResults[k]; ok {
				// Determine kind for interface{} value so we can
				// properly convert to typed json
				if responseVal != nil {
					// Determine kind for interface{} value so we can
					// properly convert to typed json
					switch reflect.TypeOf(responseVal).Kind() {
					// If interface{} value is map, then convert convertedResults
					// and expectedMap to typed json (int64ID) to compare id
					case reflect.Map:
						var responseInt64ID int64ID

						expectedInt64, ok := expectedMap[k].(int64)

						if !ok {
							message := fmt.Sprintf(
								`apitesting: key value "%s" for "ExpectedResult" should be int64`,
								k,
							)
							return errors.New(message)
						}

						responseIDBytes, err := json.Marshal(responseVal)

						if err != nil {
							message := fmt.Sprintf("apitesting: %s", err.Error())
							return errors.New(message)
						}

						err = json.Unmarshal(responseIDBytes, &responseInt64ID)

						if err != nil {
							message := fmt.Sprintf("apitesting: %s", err.Error())
							return errors.New(message)
						}

						if expectedInt64 != responseInt64ID.ID {
							errorMessage := fmt.Sprintf(
								MapResponseErrorMessage,
								k,
								responseInt64ID.ID,
								expectedInt64,
							)
							return errors.New(errorMessage)
						}

					// If interface{} value is slice, then convert body response
					// and expectedMap to typed json (int64MapSliceID) to then
					// loop through and compare ids
					case reflect.Slice:
						var responseInt64IDs []int64ID

						expectedInt64Slice, ok := expectedMap[k].([]int64)

						if !ok {
							message := fmt.Sprintf(`apitesting: key value "%s" for "ExpectedResult" should be []int64`, k)
							return errors.New(message)
						}

						responseIDsBytes, err := json.Marshal(responseVal)

						if err != nil {
							message := fmt.Sprintf("apitesting: %s", err.Error())
							return errors.New(message)
						}

						err = json.Unmarshal(responseIDsBytes, &responseInt64IDs)

						if err != nil {
							message := fmt.Sprintf("apitesting: %s", err.Error())
							return errors.New(message)
						}

						for _, v := range expectedInt64Slice {
							containsID := false

							for _, t := range responseInt64IDs {
								if t.ID == v {
									containsID = true
									break
								}
							}

							if !containsID {
								message := fmt.Sprintf(
									"apitesting: Slice response does not contain %d", v,
								)
								return errors.New(message)
							}
						}

					// Id interface{} valie is neither struct or slice, then return err
					default:
						return errors.New("apitesting: not valid type")
					}
				} else {
					if expectedMap[k] != nil {
						errorMessage := fmt.Sprintf(
							MapResponseErrorMessage,
							k,
							nil,
							expectedMap[k],
						)
						return errors.New(errorMessage)
					}
				}
			} else {
				message := fmt.Sprintf(`apitesting: key value "%s" not in results from body`, k)
				return errors.New(message)
			}
		}

	default:
		return errors.New("apitesting: Invalid result type passed")
	}

	return nil
}

func ValidateFilteredIntArrayResponse(bodyResponse io.Reader, expectedResult interface{}) error {
	return validateIDResponse(bodyResponse, intFilteredIDResult, expectedResult)
}

func ValidateFilteredInt64ArrayResponse(bodyResponse io.Reader, expectedResult interface{}) error {
	return validateIDResponse(bodyResponse, int64FilteredIDResult, expectedResult)
}

func ValidateIntArrayResponse(bodyResponse io.Reader, expectedResult interface{}) error {
	return validateIDResponse(bodyResponse, intArrayIDResult, expectedResult)
}

func ValidateInt64ArrayResponse(bodyResponse io.Reader, expectedResult interface{}) error {
	return validateIDResponse(bodyResponse, int64ArrayIDResult, expectedResult)
}

func ValidateIntMapResponse(bodyResponse io.Reader, expectedResult interface{}) error {
	return validateIDResponse(bodyResponse, intMapIDResult, expectedResult)
}

func ValidateInt64MapResponse(bodyResponse io.Reader, expectedResult interface{}) error {
	return validateIDResponse(bodyResponse, int64MapIDResult, expectedResult)
}

func ValidateIntObjectResponse(bodyResponse io.Reader, expectedResult interface{}) error {
	return validateIDResponse(bodyResponse, intObjectIDResult, expectedResult)
}

func ValidateInt64ObjectResponse(bodyResponse io.Reader, expectedResult interface{}) error {
	return validateIDResponse(bodyResponse, int64ObjectIDResult, expectedResult)
}

func ValidateStringResponse(bodyResponse io.Reader, expectedResult interface{}) error {
	response, err := ioutil.ReadAll(bodyResponse)

	if err != nil {
		return err
	}

	if result, ok := expectedResult.(string); ok {
		if string(response) != result {
			return fmt.Errorf("Response and expected strings did not match")
		}

		return nil
	}

	return fmt.Errorf("Expected result must be string")
}

// SetJSONFromResponse takes io.Reader which will generally be a response from
// api endpoint and applies the json representation to the passed interface
func SetJSONFromResponse(bodyResponse io.Reader, item interface{}) error {
	response, err := ioutil.ReadAll(bodyResponse)

	if err != nil {
		return err
	}

	err = json.Unmarshal(response, &item)

	if err != nil {
		return err
	}

	return nil
}

// LoginUser takes email and password along with login url and form information
// to use to make a POST request to login url and if successful, return user cookie
func LoginUser(url string, loginForm interface{}) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return "", err
	}

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		buf.ReadFrom(res.Body)
		errorMessage := fmt.Sprintf("status code: %d\n  response: %s\n", res.StatusCode, buf.String())
		return "", errors.New(errorMessage)
	}

	token := res.Header.Get(webutil.TokenHeader)
	csrf := res.Header.Get(webutil.SetCookieHeader)
	buffer, _ := webutil.GetJSONBuffer(loginForm)
	req, err = http.NewRequest(http.MethodPost, url, &buffer)

	if err != nil {
		return "", err
	}

	req.Header.Set(webutil.TokenHeader, token)
	req.Header.Set(webutil.CookieHeader, csrf)
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

	return res.Header.Get(webutil.SetCookieHeader), nil
}

func NewFileUploadRequest(confs []FileUploadConfig, method, url string) (*http.Request, error) {
	body, contentType, err := FileBody(confs)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", contentType)
	return req, err
}

func FileBody(confs []FileUploadConfig) (io.Reader, string, error) {
	var typ string
	var err error
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, v := range confs {
		for _, t := range v.FileConfs {
			file, err := os.Open(t.FilePath)
			if err != nil {
				return nil, "", err
			}
			defer file.Close()

			part, err := writer.CreateFormFile(v.ParamName, filepath.Base(t.FilePath))
			if err != nil {
				return nil, "", err
			}
			_, err = io.Copy(part, file)

			if err != nil {
				return nil, "", err
			}

			if t.Params != nil {
				for key, val := range t.Params {
					err = writer.WriteField(key, val)

					if err != nil {
						return nil, "", err
					}
				}
			}

			typ = writer.FormDataContentType()
		}
	}

	err = writer.Close()

	if err != nil {
		return nil, "", err
	}

	return body, typ, nil
}

func CheckResponse(method, url string, expectedStatus int, header http.Header, form interface{}) (*http.Response, error) {
	client := &http.Client{}
	buffer := &bytes.Buffer{}
	req, _ := NewRequestWithForm(method, url, form)
	req.Header = header
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != expectedStatus {
		buffer.ReadFrom(res.Body)
		message := fmt.Sprintf("got status %d; want %d\nresponse: %s", res.StatusCode, http.StatusOK, buffer.String())
		return nil, errors.New(message)
	}

	return res, nil
}
