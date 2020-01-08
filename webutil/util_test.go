package webutil

import (
	"os"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/spf13/viper"
)

const (
	// WebUtilTestConfig is env var that should be set to point to file that
	// contains test config settings for integration tests
	WebUtilTestConfig = "WEB_UTIL_TEST_CONFIG"
)

var (
	sqlAnyMatcher = sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		return nil
	})
	testConf *testConfig
)

type testConfig struct {
	DBResetConfiguration dbResetConfiguration `yaml:"db_reset_configuration" mapstructure:"db_reset_configuration"`
	DBConnections        []DatabaseSetting    `yaml:"db_connections" mapstructure:"db_connections"`
}

type dbCommand struct {
	Command string   `yaml:"command" mapstructure:"command"`
	Args    []string `yaml:"args" mapstructure:"args"`
}

type dbResetConfiguration struct {
	DbConnections  []DatabaseSetting `yaml:"db_connections" mapstructure:"db_connections"`
	DbStopCommand  dbCommand         `yaml:"db_stop_command" mapstructure:"db_stop_command"`
	DbStartCommand dbCommand         `yaml:"db_start_command" mapstructure:"db_start_command"`
	ValidateQuery  string            `yaml:"validate_query" mapstructure:"validate_query"`
}

func initTestConfig() {
	var err error

	filePath := os.Getenv(WebUtilTestConfig)

	if filePath == "" {
		panic("env var not set\n")
	}

	v := viper.New()
	v.SetConfigFile(filePath)

	if err = v.ReadInConfig(); err != nil {
		panic(err.Error())
	}
	if err = v.Unmarshal(&testConf); err != nil {
		panic(err.Error())
	}
}

func init() {
	initTestConfig()
}
