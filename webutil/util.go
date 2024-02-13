package webutil

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

// GetHTTPResponseDefaults is util function to get HTTPResponseConfig instance with values passed
func GetHTTPResponseDefaults(defaultStatus int, defaultResponse []byte) HTTPResponseConfig {
	res := HTTPResponseConfig{}
	SetHTTPResponseDefaults(&res, defaultStatus, defaultResponse)
	return res
}
