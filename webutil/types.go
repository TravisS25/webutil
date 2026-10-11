package webutil

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type BaseServerSettings struct {
	Protocol string `json:"protocol" yaml:"protocol" mapstructure:"protocol"`
	Host     string `json:"host" yaml:"host" mapstructure:"host"`
	Port     int    `json:"port" yaml:"port" mapstructure:"port"`
	Path     string `json:"path" yaml:"path" mapstructure:"path"`
}

func (b BaseServerSettings) String() string {
	if b.Protocol == "" || b.Host == "" {
		return ""
	}

	if b.Port != 0 && b.Path != "" {
		return fmt.Sprintf(
			"%s://%s:%d/%s",
			b.Protocol,
			b.Host,
			b.Port,
			b.Path,
		)
	} else if b.Port != 0 {
		return fmt.Sprintf(
			"%s://%s:%d",
			b.Protocol,
			b.Host,
			b.Port,
		)
	} else if b.Path != "" {
		return fmt.Sprintf(
			"%s://%s/%s",
			b.Protocol,
			b.Host,
			b.Path,
		)
	}

	return fmt.Sprintf(
		"%s://%s",
		b.Protocol,
		b.Host,
	)
}

type BaseAuthSettings struct {
	Username string `json:"username" yaml:"username" mapstructure:"username"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	APIKey   string `json:"api_key" yaml:"api_key" mapstructure:"api_key"`
}

// DatabaseSettings is config struct to set up database connection
type DatabaseSettings struct {
	BaseAuthSettings   `yaml:",inline" mapstructure:",squash"`
	BaseServerSettings `yaml:",inline" mapstructure:",squash"`
}

func (d DatabaseSettings) String() string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s",
		d.Protocol,
		d.Username,
		d.Password,
		d.Host,
		d.Port,
		d.Path,
	)
}

func (d DatabaseSettings) DBName() string {
	return strings.Split(d.Path, "?")[0]
}

func (d *DatabaseSettings) SetDBName(name string) {
	pathArr := strings.Split(d.Path, "?")
	pathArr[0] = name + "?"
	d.Path = strings.Join(pathArr, "")
}

func (d *DatabaseSettings) ParseConnStr(connStr string) error {
	dbURL, err := url.Parse(connStr)
	if err != nil {
		return err
	}

	var port int

	if dbURL.Port() != "" {
		port, _ = strconv.Atoi(dbURL.Port())
	}

	d.Protocol = dbURL.Scheme
	d.Username = dbURL.User.Username()
	d.Password, _ = dbURL.User.Password()
	d.Host = dbURL.Host
	d.Port = port
	d.Path = dbURL.Path + dbURL.RawQuery

	return nil
}

// S3StorageSettings is setting for S3 backend
type S3StorageSettings struct {
	HTTPEndpoint    string `json:"http_endpoint" yaml:"http_endpoint" mapstructure:"http_endpoint"`
	EndPoint        string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
	AccessKeyID     string `json:"access_key_id" yaml:"access_key_id" mapstructure:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" yaml:"secret_access_key" mapstructure:"secret_access_key"`
	UseSSL          bool   `json:"use_ssl" yaml:"use_ssl" mapstructure:"use_ssl"`
}

type SessionAuthSettings struct {
	AuthKey    string `yaml:"auth_key" mapstructure:"auth_key"`
	EncryptKey string `yaml:"encrypt_key" mapstructure:"encrypt_key"`
}
