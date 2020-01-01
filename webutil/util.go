package webutil

import (
	"fmt"

	"github.com/pkg/errors"
)

// ServerErrorConfig is a generic config struct to use
// with various functions to give default responses if error occurs
type ServerErrorConfig struct {
	// ServerErrorConf is used to give response header and text
	// of when a server error occurs
	ServerErrorConf HTTPResponseConfig

	// RecoverDB is optional parameter used to try to recover
	// from error if one occurs
	//
	// Allowed to be nil
	RecoverDB RecoverDB
}

// ServerAndClientErrorConfig is wrapper struct for ServerErrorConfig
// with extra config field for setting response if error is
// a client error
type ServerAndClientErrorConfig struct {
	ServerErrorConfig

	// ClientErrorConf is used to give response header and text
	// of when a client error occurs
	ClientErrorConf HTTPResponseConfig
}

// CheckError simply prints given error in verbose to stdout
func CheckError(err error, customMessage string) {
	err = errors.Wrap(err, customMessage)
	fmt.Printf("%+v\n", err)
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
