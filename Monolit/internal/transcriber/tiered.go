package transcriber

import (
	"context"
	"fmt"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/transcriber/diarizer"
	"calllens/monolit/internal/transcriber/hybrid"
	localTranscriber "calllens/monolit/internal/transcriber/local"
	"calllens/monolit/internal/transcriber/openrouter"
)

type tieredTranscriber struct {
	standard Transcriber
	diarized Transcriber
}

func newLocalTieredTranscriber(transcriberURL string, diarizerURL string) (Transcriber, error) {
	standard, err := localTranscriber.New(transcriberURL)
	if err != nil {
		return nil, err
	}
	diarization, err := diarizer.New(diarizerURL)
	if err != nil {
		return nil, err
	}
	return &tieredTranscriber{
		standard: standard,
		diarized: hybrid.NewWithDependencies(standard, diarization),
	}, nil
}

func newTieredTranscriber(apiKey, model, fallbackModel, diarizerURL string) (Transcriber, error) {
	standardPrimary, err := openrouter.New(apiKey, model)
	if err != nil {
		return nil, err
	}
	standard, err := newFallbackTranscriber(standardPrimary, apiKey, fallbackModel)
	if err != nil {
		return nil, err
	}
	diarizedPrimary, err := hybrid.New(apiKey, model, diarizerURL)
	if err != nil {
		return nil, err
	}
	diarized, err := newHybridWithFallback(diarizedPrimary, apiKey, fallbackModel, diarizerURL)
	if err != nil {
		return nil, err
	}
	return &tieredTranscriber{standard: standard, diarized: diarized}, nil
}

func (t *tieredTranscriber) Provider() string { return t.standard.Provider() }

func (t *tieredTranscriber) Transcribe(ctx context.Context, file models.File) (models.TranscriptionResult, error) {
	return t.standard.Transcribe(ctx, file)
}

func (t *tieredTranscriber) ProviderForMode(mode models.TranscriptionMode) string {
	if mode == models.TranscriptionModeDiarized {
		return t.diarized.Provider()
	}
	return t.standard.Provider()
}

func (t *tieredTranscriber) TranscribeForMode(ctx context.Context, file models.File, mode models.TranscriptionMode) (models.TranscriptionResult, error) {
	switch mode {
	case "", models.TranscriptionModeStandard:
		return t.standard.Transcribe(ctx, file)
	case models.TranscriptionModeDiarized:
		return t.diarized.Transcribe(ctx, file)
	default:
		return models.TranscriptionResult{}, fmt.Errorf("unsupported transcription mode: %s", mode)
	}
}
