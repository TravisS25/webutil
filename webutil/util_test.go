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
	DBTestConf  dbTestConfig  `yaml:"db_test_config" mapstructure:"db_test_config"`
}

type dbTestConfig struct {
	CreateDBCommand   cmdCommand `yaml:"create_db_command" mapstructure:"create_db_command"`
	LoadDataCommand   cmdCommand `yaml:"load_data_command" mapstructure:"load_data_command"`
	RemoveDataCommand cmdCommand `yaml:"remove_data_command" mapstructure:"remove_data_command"`
}

type portConfig struct {
	FlagKey    string `yaml:"flag_key" mapstructure:"flag_key"`
	Port       string `yaml:"port" mapstructure:"port"`
	DockerPort string `yaml:"docker_port" mapstructure:"docker_port"`
}

type teardownConfig struct {
	ChosenPort int
	DockerName string
}

type cmdCommand struct {
	Command string   `yaml:"command" mapstructure:"command"`
	Args    []string `yaml:"args" mapstructure:"args"`
}

type dbCommand struct {
	CmdCommand cmdCommand `yaml:"cmd_command" mapstructure:"cmd_command"`
	PortConfig portConfig `yaml:"port_config" mapstructure:"port_config"`
}

type dbResetConfig struct {
	DBType          string            `yaml:"db_type" mapstructure:"db_type"`
	MaxPortAttempts int               `yaml:"max_port_attempts" mapstructure:"max_port_attempts"`
	BaseConnection  DatabaseSetting   `yaml:"base_connection" mapstructure:"base_connection"`
	DBConnections   []DatabaseSetting `yaml:"db_connections" mapstructure:"db_connections"`
	DBRemoveCommand dbCommand         `yaml:"db_remove_command" mapstructure:"db_remove_command"`
	DBStartCommand  dbCommand         `yaml:"db_start_command" mapstructure:"db_start_command"`
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

func init() {
	//initTestConfig()
}
