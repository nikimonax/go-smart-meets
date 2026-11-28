package config

import (
	"fmt"
	"net/url"

	"github.com/nikimonax/go-smart-meets/internal/server/global"
)

var _ fmt.Stringer = (*ListenConfig)(nil)

type ListenConfig struct {
	Host string `validate:"required,hostname_rfc1123|ip"`
	Port int    `validate:"required,port"`
}

// String implements [fmt.Stringer].
func (cfg *ListenConfig) String() string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

type HTTPConfig struct {
	Listen ListenConfig
}

type DatabaseConfig struct {
	Host    string `validate:"required,hostname_rfc1123|ip"`
	Port    int    `validate:"required,port"`
	User    string
	Pass    string
	Name    string `validate:"required"`
	Migrate bool
}

func (cfg *DatabaseConfig) URL() *url.URL {
	var user *url.Userinfo

	if cfg.User != "" {
		if cfg.Pass != "" {
			user = url.UserPassword(cfg.User, cfg.Pass)
		} else {
			user = url.User(cfg.User)
		}
	}

	return &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		User:   user,
		Path:   cfg.Name,
	}

}

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
}

func (cfg *Config) Validate() error {
	return global.Validate.Struct(cfg)
}
