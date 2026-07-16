package local

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/transcriber/cleaner"
)

const providerName = "local-pyannote"

type Transcriber struct {
	endpoint string
	client   *http.Client
}

type transcriptionResponse struct {
	Text     string                 `json:"text"`
	Language *string                `json:"language"`
	Segments []transcriptionSegment `json:"segments"`
}

type transcriptionSegment struct {
	Speaker      string  `json:"speaker"`
	StartSeconds float64 `json:"start_seconds"`
	EndSeconds   float64 `json:"end_seconds"`
	Text         string  `json:"text"`
}

func New(rawURL string) (*Transcriber, error) {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("local transcriber URL is invalid")
	}
	return &Transcriber{
		endpoint: strings.TrimRight(rawURL, "/") + "/v1/transcribe",
		client:   &http.Client{Timeout: 30 * time.Minute},
	}, nil
}

func (t *Transcriber) Provider() string { return providerName }

func (t *Transcriber) Transcribe(ctx context.Context, file models.File) (models.TranscriptionResult, error) {
	if file.Content == nil {
		return models.TranscriptionResult{}, fmt.Errorf("%w: empty media content", models.ErrUnsupportedAudioType)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", safeFilename(file.OriginalFilename))
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("create local transcription upload: %w", err)
	}
	if _, err = io.Copy(part, file.Content); err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("copy local transcription upload: %w", err)
	}
	if err = writer.Close(); err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("close local transcription upload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, &body)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("build local transcription request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := t.client.Do(req)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("send local transcription request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return models.TranscriptionResult{}, fmt.Errorf("local transcription failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(message)))
	}

	var result transcriptionResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("decode local transcription response: %w", err)
	}
	segments := make([]models.TranscriptionSegment, 0, len(result.Segments))
	for _, segment := range result.Segments {
		text := cleaner.Clean(segment.Text)
		if text == "" || segment.EndSeconds <= segment.StartSeconds {
			continue
		}
		start, end := segment.StartSeconds, segment.EndSeconds
		segments = append(segments, models.TranscriptionSegment{
			Speaker: strings.TrimSpace(segment.Speaker), StartSeconds: &start, EndSeconds: &end, Text: text,
		})
	}
	result.Text = strings.TrimSpace(result.Text)
	if result.Text == "" || len(segments) == 0 {
		return models.TranscriptionResult{}, errors.New("local transcription response has no diarized segments")
	}
	return models.TranscriptionResult{Text: result.Text, Segments: segments, Language: result.Language}, nil
}

func safeFilename(filename string) string {
	filename = filepath.Base(strings.ReplaceAll(strings.TrimSpace(filename), "\\", "/"))
	if filename == "" || filename == "." {
		return "call-media"
	}
	return filename
}
