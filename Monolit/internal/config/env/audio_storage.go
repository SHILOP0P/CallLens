package env

import "github.com/caarlos0/env/v11"

type uploadEnvConfig struct {
	Path string `env:"UPLOAD_PATH,required"`
}

type uploadConfig struct {
	raw uploadEnvConfig
}

func NewUploadConfig() (*uploadConfig, error) {
	var raw uploadEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &uploadConfig{raw: raw}, nil
}

func (config *uploadConfig) Path() string {
	return config.raw.Path
}
