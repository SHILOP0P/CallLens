package transcriber

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/transcriber/hybrid"
	"calllens/monolit/internal/transcriber/openrouter"
)

// fallbackTranscriber retries temporary OpenRouter failures with a separately
// configured model. The media is buffered once because each provider consumes
// the input stream.
type fallbackTranscriber struct {
	primary  Transcriber
	fallback Transcriber
}

func newFallbackTranscriber(primary Transcriber, apiKey, fallbackModel string) (Transcriber, error) {
	fallbackModel = strings.TrimSpace(fallbackModel)
	if fallbackModel == "" {
		return primary, nil
	}
	fallback, err := openrouter.New(apiKey, fallbackModel)
	if err != nil {
		return nil, err
	}
	return withFallback(primary, fallback), nil
}

func newHybridWithFallback(primary Transcriber, apiKey, fallbackModel, diarizerURL string) (Transcriber, error) {
	fallbackModel = strings.TrimSpace(fallbackModel)
	if fallbackModel == "" {
		return primary, nil
	}
	fallback, err := hybrid.New(apiKey, fallbackModel, diarizerURL)
	if err != nil {
		return nil, err
	}
	return withFallback(primary, fallback), nil
}

func withFallback(primary, fallback Transcriber) Transcriber {
	return &fallbackTranscriber{primary: primary, fallback: fallback}
}

func (t *fallbackTranscriber) Provider() string { return t.primary.Provider() }

func (t *fallbackTranscriber) Transcribe(ctx context.Context, file models.File) (models.TranscriptionResult, error) {
	if t.fallback == nil || file.Content == nil {
		return t.primary.Transcribe(ctx, file)
	}
	content, err := io.ReadAll(file.Content)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("buffer media for transcription fallback: %w", err)
	}
	clone := func() models.File {
		copy := file
		copy.Content = io.NopCloser(bytes.NewReader(content))
		return copy
	}
	result, err := t.primary.Transcribe(ctx, clone())
	if err == nil || !openrouter.IsTemporaryError(err) {
		return result, err
	}
	return t.fallback.Transcribe(ctx, clone())
}
