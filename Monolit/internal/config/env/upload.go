package env

import "github.com/caarlos0/env/v11"

type uploadEnvConfig struct {
	path string `env:"UPLOAD_PATH,required"`
}

type uploadConfig struct {
	Raw uploadEnvConfig
}

func NewUploadConfig() (*uploadConfig, error) {
	var raw uploadEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &uploadConfig{Raw: raw}, nil
}

func (config *uploadConfig) Path() string {
	return config.Raw.path
}
