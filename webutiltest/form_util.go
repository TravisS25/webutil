package webutiltest

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/objx"

	validation "github.com/go-ozzo/ozzo-validation"
)

func ValidateFormError(t *testing.T, err error, validatorMap map[string]string) {
	if err != nil {
		t.Helper()

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

			var errMap map[string]interface{}

			if err = json.Unmarshal(errBytes, &errMap); err != nil {
				t.Errorf("%+v", err)
				return
			}

			if validatorMap == nil {
				t.Errorf("There are form errors but validatorMap is nil.  Errors: %v\n", errMap)
				return
			}

			keys := make([]string, 0, len(validatorMap))

			// Get keys from passed validatorMap to loop through
			// and compare against form error
			for k := range validatorMap {
				keys = append(keys, k)
			}

			errObj := objx.New(errMap)
			msg := ""

			for _, key := range keys {
				// Use given key to get value from our errObj
				errObjVal := errObj.Get(key)

				// If value is nil in errObj, this tells us that given key
				// is not in errObj so add to our error msg
				//
				// Else compare values
				if errObjVal.IsNil() {
					msg += fmt.Sprintf("Key \"%s\" was not found in form errors\n", key)
				} else {
					if errObjVal.Str() != validatorMap[key] {
						msg += fmt.Sprintf(
							"Key \"%s\" threw err: \"%s\" \n expected: \"%s\"\n",
							key,
							errObjVal.Str(),
							validatorMap[key],
						)
					}

					// walkMapFunc helps us remove key from errMap so at
					// the end of the key loop, if there is any values left,
					// those will be given back in an error at those will be
					// errors that weren't in the validatorMap passed
					var walkMapFunc func(string)

					walkMapFunc = func(mapKey string) {
						mapLvls := strings.Split(mapKey, ".")

						// getMapKey is anon util function that helps us get the key
						// for the map above current mapKey
						getMapKey := func() string {
							k := ""

							for i, v := range mapLvls[:len(mapLvls)-1] {
								k += v

								if i != len(mapLvls)-2 {
									k += "."
								}
							}

							return k
						}

						// If errObj with mapKey currently has str value, we know we want to
						// delete this entry
						//
						// Else mapKey should be for entry for a map[string]interface{}
						if errObj.Get(mapKey).IsStr() {
							// If mapLvls is 1, then current key is entry in our errMap so simply delete
							//
							// Else current key is for nested map so we must traverse
							if len(mapLvls) == 1 {
								delete(errMap, mapKey)
							} else {
								// Get key for map above current key entry and remove entry from that map
								k := getMapKey()
								m := errObj.Get(k).Data().(map[string]interface{})
								mapProperty := mapLvls[len(mapLvls)-1]
								delete(m, mapProperty)

								// Pass in mapKey that got us map "m" into walkMapFunc for recursion
								walkMapFunc(k)
							}
						} else {
							// Get map[string]interface{} from mapKey
							tmpMap := errObj.Get(mapKey).Data().(map[string]interface{})

							// If map has no entries, begin trying to remove
							if len(tmpMap) == 0 {
								// If mapLvls is 1, then current map is entry in our errMap
								// so simple remove
								//
								// Else get map above current map and delete current map from
								// above map entry
								if len(mapLvls) == 1 {
									delete(errMap, mapKey)
								} else {
									k := getMapKey()
									m := errObj.Get(k).Data().(map[string]interface{})
									mapProperty := mapLvls[len(mapLvls)-1]
									delete(m, mapProperty)

									// Pass in mapKey that got us map "m" into walkMapFunc for recursion
									walkMapFunc(k)
								}
							}
						}
					}

					walkMapFunc(key)
				}
			}

			if len(errMap) > 0 {
				msg += fmt.Sprintf("\n\n These are form errors not found in validatorMap: %v\n", errMap)
			}

			if msg != "" {
				t.Errorf(msg)
				return
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

	//return nil
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
