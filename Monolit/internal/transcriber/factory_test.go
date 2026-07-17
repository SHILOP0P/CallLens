package transcriber

import (
	"context"
	"strings"
	"testing"

	"calllens/monolit/internal/models"
)

type testConfig struct{ provider, apiKey, model, url string }

func (c testConfig) Provider() string { return c.provider }
func (c testConfig) APIKey() string   { return c.apiKey }
func (c testConfig) Model() string    { return c.model }
func (c testConfig) URL() string      { return c.url }

func TestNewFromConfigAndMockTranscriber(t *testing.T) {
	for _, provider := range []string{"", "mock", " MOCK "} {
		got, err := NewFromConfig(testConfig{provider: provider})
		if err != nil || got.Provider() != "mock" {
			t.Fatalf("provider %q: transcriber=%v err=%v", provider, got, err)
		}
		result, err := got.Transcribe(context.Background(), models.File{OriginalFilename: "call.wav"})
		if err != nil || !strings.Contains(result.Text, "call.wav") {
			t.Fatalf("unexpected result: %+v err=%v", result, err)
		}
	}

	if _, err := NewFromConfig(testConfig{provider: "openai"}); err == nil {
		t.Fatal("expected not implemented error")
	}
	if _, err := NewFromConfig(testConfig{provider: "unknown"}); err == nil {
		t.Fatal("expected unsupported provider error")
	}
	if _, err := NewFromConfig(testConfig{provider: "openrouter"}); err == nil {
		t.Fatal("expected invalid OpenRouter config error")
	}

	got, _ := NewFromConfig(testConfig{provider: "mock"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := got.Transcribe(ctx, models.File{}); err == nil {
		t.Fatal("expected canceled context error")
	}
	hybrid, err := NewFromConfig(testConfig{provider: "hybrid", apiKey: "key", model: "openai/whisper-large-v3", url: "http://localhost:8090"})
	if err != nil || hybrid.Provider() != "openrouter" {
		t.Fatalf("hybrid provider = %T, %v", hybrid, err)
	}
}
