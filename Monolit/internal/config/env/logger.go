package env

import "github.com/caarlos0/env/v11"

type loggerEnvConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	AsJSON bool   `env:"LOG_AS_JSON" envDefault:"false"`
}

type loggerConfig struct {
	raw loggerEnvConfig
}

func NewLoggerConfig() (*loggerConfig, error) {
	var raw loggerEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &loggerConfig{raw: raw}, nil
}

func (cfg *loggerConfig) Level() string {
	return cfg.raw.Level
}

func (cfg *loggerConfig) AsJSON() bool {
	return cfg.raw.AsJSON
}
