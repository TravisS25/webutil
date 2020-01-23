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
	PostExecute func(form interface{})

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
		if formTest.Validatable == nil && formTest.Validator == nil {
			t.Fatalf("Validatable or validator is required")
		}
		if formTest.Method == "" {
			formTest.Method = http.MethodGet
		}
		if formTest.URL == "" {
			formTest.URL = "/url"
		}

		t.Run(formTest.TestName, func(t *testing.T) {
			var formErr error
			var form interface{}

			panicked := true
			defer func() {
				if deferFunc != nil && panicked {
					err := deferFunc()

					if err != nil {
						fmt.Printf("deferFunc: " + err.Error())
					}
				}
			}()

			if formTest.Validatable != nil {
				formErr = formTest.Validatable.Validate()
			} else {
				jsonBytes, err := json.Marshal(&formTest.Form)

				if err != nil {
					t.Fatalf(err.Error())
				}

				buf := bytes.NewBuffer(jsonBytes)
				req, err := http.NewRequest(formTest.Method, formTest.URL, buf)

				if err != nil {
					t.Fatalf(err.Error())
				}

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

			if formErr == nil {
				if formTest.ValidationErrors != nil {
					t.Errorf("Form has no errors, but 'ValidationErrors' was passed\n")
				}
			} else {
				if validationErrors, ok := formErr.(validation.Errors); ok {
					//fmt.Printf("validation err: %v\n", validationErrors)

					for key, expectedVal := range formTest.ValidationErrors {
						if fErr, valid := validationErrors[key]; valid {
							err := formValidation(t, key, fErr, expectedVal)

							if err != nil {
								t.Errorf(err.Error())
							}
						} else {
							t.Errorf("Key \"%s\" found in \"ValidationErrors\" that is not in form errors\n\n", key)
						}
					}

					for k, v := range validationErrors {
						if fErr, valid := formTest.ValidationErrors[k]; valid {
							err := formValidation(t, k, v, fErr)

							if err != nil {
								t.Errorf(err.Error())
							}
						} else {
							t.Errorf(
								"Key \"%s\" found in form errors that is not in \"ValidationErrors\"\n  Threw err: %s\n\n",
								k,
								v.Error(),
							)
						}
					}
				} else {
					if formTest.InternalError != formErr.Error() {
						t.Errorf("Internal Error: %s\n", formErr.Error())
					}
				}
			}

			if formTest.PostExecute != nil {
				formTest.PostExecute(form)
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
						//fmt.Printf("map val switch\n")
						err = formValidation(t, innerExpectedKey, innerFormVal, innerExpectedVal)

						if err != nil {
							//fmt.Printf("form err\n")
							return err
						}
					case string:
						//fmt.Printf("string val switch\n")

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
		//fmt.Printf("made to non map\n")
		if formValidationErr.Error() != expectedErr {
			t.Errorf("form testing: Key \"%s\" threw err: \"%s\" \n expected: \"%s\" \n", mapKey, formValidationErr.Error(), expectedErr)
		}
	}

	return nil
}