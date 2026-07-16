package local

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"calllens/monolit/internal/models"
)

func TestTranscribeReturnsDiarizedSegments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/transcribe" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = file.Close() }()
		content, _ := io.ReadAll(file)
		if header.Filename != "meeting.mp4" || string(content) != "media" {
			t.Fatalf("upload = %q %q", header.Filename, content)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"text": "Спикер 1: Добрый день.\nСпикер 2: Здравствуйте.", "language": "ru",
			"segments": []map[string]any{
				{"speaker": "Спикер 1", "start_seconds": 0.1, "end_seconds": 1.2, "text": "Добрый день."},
				{"speaker": "Спикер 2", "start_seconds": 1.3, "end_seconds": 2.1, "text": "Здравствуйте."},
			},
		})
	}))
	defer server.Close()

	transcriber, err := New(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	result, err := transcriber.Transcribe(context.Background(), models.File{
		Content: io.NopCloser(strings.NewReader("media")), OriginalFilename: "meeting.mp4", MimeType: "video/mp4",
	})
	if err != nil {
		t.Fatal(err)
	}
	if transcriber.Provider() != providerName || len(result.Segments) != 2 || result.Segments[1].Speaker != "Спикер 2" {
		t.Fatalf("result = %+v", result)
	}
}

func TestValidationAndErrors(t *testing.T) {
	if _, err := New("not-a-url"); err == nil {
		t.Fatal("expected URL error")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "model unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	transcriber, _ := New(server.URL)
	if _, err := transcriber.Transcribe(context.Background(), models.File{}); err == nil {
		t.Fatal("expected content error")
	}
	if _, err := transcriber.Transcribe(context.Background(), models.File{Content: io.NopCloser(strings.NewReader("x"))}); err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("unexpected error: %v", err)
	}
}
