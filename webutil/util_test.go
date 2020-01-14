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
	DBResetConf dbResetConfig `yaml:"db_reset_configuration" mapstructure:"db_reset_configuration"`
	//DBConnections        []DatabaseSetting    `yaml:"db_connections" mapstructure:"db_connections"`
}

type portConfig struct {
	FlagKey    string `yaml:"flag_key" mapstructure:"flag_key"`
	Port       string `yaml:"port" mapstructure:"port"`
	DockerPort string `yaml:"docker_port" mapstructure:"docker_port"`
	// FlagValueFormat string `yaml:"flag_value_format" mapstructure:"flag_value_format"`
	// NumOfArgs       uint16 `yaml:"num_of_args" mapstructure:"num_of_args"`
}

type teardownConfig struct {
	ChosenPort int
	DockerName string
}

type dbCommand struct {
	Command    string     `yaml:"command" mapstructure:"command"`
	Args       []string   `yaml:"args" mapstructure:"args"`
	PortConfig portConfig `yaml:"port_config" mapstructure:"port_config"`
}

type dbResetConfig struct {
	DBType          string            `yaml:"db_type" mapstructure:"db_type"`
	MaxPortAttempts int               `yaml:"max_port_attempts" mapstructure:"max_port_attempts"`
	BaseConnection  DatabaseSetting   `yaml:"base_connection" mapstructure:"base_connection"`
	DBConnections   []DatabaseSetting `yaml:"db_connections" mapstructure:"db_connections"`
	DBRemoveCommand dbCommand         `yaml:"db_remove_command" mapstructure:"db_remove_command"`
	DBStartCommand  dbCommand         `yaml:"db_start_command" mapstructure:"db_start_command"`
	DBRunCommand    dbCommand         `yaml:"db_run_command" mapstructure:"db_run_command"`
	ValidateQuery   string            `yaml:"validate_query" mapstructure:"validate_query"`
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

// func initDB() {
// 	if db == nil {
// 		dbMutex.Lock()
// 		defer dbMutex.Unlock()
// 		if db == nil {
// 			var err error
// 			db, err = NewDBWithList(testConf.DBConnections, Postgres)

// 			if err != nil {
// 				panic(err.Error())
// 			}

// 		}
// 	}
// }

func init() {
	initTestConfig()
}
