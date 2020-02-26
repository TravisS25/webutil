package webutiltest

import (
	"os"
	"os/exec"
	"testing"

	"github.com/TravisS25/webutil/webutil"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var (
	testConf *testConfig

	// WebUtilTestConfig is env var that should be set to point to file that
	// contains test config settings for integration tests
	WebUtilTestConfig = "WEB_UTIL_TEST_CONFIG"
)

type testConfig struct {
	DBTestConfig dbTestConfig `yaml:"db_test_config" mapstructure:"db_test_config"`
}

type cmdCommand struct {
	Command string   `yaml:"command" mapstructure:"command"`
	Args    []string `yaml:"args" mapstructure:"args"`
}

type dbTestConfig struct {
	DBType            string                  `yaml:"db_type" mapstructure:"db_type"`
	CreateDBCommand   cmdCommand              `yaml:"create_db_command" mapstructure:"create_db_command"`
	LoadDataCommand   cmdCommand              `yaml:"load_data_command" mapstructure:"load_data_command"`
	RemoveDataCommand cmdCommand              `yaml:"remove_data_command" mapstructure:"remove_data_command"`
	DBConnection      webutil.DatabaseSetting `yaml:"db_connection" mapstructure:"db_connection"`
}

func init() {
	initConfig()
}

func initConfig() {
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

func TestDBSetupIntegrationTest(t *testing.T) {
	var err error
	var args []interface{}

	cmd := exec.Command(
		testConf.DBTestConfig.CreateDBCommand.Command,
		testConf.DBTestConfig.CreateDBCommand.Args...,
	)

	if err = cmd.Run(); err != nil {
		t.Fatalf(err.Error())
	}

	cmd = exec.Command(
		testConf.DBTestConfig.LoadDataCommand.Command,
		testConf.DBTestConfig.LoadDataCommand.Args...,
	)

	if err = cmd.Run(); err != nil {
		cmd = exec.Command(
			testConf.DBTestConfig.RemoveDataCommand.Command,
			testConf.DBTestConfig.RemoveDataCommand.Args...,
		)
		cmd.Run()
		t.Fatalf(err.Error())
	}

	db, err := webutil.NewDB(testConf.DBTestConfig.DBConnection, testConf.DBTestConfig.DBType)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if _, err = db.Exec(
		`
		insert into database_table (id, name, display_name, column_name)
		values(1, 'user_profile', 'User Profile', 'last_name');
		`,
	); err != nil {
		t.Fatalf(err.Error())
	}

	query :=
		`
	insert into logging (id, data_created, data, primary_key_id, primary_key_uuid, been_viewed, database_action_id, database_table_id, user_profile_id)
	values(?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	var bindVar int

	switch testConf.DBTestConfig.DBType {
	case webutil.Postgres:
		bindVar = sqlx.DOLLAR
	default:
		bindVar = sqlx.QUESTION
	}

	query, args, err = webutil.InQueryRebind(bindVar, query)

	_, err = db.Exec(
		`
		insert into logging (id, data_created, data, primary_key_id, primary_key_uuid, been_viewed, database_action_id, database_table_id, user_profile_id)
		values();
		`,
	)
}
