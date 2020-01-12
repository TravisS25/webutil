package webutil

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// EmailSetting is config struct for email
type EmailSetting struct {
	BaseAuthSetting
}

// SessionSetting is config struct for setting up session AuthEncryptionSetting
// for redis server
type SessionSetting struct {
	AuthEncryptionSetting
	Size     int    `yaml:"size"`
	Network  string `yaml:"network"`
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
}

// CacheSetting is config struct for setting up caching for
// a redis server
type CacheSetting struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// CookieSetting is config struct for storing sessions
// in cookies
type CookieSetting struct {
	AuthEncryptionSetting
}

// FileSystemSetting is config struct for storing sessions
// in the file system
type FileSystemSetting struct {
	AuthEncryptionSetting
	Dir string `yaml:"dir"`
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
	EndPoint        string `yaml:"end_point"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
}

// Settings is the configuration settings for the app
type Settings struct {
	Prod           bool     `yaml:"prod"`
	Domain         string   `yaml:"domain"`
	ClientDomain   string   `yaml:"client_domain"`
	CSRF           string   `yaml:"csrf"`
	TemplatesDir   string   `yaml:"templates_dir"`
	HTTPS          bool     `yaml:"https"`
	AssetsLocation string   `yaml:"assets_location"`
	AllowedOrigins []string `yaml:"allowed_origins"`

	PaymentConfig  map[string]string            `yaml:"payment_config"`
	CacheConfig    map[string]CacheSetting      `yaml:"cache_config"`
	SessionConfig  map[string]SessionSetting    `yaml:"session_config"`
	S3Config       map[string]S3StorageSetting  `yaml:"s3_config"`
	DatabaseConfig map[string][]DatabaseSetting `yaml:"database_config"`
	EmailConfig    map[string]EmailSetting      `yaml:"email_config"`
}

// ConfigSettings simply takes a string which should reference an enviroment variable
// that points to config file used for application
func ConfigSettings(envString string) (*Settings, error) {
	var settings *Settings
	configFile := os.Getenv(envString)
	source, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(source, &settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

type AuthEncryptionSetting struct {
	AuthKey    string `yaml:"auth_key"`
	EncryptKey string `yaml:"encrypt_key"`
}

type BaseAuthSetting struct {
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
}
