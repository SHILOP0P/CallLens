package openrouter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/transcriber/cleaner"
)

const (
	defaultBaseURL  = "https://openrouter.ai/api/v1"
	transcribePath  = "/audio/transcriptions"
	providerName    = "openrouter"
	defaultLanguage = "ru"
)

type Transcriber struct {
	apiKey   string
	model    string
	baseURL  string
	language string
	client   *http.Client
}

type transcriptionRequest struct {
	Model       string     `json:"model"`
	InputAudio  inputAudio `json:"input_audio"`
	Language    string     `json:"language,omitempty"`
	Temperature *float64   `json:"temperature,omitempty"`
}

type inputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type transcriptionResponse struct {
	Text     string                 `json:"text"`
	Segments []transcriptionSegment `json:"segments"`
}

type transcriptionSegment struct {
	Speaker      string   `json:"speaker"`
	Start        *float64 `json:"start"`
	End          *float64 `json:"end"`
	StartSeconds *float64 `json:"start_seconds"`
	EndSeconds   *float64 `json:"end_seconds"`
	Text         string   `json:"text"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    any    `json:"code"`
	} `json:"error"`
}

func New(apiKey string, model string) (*Transcriber, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, errors.New("openrouter transcriber api key is required")
	}

	model = strings.TrimSpace(model)
	if model == "" {
		return nil, errors.New("openrouter transcriber model is required")
	}

	return &Transcriber{
		apiKey:   apiKey,
		model:    model,
		baseURL:  defaultBaseURL,
		language: defaultLanguage,
		client: &http.Client{
			Timeout: 90 * time.Second,
		},
	}, nil
}

func (t *Transcriber) Provider() string {
	return providerName
}

func (t *Transcriber) Transcribe(ctx context.Context, file models.File) (models.TranscriptionResult, error) {
	if file.Content == nil {
		return models.TranscriptionResult{}, fmt.Errorf("%w: empty audio content", models.ErrUnsupportedAudioType)
	}

	format, err := audioFormat(file)
	if err != nil {
		return models.TranscriptionResult{}, err
	}

	audio, err := io.ReadAll(file.Content)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("read audio content: %w", err)
	}
	if len(audio) == 0 {
		return models.TranscriptionResult{}, fmt.Errorf("%w: empty audio content", models.ErrUnsupportedAudioType)
	}

	temperature := 0.0
	payload := transcriptionRequest{
		Model: t.model,
		InputAudio: inputAudio{
			Data:   base64.StdEncoding.EncodeToString(audio),
			Format: format,
		},
		Language:    t.language,
		Temperature: &temperature,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("marshal openrouter transcription request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint(), bytes.NewReader(requestBody))
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("build openrouter transcription request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("send openrouter transcription request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return models.TranscriptionResult{}, decodeError(resp)
	}

	var result transcriptionResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.TranscriptionResult{}, fmt.Errorf("decode openrouter transcription response: %w", err)
	}

	segments := normalizeSegments(result.Segments)
	result.Text = cleaner.Clean(result.Text)
	if result.Text == "" && len(segments) > 0 {
		result.Text = textFromSegments(segments)
	}
	if result.Text == "" {
		return models.TranscriptionResult{}, errors.New("openrouter transcription response is empty")
	}

	language := t.language
	return models.TranscriptionResult{
		Text:     result.Text,
		Segments: segments,
		Language: &language,
	}, nil
}

func (t *Transcriber) endpoint() string {
	return strings.TrimRight(t.baseURL, "/") + transcribePath
}

func audioFormat(file models.File) (string, error) {
	mimeType := strings.ToLower(strings.TrimSpace(strings.Split(file.MimeType, ";")[0]))
	switch mimeType {
	case "audio/mpeg", "audio/mp3":
		return "mp3", nil
	case "audio/wav", "audio/x-wav", "audio/wave", "audio/vnd.wave":
		return "wav", nil
	case "audio/mp4", "audio/x-m4a":
		return "m4a", nil
	case "audio/flac":
		return "flac", nil
	case "audio/ogg":
		return "ogg", nil
	case "audio/webm":
		return "webm", nil
	case "audio/aac":
		return "aac", nil
	}

	if format := supportedFormatFromExt(file.OriginalFilename); format != "" {
		return format, nil
	}
	if format := supportedFormatFromExt(file.Path); format != "" {
		return format, nil
	}

	return "", fmt.Errorf("%w: unsupported transcription audio format %q", models.ErrUnsupportedAudioType, file.MimeType)
}

func supportedFormatFromExt(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mp3":
		return "mp3"
	case ".wav":
		return "wav"
	case ".m4a":
		return "m4a"
	case ".flac":
		return "flac"
	case ".ogg":
		return "ogg"
	case ".webm":
		return "webm"
	case ".aac":
		return "aac"
	default:
		return ""
	}
}

func normalizeSegments(segments []transcriptionSegment) []models.TranscriptionSegment {
	result := make([]models.TranscriptionSegment, 0, len(segments))
	for _, segment := range segments {
		text := cleaner.Clean(segment.Text)
		if text == "" {
			continue
		}

		start := segment.StartSeconds
		if start == nil {
			start = segment.Start
		}
		end := segment.EndSeconds
		if end == nil {
			end = segment.End
		}

		result = append(result, models.TranscriptionSegment{
			Speaker:      strings.TrimSpace(segment.Speaker),
			StartSeconds: start,
			EndSeconds:   end,
			Text:         text,
		})
	}

	return result
}

func textFromSegments(segments []models.TranscriptionSegment) string {
	lines := make([]string, 0, len(segments))
	for _, segment := range segments {
		text := strings.TrimSpace(segment.Text)
		if text == "" {
			continue
		}
		if speaker := strings.TrimSpace(segment.Speaker); speaker != "" {
			lines = append(lines, speaker+": "+text)
			continue
		}
		lines = append(lines, text)
	}

	return strings.Join(lines, "\n")
}

func decodeError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("openrouter transcription failed with status %d: read error response: %w", resp.StatusCode, err)
	}

	message := strings.TrimSpace(string(body))
	var apiErr errorResponse
	if err = json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		message = apiErr.Error.Message
		if apiErr.Error.Code != nil {
			message = fmt.Sprintf("%s (code: %v)", message, apiErr.Error.Code)
		}
	}

	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}

	return fmt.Errorf("openrouter transcription failed with status %d: %s", resp.StatusCode, message)
}
