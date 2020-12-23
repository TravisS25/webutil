package webutil

import (
	"encoding/json"
	"net/http"
)

type LogConfig struct {
	CauseErr error

	RecoverDBErr error
	RetryDBErr   error

	RecoverCacheErr error
	RetryCacheErr   error

	RecoverFormErr error
	RetryFormErr   error

	PanicErr error
}

// RequestLog is func that should implement logging for when
// server error occurs
type RequestLog func(r *http.Request, conf LogConfig)

// ServerErrorConfig is a generic config struct to use
// with various functions to give default responses if error occurs
type ServerErrorConfig struct {
	RecoverConfig

	// ServerErrorResponse is used to give response header and text
	// of when a server error occurs
	ServerErrorResponse HTTPResponseConfig

	// Logger is used to take in a request with thrown
	// error and implement custom logging of error
	Logger RequestLog
}

// RecoverConfig is config struct used to allow user to implement
// a way of recovering from different types of failures
type RecoverConfig struct {
	// RecoverDB is func that should be used to try to recover
	// from a db failure
	RecoverDB RecoverDB

	// RecoverCache is func that should be used to try to recover
	// from cache failure
	RecoverCache RecoverCache

	// RecoverForm is func that should be used to try to recover
	// from form failure
	RecoverForm RecoverForm

	// // ResetDB is optional parameter generally used in api
	// // enpoints or form validators to reset their DBInterface after
	// // being recovered from RecoverDB function
	// ResetDB ResetDB

	// // ResetCache is optional parameter generally used in api
	// // enpoints or form validators to reset their CacheStore after
	// // being recovered from RecoverCache function
	// ResetCache ResetCache
}

// ServerErrorCacheConfig is config struct used to respond to server
// error but also have ability to use cache
type ServerErrorCacheConfig struct {
	ServerErrorConfig
	CacheConfig
}

//////////////////////////////////////////////////////////////////
//------------------------ FUNCTIONS --------------------------
//////////////////////////////////////////////////////////////////

// InsertAt is util function to insert passed value into passed slice
// at passed index
func InsertAt(slice []interface{}, val interface{}, idx int) []interface{} {
	if len(slice) == 0 {
		slice = append(slice, val)
	} else {
		slice = append(slice, 0)
		copy(slice[idx+1:], slice[idx:])
		slice[idx] = val
	}

	return slice
}

// SetHTTPResponseDefaults is util function to set default values for passed
// config if values for nil
func SetHTTPResponseDefaults(config *HTTPResponseConfig, defaultStatus int, defaultResponse []byte) {
	if config.HTTPStatus == nil {
		config.HTTPStatus = &defaultStatus
	}
	if config.HTTPResponse == nil {
		config.HTTPResponse = defaultResponse
	}
}

// GetHTTPResponseDefaults is util function to get HTTPResponseConfig instance with
// values passed
func GetHTTPResponseDefaults(defaultStatus int, defaultResponse []byte) HTTPResponseConfig {
	res := HTTPResponseConfig{}
	SetHTTPResponseDefaults(&res, defaultStatus, defaultResponse)
	return res
}

func ExtendJSONMarshal(val interface{}, extBytes []byte) ([]byte, error) {
	entityBytes, err := json.Marshal(&val)

	if err != nil {
		return nil, err
	}

	t := entityBytes[:len(entityBytes)-1]
	t = append(t, ',')
	v := extBytes[1:]

	t = append(t, v...)

	return t, nil
}
