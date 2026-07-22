package diarizer

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
)

type Turn struct {
	StartSeconds float64 `json:"start_seconds"`
	EndSeconds   float64 `json:"end_seconds"`
	Speaker      string  `json:"speaker"`
}

type Client struct {
	endpoint string
	client   *http.Client
}

func New(baseURL string) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("diarizer url is required")
	}
	return &Client{endpoint: baseURL + "/v1/diarize", client: &http.Client{Timeout: 5 * time.Minute}}, nil
}

func (c *Client) Diarize(ctx context.Context, file models.File) ([]Turn, error) {
	if file.Content == nil {
		return nil, errors.New("empty media content")
	}
	reader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)
	filename := filepath.Base(strings.ReplaceAll(strings.TrimSpace(file.OriginalFilename), "\\", "/"))
	if filename == "" || filename == "." {
		filename = "call-media"
	}
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		defer pipeWriter.Close()
		defer writer.Close()
		part, err := writer.CreateFormFile("file", filename)
		if err == nil {
			_, err = io.Copy(part, file.Content)
		}
		errCh <- err
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, reader)
	if err != nil {
		return nil, fmt.Errorf("build diarization request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send diarization request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if copyErr := <-errCh; copyErr != nil {
		return nil, fmt.Errorf("copy diarization media: %w", copyErr)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("diarization failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(message)))
	}
	var result struct {
		Turns []Turn `json:"turns"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode diarization response: %w", err)
	}
	turns := make([]Turn, 0, len(result.Turns))
	for _, turn := range result.Turns {
		if turn.EndSeconds > turn.StartSeconds && strings.TrimSpace(turn.Speaker) != "" {
			turns = append(turns, turn)
		}
	}
	if len(turns) == 0 {
		return nil, errors.New("diarization response has no speaker turns")
	}
	return turns, nil
}
