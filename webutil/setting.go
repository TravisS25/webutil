package webutil

import (
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/viper"
)

// SessionSetting is config struct for setting up session AuthEncryptionSetting
// for redis server
type SessionSetting struct {
	SessionAuth `yaml:"auth_encryption_setting" mapstructure:"auth_encryption_setting"`
	Size        int    `yaml:"size" mapstructure:"size"`
	Network     string `yaml:"network" mapstructure:"network"`
	Address     string `yaml:"address" mapstructure:"address"`
	Password    string `yaml:"password" mapstructure:"password"`
}

// CacheSetting is config struct for setting up caching for
// a redis server
type CacheSetting struct {
	Address  string `yaml:"address" mapstructure:"address"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
}

// FileSystemSetting is config struct for storing sessions
// in the file system
type FileSystemSetting struct {
	SessionAuth `yaml:"auth_encryption_setting" mapstructure:"auth_encryption_setting"`
	Dir         string `yaml:"dir" mapstructure:"dir"`
}

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

func GetSettings(envString string, settings interface{}, opts ...viper.DecoderConfigOption) error {
	var err error

	if reflect.ValueOf(settings).Type().Kind() != reflect.Ptr {
		return fmt.Errorf("settings parameter must be pointer")
	}

	filePath := os.Getenv(envString)
	v := viper.New()
	v.SetConfigFile(filePath)

	if err = v.ReadInConfig(); err != nil {
		return err
	}
	if err = v.Unmarshal(settings, opts...); err != nil {
		return err
	}

	return nil
}
