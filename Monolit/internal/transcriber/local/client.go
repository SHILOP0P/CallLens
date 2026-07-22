package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/transcriber/cleaner"
)

const providerName = "local-faster-whisper"

type Transcriber struct {
	endpoint string
	client   *http.Client
}

func New(baseURL string) (*Transcriber, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("local transcriber url is required")
	}
	return &Transcriber{
		endpoint: baseURL + "/v1/audio/transcriptions",
		client:   &http.Client{Timeout: 45 * time.Minute},
	}, nil
}

func (t *Transcriber) Provider() string { return providerName }

func (t *Transcriber) Transcribe(ctx context.Context, file models.File) (models.TranscriptionResult, error) {
	if file.Content == nil {
		return models.TranscriptionResult{}, errors.New("empty media content")
	}

	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		defer writer.Close()
		defer multipartWriter.Close()
		filename := filepath.Base(strings.TrimSpace(file.OriginalFilename))
		if filename == "." || filename == "" {
			filename = "call-media"
		}
		part, err := multipartWriter.CreateFormFile("file", filename)
		if err == nil {
			_, err = io.Copy(part, file.Content)
		}
		errCh <- err
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, reader)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("build local transcription request: %w", err)
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	resp, err := t.client.Do(req)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("send local transcription request: %w", err)
	}
	defer resp.Body.Close()
	if copyErr := <-errCh; copyErr != nil {
		return models.TranscriptionResult{}, fmt.Errorf("copy media to local transcriber: %w", copyErr)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return models.TranscriptionResult{}, fmt.Errorf("local transcription failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(message)))
	}

	var response struct {
		Text     string                        `json:"text"`
		Language string                        `json:"language"`
		Segments []models.TranscriptionSegment `json:"segments"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("decode local transcription response: %w", err)
	}
	response.Text = cleaner.Clean(response.Text)
	if response.Text == "" {
		return models.TranscriptionResult{}, errors.New("local transcription response is empty")
	}
	language := strings.TrimSpace(response.Language)
	if language == "" {
		language = "ru"
	}
	return models.TranscriptionResult{Text: response.Text, Segments: response.Segments, Language: &language}, nil
}
