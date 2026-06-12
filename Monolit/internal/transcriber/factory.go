package transcriber

import (
	mockTranscriber "calllens/monolit/internal/transcriber/mock"
	openrouterTranscriber "calllens/monolit/internal/transcriber/openrouter"
	"fmt"
	"strings"
)

type Config interface {
	Provider() string
	APIKey() string
	Model() string
}

func NewFromConfig(cfg Config) (Transcriber, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider()))

	switch provider {
	case "", "mock":
		return mockTranscriber.New(), nil
	case "openrouter":
		return openrouterTranscriber.New(cfg.APIKey(), cfg.Model())
	case "openai":
		return nil, fmt.Errorf("openai transcriber is not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported transcriber provider: %s", cfg.Provider())
	}
}
