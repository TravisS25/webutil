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
	AuthEncryptionSetting `yaml:"auth_encryption_setting" mapstructure:"auth_encryption_setting"`
	Size                  int    `yaml:"size" mapstructure:"size"`
	Network               string `yaml:"network" mapstructure:"network"`
	Address               string `yaml:"address" mapstructure:"address"`
	Password              string `yaml:"password" mapstructure:"password"`
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
	AuthEncryptionSetting `yaml:"auth_encryption_setting" mapstructure:"auth_encryption_setting"`
	Dir                   string `yaml:"dir" mapstructure:"dir"`
}

// DatabaseSetting is config struct to set up database connection
type DatabaseSetting struct {
	BaseAuthSetting `yaml:"base_auth_setting" mapstructure:"base_auth_setting"`
	DBName          string `yaml:"db_name" mapstructure:"db_name"`
	SSLMode         string `yaml:"ssl_mode" mapstructure:"ssl_mode"`
	SSL             bool   `yaml:"ssl" mapstructure:"ssl"`
	SSLRootCert     string `yaml:"ssl_root_cert" mapstructure:"ssl_root_cert"`
	SSLKey          string `yaml:"ssl_key" mapstructure:"ssl_key"`
	SSLCert         string `yaml:"ssl_cert" mapstructure:"ssl_cert"`
}

// S3StorageSetting is setting for S3 backend
type S3StorageSetting struct {
	EndPoint        string `yaml:"end_point" mapstructure:"end_point"`
	AccessKeyID     string `yaml:"access_key_id" mapstructure:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key" mapstructure:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl" mapstructure:"use_ssl"`
}

// AuthEncryptionSetting is config struct for other config
// structs that require encryption
type AuthEncryptionSetting struct {
	AuthKey    string `yaml:"auth_key" mapstructure:"auth_key"`
	EncryptKey string `yaml:"encrypt_key" mapstructure:"encrypt_key"`
}

// BaseAuthSetting is config struct for other config structs
// for basic authentication
type BaseAuthSetting struct {
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
}

// Settings is the configuration settings for the app
type Settings struct {
	Prod           bool     `yaml:"prod" mapstructure:"prod"`
	Domain         string   `yaml:"domain" mapstructure:"domain"`
	ClientDomain   string   `yaml:"client_domain" mapstructure:"client_domain"`
	CSRF           string   `yaml:"csrf" mapstructure:"csrf"`
	TemplatesDir   string   `yaml:"templates_dir" mapstructure:"templates_dir"`
	HTTPS          bool     `yaml:"https" mapstructure:"https"`
	AssetsLocation string   `yaml:"assets_location" mapstructure:"assets_location"`
	AllowedOrigins []string `yaml:"allowed_origins" mapstructure:"allowed_origins"`
	RootDir        string   `yaml:"root_dir" mapstructure:"root_dir"`

	PaymentConfig  map[string]string            `yaml:"payment_config" mapstructure:"payment_config"`
	CacheConfig    map[string]CacheSetting      `yaml:"cache_config" mapstructure:"cache_config"`
	SessionConfig  map[string]SessionSetting    `yaml:"session_config" mapstructure:"session_config"`
	S3Config       map[string]S3StorageSetting  `yaml:"s3_config" mapstructure:"s3_config"`
	DatabaseConfig map[string][]DatabaseSetting `yaml:"database_config" mapstructure:"database_config"`
	EmailConfig    map[string]BaseAuthSetting   `yaml:"email_config" mapstructure:"email_config"`
}

// ConfigSettings simply takes a string which should reference an enviroment variable
// that points to config file used for application
func ConfigSettings(envString string) (*Settings, error) {
	var settings *Settings
	var err error

	filePath := os.Getenv(envString)
	v := viper.New()
	v.SetConfigFile(filePath)

	if err = v.ReadInConfig(); err != nil {
		panic(err.Error())
	}
	if err = v.Unmarshal(&settings); err != nil {
		panic(err.Error())
	}

	return settings, nil
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
