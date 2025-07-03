package webutil

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type BaseAuth struct {
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
}

// DatabaseSetting is config struct to set up database connection
type DatabaseSetting struct {
	User        string `yaml:"user" mapstructure:"user"`
	Password    string `yaml:"password" mapstructure:"password"`
	Host        string `yaml:"host" mapstructure:"host"`
	Port        int    `yaml:"port" mapstructure:"port"`
	DBType      string `yaml:"db_type" mapstructure:"db_type"`
	DBName      string `yaml:"db_name" mapstructure:"db_name"`
	SSLMode     string `yaml:"ssl_mode" mapstructure:"ssl_mode"`
	SSLRootCert string `yaml:"ssl_root_cert" mapstructure:"ssl_root_cert"`
	SSLKey      string `yaml:"ssl_key" mapstructure:"ssl_key"`
	SSLCert     string `yaml:"ssl_cert" mapstructure:"ssl_cert"`
	SearchPath  string `yaml:"search_path" mapstructure:"search_path"`
}

func (d DatabaseSetting) String() string {
	return fmt.Sprintf(
		DB_CONN_STR,
		d.DBType,
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.DBName,
		d.SSLMode,
		d.SSLRootCert,
		d.SSLKey,
		d.SSLCert,
		d.SearchPath,
	)
}

func (d *DatabaseSetting) ParseConnStr(connStr string) error {
	dbURL, err := url.Parse(connStr)
	if err != nil {
		return err
	}

	var port int

	if dbURL.Port() != "" {
		port, _ = strconv.Atoi(dbURL.Port())
	}

	var dbName string

	paths := strings.Split(dbURL.Path, "/")

	if len(paths) > 0 {
		dbName = paths[0]
	}

	d.User = dbURL.User.Username()
	d.Password, _ = dbURL.User.Password()
	d.Host = dbURL.Host
	d.Port = port
	d.DBType = dbURL.Scheme
	d.DBName = dbName
	d.SSLMode = dbURL.Query().Get("ssl_mode")
	d.SSLRootCert = dbURL.Query().Get("ssl_root_cert")
	d.SSLKey = dbURL.Query().Get("ssl_key")
	d.SSLCert = dbURL.Query().Get("ssl_cert")
	d.SearchPath = dbURL.Query().Get("search_path")

	return nil
}

// S3StorageSetting is setting for S3 backend
type S3StorageSetting struct {
	EndPoint        string `yaml:"end_point" mapstructure:"end_point"`
	AccessKeyID     string `yaml:"access_key_id" mapstructure:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key" mapstructure:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl" mapstructure:"use_ssl"`
}

type SessionAuth struct {
	AuthKey    string `mapstructure:"auth_key"`
	EncryptKey string `mapstructure:"encrypt_key"`
}

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
