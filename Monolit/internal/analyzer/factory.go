package analyzer

import (
	mockAnalyzer "calllens/monolit/internal/analyzer/mock"
	"fmt"
	"strings"
)

type Config interface {
	Provider() string
	APIKey() string
	Model() string
}

func NewFromConfig(cfg Config) (Analyzer, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider()))

	switch provider {
	case "", "mock":
		return mockAnalyzer.New(cfg.Model()), nil
	case "openai":
		return nil, fmt.Errorf("openai analyzer is not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported analyzer provider: %s", cfg.Provider())
	}
}
