package webutiltest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/nqd/flat"
	"github.com/pkg/errors"

	validation "github.com/go-ozzo/ozzo-validation"
)

type MultiPartFormParams struct {
	Key  string
	Form any
}

type MultiPartFileParams struct {
	Key       string
	Filenames []string
}

// MultiPartFormRequestBuilder builds out request to include files along with optional form body
// formKey and form parameters can be left empty/nil if body request is not needed
func MultipartFormRequestBuilder(
	t TestLog,
	method,
	url string,
	formParams []MultiPartFormParams,
	fileParams []MultiPartFileParams,
	ctxVals map[any]any,
) *http.Request {
	t.Helper()

	// Create a buffer to hold the request body
	var buf bytes.Buffer

	// Create a multipart writer to construct the multipart/form-data body
	writer := multipart.NewWriter(&buf)

	for idx, formParam := range formParams {
		if formParam.Key == "" {
			t.Errorf("Must set field %q at index %d for formParams", "Key", idx)
		}

		if formParam.Form != nil {
			formBytes, err := json.Marshal(formParam.Form)
			if err != nil {
				t.Errorf("Error marshaling form: %s", err)
			}

			if err = writer.WriteField(formParam.Key, string(formBytes)); err != nil {
				t.Errorf("Error writing field %q to writer: %s", formParam.Key, err)
			}
		}
	}

	for idx, fileParam := range fileParams {
		if fileParam.Key == "" {
			t.Errorf("Must set field %q at index %q for fileParams", "Key", idx)
		}

		for _, filename := range fileParam.Filenames {
			// Open the file to be uploaded
			file, err := os.Open(filename)
			if err != nil {
				t.Errorf("Error opening file: %s", err)
			}
			defer file.Close()

			// Create the file field in the multipart form (key = "file", value = file content)
			part, err := writer.CreateFormFile(fileParam.Key, filepath.Base(file.Name()))
			if err != nil {
				t.Errorf("Error creating form file: %s", err)
			}

			// Copy the contents of the file into the form data
			_, err = io.Copy(part, file)
			if err != nil {
				t.Errorf("Error copying file data: %s", err)
			}
		}
	}

	// Close the writer to finalize the multipart form
	writer.Close()

	// Create a POST request with the correct Content-Type
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		t.Errorf("Error creating request: %s", err)
	}

	// Set the Content-Type to the value from the writer (includes boundary)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if ctxVals != nil {
		ctx := req.Context()

		for k, v := range ctxVals {
			ctx = context.WithValue(ctx, k, v)
		}

		req = req.WithContext(ctx)
	}

	return req
}

// FormRequestBuilder takes in various params and builds a request object suited for form validation
func FormRequestBuilder(t TestLog, method, url string, form any, ctxVals map[any]any) *http.Request {
	t.Helper()

	formBytes, err := json.Marshal(form)
	if err != nil {
		t.Errorf("%v", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(formBytes))
	if err != nil {
		t.Errorf("%v", err)
	}

	if ctxVals != nil {
		ctx := req.Context()

		for k, v := range ctxVals {
			ctx = context.WithValue(ctx, k, v)
		}

		req = req.WithContext(ctx)
	}

	return req
}

// ValidateFormError determines whether the error passed is a form error
// and whether it returns the expected errors with the validatorMap parameter
//
// If no errors are expected, validatorMap parameter should be nil
func ValidateFormError(t TestLog, err error, validatorMap map[string]string) {
	t.Helper()

	if err != nil {
		var valErr validation.Errors

		// If error is validation.Errors, determine if passed validationMap
		// errors correspond with errors in validation.Errors map
		//
		// Else simply return err
		if errors.As(err, &valErr) {
			errBytes, err := err.(validation.Errors).MarshalJSON()
			if err != nil {
				t.Errorf("%+v", err)
				return
			}

			var errMap map[string]any

			if err = json.Unmarshal(errBytes, &errMap); err != nil {
				t.Errorf("%+v", err)
				return
			}

			expectedNotFoundStr := ""
			valueMisMatchStr := ""
			errNotFoundStr := ""

			errFlatMap, err := flat.Flatten(errMap, nil)
			if err != nil {
				t.Errorf(err.Error())
			}

			for validatorKey, validatorVal := range validatorMap {
				if errVal, ok := errFlatMap[validatorKey]; ok {
					if errVal != validatorVal {
						valueMisMatchStr += fmt.Sprintf("(key: %s) expected: %s; got: %s\n", validatorKey, validatorVal, errVal)
					}
				} else {
					expectedNotFoundStr += fmt.Sprintf("key: %s; value: %s\n", validatorKey, validatorVal)
				}
			}

			for errKey, errVal := range errFlatMap {
				if _, ok := validatorMap[errKey]; !ok {
					errNotFoundStr += fmt.Sprintf("key: %s; value: %s\n", errKey, errVal)
				}
			}

			errStr := ""

			if expectedNotFoundStr != "" {
				errStr += "\nThe following key/values were given but not found:\n" + expectedNotFoundStr
			}
			if valueMisMatchStr != "" {
				errStr += "\nThe following key/values don't match:\n" + valueMisMatchStr
			}
			if errNotFoundStr != "" {
				errStr += "\nThe following key/values were found but not expected:\n" + errNotFoundStr
			}

			if errStr != "" {
				t.Errorf(errStr)
			}
		} else {
			t.Errorf("%+v", err)
			return
		}
	} else {
		if validatorMap != nil {
			t.Errorf("There are no form errors but validatorMap was passed")
		}
	}
}

// RandomString is util test function that takes in int and will return random string with that length
//
// This is useful for testing for fields with min/max length constraints
func RandomString(len int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:len]
}
