package env

import "github.com/caarlos0/env/v11"

type transcriberEnvConfig struct {
	Provider      string `env:"TRANSCRIBER_PROVIDER" envDefault:"mock"`
	APIKey        string `env:"TRANSCRIBER_API_KEY"`
	Model         string `env:"TRANSCRIBER_MODEL"`
	FallbackModel string `env:"TRANSCRIBER_FALLBACK_MODEL"`
	URL           string `env:"TRANSCRIBER_URL" envDefault:"http://localhost:8090"`
	DiarizerURL   string `env:"DIARIZER_URL" envDefault:"http://localhost:8090"`
}

type transcriberConfig struct {
	raw transcriberEnvConfig
}

func NewTranscriberConfig() (*transcriberConfig, error) {
	var raw transcriberEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &transcriberConfig{raw: raw}, nil
}

func (cfg *transcriberConfig) Provider() string {
	return cfg.raw.Provider
}

func (cfg *transcriberConfig) APIKey() string {
	return cfg.raw.APIKey
}

func (cfg *transcriberConfig) Model() string {
	return cfg.raw.Model
}

func (cfg *transcriberConfig) FallbackModel() string {
	return cfg.raw.FallbackModel
}

func (cfg *transcriberConfig) URL() string {
	return cfg.raw.URL
}

func (cfg *transcriberConfig) DiarizerURL() string {
	return cfg.raw.DiarizerURL
}
