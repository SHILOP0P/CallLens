package transcriber

import (
	"fmt"
	"strings"

	localTranscriber "calllens/monolit/internal/transcriber/local"
	mockTranscriber "calllens/monolit/internal/transcriber/mock"
	openrouterTranscriber "calllens/monolit/internal/transcriber/openrouter"
)

type Config interface {
	Provider() string
	APIKey() string
	Model() string
	FallbackModel() string
	URL() string
	DiarizerURL() string
}

func NewFromConfig(cfg Config) (Transcriber, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider()))

	switch provider {
	case "", "mock":
		return mockTranscriber.New(), nil
	case "openrouter":
		return newOpenRouterWithFallback(cfg.APIKey(), cfg.Model(), cfg.FallbackModel())
	case "hybrid", "openrouter-pyannote":
		return newTieredTranscriber(cfg.APIKey(), cfg.Model(), cfg.FallbackModel(), cfg.DiarizerURL())
	case "local":
		return localTranscriber.New(cfg.URL())
	case "local-pyannote":
		return newLocalTieredTranscriber(cfg.URL(), cfg.DiarizerURL())
	case "openai":
		return nil, fmt.Errorf("openai transcriber is not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported transcriber provider: %s", cfg.Provider())
	}
}

func newOpenRouterWithFallback(apiKey, model, fallbackModel string) (Transcriber, error) {
	primary, err := openrouterTranscriber.New(apiKey, model)
	if err != nil {
		return nil, err
	}
	return newFallbackTranscriber(primary, apiKey, fallbackModel)
}
