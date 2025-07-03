package webutil

import "github.com/google/uuid"

//////////////////////////////////////////////////////////////////
//------------------------ FUNCTIONS --------------------------
//////////////////////////////////////////////////////////////////

// InsertAt is util function to insert passed value into passed slice
// at passed index
func InsertAt(slice []any, val any, idx int) []any {
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

// NewV7UUID is wrapper for uuid.NewV7()
func NewV7UUID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

func NewV7UUIDString() string {
	return uuid.Must(uuid.NewV7()).String()
}
