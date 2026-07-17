package hybrid

import (
	"context"
	"io"
	"strings"
	"testing"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/transcriber/diarizer"
)

type asrStub struct{ result models.TranscriptionResult }

func (s asrStub) Transcribe(_ context.Context, file models.File) (models.TranscriptionResult, error) {
	_, _ = io.ReadAll(file.Content)
	return s.result, nil
}

type diarizerStub struct{ turns []diarizer.Turn }

func (s diarizerStub) Diarize(_ context.Context, file models.File) ([]diarizer.Turn, error) {
	_, _ = io.ReadAll(file.Content)
	return s.turns, nil
}

func TestTranscribeBuildsTimestampedDialogue(t *testing.T) {
	startA, endA := 0.0, 1.2
	startB, endB := 1.2, 2.5
	language := "ru"
	transcriber := &Transcriber{
		asr: asrStub{result: models.TranscriptionResult{Language: &language, Segments: []models.TranscriptionSegment{
			{StartSeconds: &startA, EndSeconds: &endA, Text: "Добрый день."},
			{StartSeconds: &startB, EndSeconds: &endB, Text: "Здравствуйте."},
		}}},
		diarizer: diarizerStub{turns: []diarizer.Turn{
			{StartSeconds: 0, EndSeconds: 1.2, Speaker: "SPEAKER_01"},
			{StartSeconds: 1.2, EndSeconds: 3, Speaker: "SPEAKER_07"},
		}},
	}

	result, err := transcriber.Transcribe(context.Background(), models.File{
		Content:          io.NopCloser(strings.NewReader("media")),
		OriginalFilename: "call.mp4",
		MimeType:         "video/mp4",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Segments) != 2 || result.Segments[0].Speaker != "Спикер 1" || result.Segments[1].Speaker != "Спикер 2" {
		t.Fatalf("segments = %#v", result.Segments)
	}
	if result.Text != "Спикер 1: Добрый день.\nСпикер 2: Здравствуйте." {
		t.Fatalf("text = %q", result.Text)
	}
}

func TestTranscribeRejectsUntimestampedOpenRouterResult(t *testing.T) {
	transcriber := &Transcriber{
		asr:      asrStub{result: models.TranscriptionResult{Text: "Без временных меток"}},
		diarizer: diarizerStub{turns: []diarizer.Turn{{StartSeconds: 0, EndSeconds: 1, Speaker: "SPEAKER_00"}}},
	}
	_, err := transcriber.Transcribe(context.Background(), models.File{Content: io.NopCloser(strings.NewReader("media"))})
	if err == nil || !strings.Contains(err.Error(), "timestamped segments") {
		t.Fatalf("error = %v", err)
	}
}
