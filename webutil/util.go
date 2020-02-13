package webutil

// ServerErrorConfig is a generic config struct to use
// with various functions to give default responses if error occurs
type ServerErrorConfig struct {
	RecoverConfig

	// ServerErrorResponse is used to give response header and text
	// of when a server error occurs
	ServerErrorResponse HTTPResponseConfig

	// ClientErrorResponse is used to give response header and text
	// of when a client error occurs
	ClientErrorResponse HTTPResponseConfig
}

// RecoverConfig is config struct used to allow user to implement
// a way of recovering from a db failure and optionally
// re-querying the db
type RecoverConfig struct {
	// RecoverDB is func that should be used to try to recover
	// from a db failure
	RecoverDB RecoverDB

	// DBInterfaceRecover is optional parameter generally used in api
	// enpoints or form validators to reset their DBInterface after
	// being recovered from RecoverDB function
	DBInterfaceRecover DBInterfaceRecover
}

// ServerErrorCacheConfig is config struct used to respond to server
// error but also have ability to use cache
type ServerErrorCacheConfig struct {
	ServerErrorConfig
	CacheConfig
}

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
