package env

import (
	"net"
	"time"

	"github.com/caarlos0/env/v11"
)

type httpEnvConfig struct {
	Host        string        `env:"HTTP_HOST,required"`
	Port        string        `env:"HTTP_PORT,required"`
	ReadTimeout time.Duration `env:"HTTP_READ_TIMEOUT,required"`
}

type httpConfig struct {
	raw httpEnvConfig
}

func NewHTTPConfig() (*httpConfig, error) {
	var raw httpEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &httpConfig{raw: raw}, nil
}

func (config *httpConfig) Address() string {
	return net.JoinHostPort(config.raw.Host, config.raw.Port)
}

func (config *httpConfig) ReadTimeout() time.Duration {
	return config.raw.ReadTimeout
}
