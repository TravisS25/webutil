package webutil

import (
	"fmt"
	"os"
	"reflect"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

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
	if config.ErrorStatus == 0 {
		config.ErrorStatus = defaultStatus
	}
	if config.ErrorResponse == nil {
		config.ErrorResponse = defaultResponse
	}
}

// NewV7UUID is wrapper for uuid.NewV7()
func NewV7UUID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

func NewV7UUIDString() string {
	return uuid.Must(uuid.NewV7()).String()
}

func SetConfigSettings(path string, settings any, opts ...viper.DecoderConfigOption) error {
	var err error

	if reflect.ValueOf(settings).Type().Kind() != reflect.Pointer {
		return fmt.Errorf("settings parameter must be pointer")
	}

	var fp string

	if os.Getenv(path) != "" {
		fp = os.Getenv(path)
	} else {
		fp = path
	}

	v := viper.New()
	v.SetConfigFile(fp)

	if err = v.ReadInConfig(); err != nil {
		return err
	}
	if err = v.Unmarshal(settings, opts...); err != nil {
		return err
	}

	return nil
}
