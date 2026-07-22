package hybrid

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// NewWithDependencies composes a timestamp-capable ASR with a diarizer.
func NewWithDependencies(asr speechToText, diarization speakerDiarizer) *Transcriber {
	return &Transcriber{asr: asr, diarizer: diarization}
}

func (t *Transcriber) Provider() string { return providerName }

func (t *Transcriber) Transcribe(ctx context.Context, file models.File) (models.TranscriptionResult, error) {
	if file.Content == nil {
		return models.TranscriptionResult{}, errors.New("empty media content")
	}
	temp, err := os.CreateTemp("", "calllens-hybrid-*")
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("create temporary media for hybrid transcription: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	written, copyErr := io.Copy(temp, file.Content)
	closeErr := temp.Close()
	if copyErr != nil {
		return models.TranscriptionResult{}, fmt.Errorf("copy media for hybrid transcription: %w", copyErr)
	}
	if closeErr != nil {
		return models.TranscriptionResult{}, fmt.Errorf("close temporary media for hybrid transcription: %w", closeErr)
	}
	if written == 0 {
		return models.TranscriptionResult{}, errors.New("empty media content")
	}

	clone := func() (models.File, error) {
		content, openErr := os.Open(tempPath)
		if openErr != nil {
			return models.File{}, openErr
		}
		copy := file
		copy.Content = content
		copy.Path = filepath.Base(tempPath)
		return copy, nil
	}
	asrFile, err := clone()
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("open media for transcription: %w", err)
	}
	transcript, err := t.asr.Transcribe(ctx, asrFile)
	_ = asrFile.Content.Close()
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("transcribe with OpenRouter: %w", err)
	}
	diarizerFile, err := clone()
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("open media for diarization: %w", err)
	}
	turns, err := t.diarizer.Diarize(ctx, diarizerFile)
	_ = diarizerFile.Content.Close()
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
