package webutiltest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/mux"

	"github.com/TravisS25/webutil/webutil"

	validation "github.com/go-ozzo/ozzo-validation"
)

// FormRequestConfig is config struct used to run against
// RunRequestFormTests function
type FormRequestConfig struct {
	// TestName is the name of current test - Required
	TestName string

	// Method is http method to use for request - Optional
	Method string

	// URL is uri you wish to request - Optional
	URL string

	// Validatable is used for forms that validate themselves, generally inner forms - Optional
	Validatable validation.Validatable

	// Validate is interface for struct that will validate form - Optional
	Validator webutil.RequestValidator

	// FileUploadConf is used to simulate file upload to request and be used
	// within form validation
	FileUploadConf *FileUploadConfig

	// Form is form values to use to inject into request - Required
	Form interface{}

	// Instance is instance of a model in which a form might need, usually
	// on an edit request - Optional
	Instance interface{}

	// RouterValues is used to inject router variables into the request - Optional
	RouterValues map[string]string

	// ContextValues are context#Context to use for request - Optional
	ContextValues map[interface{}]interface{}

	// PostExecute can be used to exec some logic that you may need to run inbetween test cases
	// such as clean up logic before the next test is run - Optional
	PostExecute func(form interface{}) error

	// ValidationErrors is a map of what errors you expect to return from test
	// The key is the json name of the field and value is the error message the
	// field should return - Optional
	ValidationErrors map[string]interface{}

	InternalError string
}

// RunRequestFormTests runs tests against the config struct it is given for
// form validation
//
// The deferFunc parameter is used to execute if a panic occurs during a test
// This can happen due any number of reasons but is used here in mind
// for cases where we are performing integration tests with a test database
// If we have a panic, the test caller of this function doesn't get to finish
// execution so if a user enters information into database, then calls this function
// and it panics, there is no teardown of the database as the panic stops the execution
// of the caller function
func RunRequestFormTests(t *testing.T, deferFunc func() error, formTests []FormRequestConfig) {
	for _, formTest := range formTests {
		if formTest.TestName == "" {
			t.Fatalf("TestName required")
		}
		if formTest.Validatable == nil && formTest.Validator == nil && formTest.FileUploadConf == nil {
			t.Fatalf("Validatable, Validator or FileUploadConfs is required")
		}
		if formTest.Method == "" {
			formTest.Method = http.MethodGet
		}
		if formTest.URL == "" {
			formTest.URL = "/url"
		}

		t.Run(formTest.TestName, func(s *testing.T) {
			var formErr error
			var form interface{}

			var req *http.Request
			var err error

			panicked := true
			defer func() {
				if deferFunc != nil && panicked {
					err := deferFunc()

					if err != nil {
						fmt.Printf("deferFunc: " + err.Error())
					}
				}
			}()

			setFormValues := func() {
				if formTest.ContextValues != nil {
					ctx := req.Context()

					for key, value := range formTest.ContextValues {
						ctx = context.WithValue(ctx, key, value)
					}

					req = req.WithContext(ctx)
				}

				req = mux.SetURLVars(req, formTest.RouterValues)
				form, formErr = formTest.Validator.Validate(req, formTest.Instance)
			}

			if formTest.Validatable != nil {
				formErr = formTest.Validatable.Validate()
			} else if formTest.Form != nil {
				jsonBytes, err := json.Marshal(&formTest.Form)

				if err != nil {
					s.Fatalf(err.Error())
				}

				buf := bytes.NewBuffer(jsonBytes)
				req, err = http.NewRequest(formTest.Method, formTest.URL, buf)

				if err != nil {
					s.Fatalf(err.Error())
				}

				setFormValues()
			} else if formTest.FileUploadConf != nil {
				req, err = NewFileUploadRequest(
					formTest.FileUploadConf.ParamConfs,
					formTest.Method,
					formTest.URL,
				)

				if err != nil {
					s.Fatal(err)
				}

				if err = req.ParseMultipartForm(formTest.FileUploadConf.MaxMemory); err != nil {
					s.Fatalf(err.Error())
				}

				setFormValues()
			} else {
				req, err = http.NewRequest(formTest.Method, formTest.URL, nil)

				if err != nil {
					s.Fatalf(err.Error())
				}

				setFormValues()
			}

			if formErr == nil {
				if formTest.ValidationErrors != nil {
					s.Errorf("Form has no errors, but 'ValidationErrors' was passed\n")
				}
			} else {
				if validationErrors, ok := formErr.(validation.Errors); ok {
					foundKeys := make(map[string]bool, 0)

					for key, expectedVal := range formTest.ValidationErrors {
						if fErr, valid := validationErrors[key]; valid {
							foundKeys[key] = true
							err := formValidation(s, key, fErr, expectedVal)

							if err != nil {
								s.Errorf(err.Error())
							}
						} else {
							s.Errorf("Key \"%s\" found in \"ValidationErrors\" that is not in form errors\n\n", key)
						}
					}

					for k, v := range validationErrors {
						if fErr, valid := formTest.ValidationErrors[k]; valid {
							if _, found := foundKeys[k]; !found {
								err := formValidation(s, k, v, fErr)

								if err != nil {
									s.Errorf(err.Error())
								}
							}
						} else {
							s.Errorf(
								"Key \"%s\" found in form errors that is not in \"ValidationErrors\"\n  Threw err: %s\n\n",
								k,
								v.Error(),
							)
						}
					}
				} else {
					if formErr == webutil.ErrBodyRequired || formErr == webutil.ErrInvalidJSON {
						s.Errorf("%+v\n", formErr)
					} else {
						s.Errorf("Internal Error: %+v\n", formErr)
					}
				}
			}

			if formTest.PostExecute != nil {
				if err = formTest.PostExecute(form); err != nil {
					t.Errorf("post execute error: %+v\n", err)
				}
			}

			panicked = false
		})
	}
}

func formValidation(t *testing.T, mapKey string, formValidationErr error, expectedErr interface{}) error {
	var err error

	if innerExpectedErr, k := expectedErr.(map[string]interface{}); k {
		if innerFormErr, j := formValidationErr.(validation.Errors); j {
			for innerExpectedKey := range innerExpectedErr {
				if innerFormVal, ok := innerFormErr[innerExpectedKey]; ok {
					innerExpectedVal := innerExpectedErr[innerExpectedKey]

					switch innerExpectedVal.(type) {
					case map[string]interface{}:
						err = formValidation(t, innerExpectedKey, innerFormVal, innerExpectedVal)

						if err != nil {
							return err
						}
					case string:
						if len(innerExpectedErr) != len(innerFormErr) {
							if len(innerExpectedErr) > len(innerFormErr) {
								for k := range innerExpectedErr {
									if _, ok := innerFormErr[k]; !ok {
										t.Errorf("form testing: Key \"%s\" found in \"ValidationErrors\" that is not in form errors", k)
									}
								}
							} else {
								for k, v := range innerFormErr {
									if _, ok := innerExpectedErr[k]; !ok {
										//t.Errorf("heeeey type: %v", validationErrors["invoiceItems"]["0"])
										t.Errorf("form testing: Key \"%s\" found in form errors that is not in \"ValidationErrors\"\n  Key \"%s\" threw err: %s\n", k, k, v.Error())
									}
								}
							}
						}

						if innerFormVal.Error() != innerExpectedVal {
							t.Errorf(
								"form testing: Key \"%s\" threw err: \"%s\" \n expected: \"%s\" \n",
								innerExpectedKey,
								innerFormVal.Error(),
								innerExpectedVal,
							)
						}
					default:
						message := fmt.Sprintf("form testing: Passed \"ValidationErrors\" has unexpected type\n")
						return errors.New(message)
					}
				} else {
					t.Errorf("form testing: Key \"%s\" was in \"ValidationErrors\" but not form errors\n", innerExpectedKey)
				}
			}
		} else {
			message := fmt.Sprintf(
				"form testing: \"ValidationErrors\" error for key \"%s\" was type map but form error was not\n.  Error thrown: %s", mapKey, formValidationErr,
			)
			return errors.New(message)
		}
	} else {
		if formValidationErr.Error() != expectedErr {
			t.Errorf("form testing: Key \"%s\" threw err: \"%s\" \n expected: \"%s\" \n", mapKey, formValidationErr.Error(), expectedErr)
		}
	}

	return nil
}
