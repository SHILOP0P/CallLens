package hybrid

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/transcriber/diarizer"
	"calllens/monolit/internal/transcriber/openrouter"
)

const providerName = "openrouter-pyannote"

type speechToText interface {
	Transcribe(context.Context, models.File) (models.TranscriptionResult, error)
}

type speakerDiarizer interface {
	Diarize(context.Context, models.File) ([]diarizer.Turn, error)
}

type Transcriber struct {
	asr      speechToText
	diarizer speakerDiarizer
}

func New(apiKey, model, diarizerURL string) (*Transcriber, error) {
	asr, err := openrouter.NewWithTimestamps(apiKey, model)
	if err != nil {
		return nil, err
	}
	diarization, err := diarizer.New(diarizerURL)
	if err != nil {
		return nil, err
	}
	return &Transcriber{asr: asr, diarizer: diarization}, nil
}

func (t *Transcriber) Provider() string { return providerName }

func (t *Transcriber) Transcribe(ctx context.Context, file models.File) (models.TranscriptionResult, error) {
	if file.Content == nil {
		return models.TranscriptionResult{}, errors.New("empty media content")
	}
	content, err := io.ReadAll(file.Content)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("read media for hybrid transcription: %w", err)
	}
	if len(content) == 0 {
		return models.TranscriptionResult{}, errors.New("empty media content")
	}

	clone := func() models.File {
		copy := file
		copy.Content = io.NopCloser(bytes.NewReader(content))
		return copy
	}
	transcript, err := t.asr.Transcribe(ctx, clone())
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("transcribe with OpenRouter: %w", err)
	}
	turns, err := t.diarizer.Diarize(ctx, clone())
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("diarize with pyannote: %w", err)
	}
	segments := assignSpeakers(transcript.Segments, turns)
	if len(segments) == 0 {
		return models.TranscriptionResult{}, errors.New("OpenRouter response has no timestamped segments to align with speakers")
	}
	return models.TranscriptionResult{Text: dialogueText(segments), Segments: segments, Language: transcript.Language}, nil
}

func assignSpeakers(transcript []models.TranscriptionSegment, turns []diarizer.Turn) []models.TranscriptionSegment {
	labels := make(map[string]string)
	result := make([]models.TranscriptionSegment, 0, len(transcript))
	for _, segment := range transcript {
		if segment.StartSeconds == nil || segment.EndSeconds == nil || *segment.EndSeconds <= *segment.StartSeconds || strings.TrimSpace(segment.Text) == "" {
			continue
		}
		speaker := bestSpeaker(*segment.StartSeconds, *segment.EndSeconds, turns)
		if _, exists := labels[speaker]; !exists {
			labels[speaker] = fmt.Sprintf("Спикер %d", len(labels)+1)
		}
		segment.Speaker = labels[speaker]
		result = append(result, segment)
	}
	return result
}

func bestSpeaker(start, end float64, turns []diarizer.Turn) string {
	best, bestOverlap := "SPEAKER_UNKNOWN", 0.0
	middle := (start + end) / 2
	for _, turn := range turns {
		overlap := max(0, min(end, turn.EndSeconds)-max(start, turn.StartSeconds))
		if overlap > bestOverlap || (bestOverlap == 0 && turn.StartSeconds <= middle && middle <= turn.EndSeconds) {
			best, bestOverlap = turn.Speaker, overlap
		}
	}
	return best
}

func dialogueText(segments []models.TranscriptionSegment) string {
	lines := make([]string, 0, len(segments))
	for _, segment := range segments {
		if text := strings.TrimSpace(segment.Text); text != "" {
			lines = append(lines, segment.Speaker+": "+text)
		}
	}
	return strings.Join(lines, "\n")
}
