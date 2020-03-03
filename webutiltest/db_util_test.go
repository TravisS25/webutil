package webutiltest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/TravisS25/webutil/webutil"
	"github.com/gofrs/uuid"
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
	DBTestConfig        dbTestConfig       `yaml:"db_test_config" mapstructure:"db_test_config"`
	TestFileUploadConfs []FileUploadConfig `yaml:"test_file_upload_confs" mapstructure:"test_file_upload_confs"`
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

	buf := &bytes.Buffer{}

	removeFn := func() error {
		cmd := exec.Command(
			testConf.DBTestConfig.RemoveDataCommand.Command,
			testConf.DBTestConfig.RemoveDataCommand.Args...,
		)
		return cmd.Run()
	}
	defer removeFn()

	cmd := exec.Command(
		testConf.DBTestConfig.CreateDBCommand.Command,
		testConf.DBTestConfig.CreateDBCommand.Args...,
	)
	cmd.Stderr = buf

	if err = cmd.Run(); err != nil {
		t.Fatalf(buf.String())
	}

	cmd = exec.Command(
		testConf.DBTestConfig.LoadDataCommand.Command,
		testConf.DBTestConfig.LoadDataCommand.Args...,
	)
	cmd.Stderr = buf

	if err = cmd.Run(); err != nil {
		removeFn()
		t.Fatalf(buf.String())
	}

	db, err := webutil.NewDB(testConf.DBTestConfig.DBConnection, testConf.DBTestConfig.DBType)

	if err != nil {
		removeFn()
		t.Fatalf(err.Error())
	}

	if _, err = db.Exec(
		`
		insert into database_table (id, name, display_name, column_name)
		values(1, 'user_profile', 'User Profile', 'last_name');
		`,
	); err != nil {
		removeFn()
		t.Fatalf(err.Error())
	}

	id, _ := uuid.NewV4()

	type userProfile struct {
		ID        int64   `json:"id,string"`
		FirstName string  `json:"firstName"`
		LastName  string  `json:"lastName"`
		IsActive  bool    `json:"isActive"`
		LastLogin *string `json:"lastLogin"`
	}

	user := userProfile{
		ID:        1,
		FirstName: "first",
		LastName:  "last",
		IsActive:  true,
	}

	userBytes, err := json.Marshal(user)

	if err != nil {
		removeFn()
		t.Fatal(err.Error())
	}

	if _, err = db.Exec(
		`
		insert into user_profile(id, email, first_name, last_name, is_active, last_login)
		values(1, 'test@email.com', 'first', 'last', true, null);
		`,
	); err != nil {
		t.Fatalf(err.Error())
	}

	query :=
		`
	insert into public.logging (id, date_created, data, primary_key_id, primary_key_uuid, been_viewed, database_action_id, database_table_id, user_profile_id)
	values(?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	var bindVar int

	switch testConf.DBTestConfig.DBType {
	case webutil.Postgres:
		bindVar = sqlx.DOLLAR
	default:
		bindVar = sqlx.QUESTION
	}

	var emptyStr *string

	if query, args, err = webutil.InQueryRebind(
		bindVar,
		query,
		id.String(),
		emptyStr,
		userBytes,
		1,
		emptyStr,
		false,
		1,
		1,
		1,
	); err != nil {
		removeFn()
		t.Fatalf(err.Error())
	}

	fmt.Printf("made past rebind")
	fmt.Printf("query: %s\n", query)
	fmt.Printf("args: %v\n", args)

	if _, err = db.Exec(query, args...); err != nil {
		removeFn()
		t.Fatalf(err.Error())
	}

	fmt.Printf("made past exec")

	teardown := DBSetup(db, sqlx.DOLLAR)

	if err = teardown(); err != nil {
		removeFn()
		t.Fatalf(err.Error())
	}

	fmt.Printf("made past teardown")
}
