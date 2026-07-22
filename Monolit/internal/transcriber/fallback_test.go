package transcriber

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/transcriber/openrouter"
)

type transcriberStub struct {
	result models.TranscriptionResult
	err    error
	media  string
}

func (s *transcriberStub) Provider() string { return "stub" }

func (s *transcriberStub) Transcribe(_ context.Context, file models.File) (models.TranscriptionResult, error) {
	if file.Content != nil {
		content, readErr := io.ReadAll(file.Content)
		if readErr != nil {
			return models.TranscriptionResult{}, readErr
		}
		s.media = string(content)
	}
	return s.result, s.err
}

func TestFallbackTranscriberUsesSecondProviderOnlyForTemporaryOpenRouterError(t *testing.T) {
	primary := &transcriberStub{err: &openrouter.HTTPStatusError{StatusCode: 502, Message: "upstream unavailable"}}
	fallback := &transcriberStub{result: models.TranscriptionResult{Text: "готово"}}
	provider := withFallback(primary, fallback)

	result, err := provider.Transcribe(context.Background(), models.File{Content: io.NopCloser(strings.NewReader("audio"))})
	if err != nil || result.Text != "готово" {
		t.Fatalf("fallback result = %+v, %v", result, err)
	}
	if primary.media != "audio" || fallback.media != "audio" {
		t.Fatalf("media was not replayed to both providers: primary=%q fallback=%q", primary.media, fallback.media)
	}
}

func TestFallbackTranscriberDoesNotHidePermanentError(t *testing.T) {
	permanent := errors.New("unsupported media")
	primary := &transcriberStub{err: permanent}
	fallback := &transcriberStub{result: models.TranscriptionResult{Text: "must not run"}}

	_, err := withFallback(primary, fallback).Transcribe(context.Background(), models.File{Content: io.NopCloser(strings.NewReader("audio"))})
	if !errors.Is(err, permanent) {
		t.Fatalf("error = %v, want permanent error", err)
	}
	if fallback.media != "" {
		t.Fatal("fallback must not run for a permanent error")
	}
}
