package env

import "github.com/caarlos0/env/v11"

type analyzerEnvConfig struct {
	Provider string `env:"ANALYZER_PROVIDER" envDefault:"mock"`
	APIKey   string `env:"ANALYZER_API_KEY"`
	Model    string `env:"ANALYZER_MODEL"`
}

type analyzerConfig struct {
	raw analyzerEnvConfig
}

func NewAnalyzerConfig() (*analyzerConfig, error) {
	var raw analyzerEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &analyzerConfig{raw: raw}, nil
}

func (cfg *analyzerConfig) Provider() string {
	return cfg.raw.Provider
}

func (cfg *analyzerConfig) APIKey() string {
	return cfg.raw.APIKey
}

func (cfg *analyzerConfig) Model() string {
	return cfg.raw.Model
}
