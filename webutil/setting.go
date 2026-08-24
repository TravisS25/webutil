package webutil

import (
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/viper"
)

func SetConfigSettings(path string, settings any, opts ...viper.DecoderConfigOption) error {
	var err error

	if reflect.ValueOf(settings).Type().Kind() != reflect.Ptr {
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
