package webutil

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// EmailSetting is config struct for email
type EmailSetting struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// RedisSessionSetting is config struct for setting up session store
// for redis server
type RedisSessionSetting struct {
	Size       int    `yaml:"size"`
	Network    string `yaml:"network"`
	Address    string `yaml:"address"`
	Password   string `yaml:"password"`
	AuthKey    string `yaml:"auth_key"`
	EncryptKey string `yaml:"encrypt_key"`
}

// RedisCacheSetting is config struct for setting up caching for
// a redis server
type RedisCacheSetting struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// CookieStoreSetting is config struct for storing sessions
// in cookies
type CookieStoreSetting struct {
	AuthKey    string `yaml:"auth_key"`
	EncryptKey string `yaml:"encrypt_key"`
}

// FileSystemStoreSetting is config struct for storing sessions
// in the file system
type FileSystemStoreSetting struct {
	Dir        string `yaml:"dir"`
	AuthKey    string `yaml:"auth_key"`
	EncryptKey string `yaml:"encrypt_key"`
}

type CacheSetting struct {
	Redis *RedisCacheSetting `yaml:"redis"`
}

// StripeSetting is config struct to set up stripe in app
type StripeSetting struct {
	TestMode            bool   `yaml:"test_mode"`
	StripeTestSecretKey string `yaml:"stripe_test_secret_key"`
	StripeLiveSecretKey string `yaml:"stripe_live_secret_key"`
}

// DatabaseSetting is config struct to set up database connection
type DatabaseSetting struct {
	DBName   string `yaml:"db_name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	SSLMode  string `yaml:"ssl_mode"`
}

type S3Config map[string]S3StorageSetting

// S3StorageSetting
type S3StorageSetting struct {
	EndPoint        string `yaml:"end_point"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
}

// Settings is the configuration settings for the app
type Settings struct {
	Prod           bool          `yaml:"prod"`
	Domain         string        `yaml:"domain"`
	ClientDomain   string        `yaml:"client_domain"`
	CSRF           string        `yaml:"csrf"`
	TemplatesDir   string        `yaml:"templates_dir"`
	HTTPS          bool          `yaml:"https"`
	AssetsLocation string        `yaml:"assets_location"`
	AllowedOrigins []string      `yaml:"allowed_origins"`
	Cache          CacheSetting  `yaml:"cache"`
	Stripe         StripeSetting `yaml:"stripe"`
	S3Config       S3Config      `yaml:"s3_config"`

	Databases map[string][]DatabaseSetting `yaml:"databases"`
	Emails    map[string]EmailSetting      `yaml:"emails"`
	StripeMap map[string]StripeSetting     `yaml:"stripe_map"`
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
