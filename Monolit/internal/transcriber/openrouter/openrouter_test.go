package openrouter

import (
	"calllens/monolit/internal/models"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRequiresAPIKey(t *testing.T) {
	_, err := New("", "qwen/qwen3-asr-flash-2026-02-10")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewRequiresModel(t *testing.T) {
	_, err := New("sk-or-v1-test", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTranscribeSendsOpenRouterRequest(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != transcribePath {
			t.Fatalf("path = %s, want %s", r.URL.Path, transcribePath)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-or-v1-test" {
			t.Fatalf("authorization = %q", got)
		}

		var req transcriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "qwen/qwen3-asr-flash-2026-02-10" {
			t.Fatalf("model = %q", req.Model)
		}
		if req.Language != defaultLanguage {
			t.Fatalf("language = %q", req.Language)
		}
		if req.Temperature == nil || *req.Temperature != 0 {
			t.Fatalf("temperature = %v", req.Temperature)
		}
		if req.InputAudio.Format != "mp3" {
			t.Fatalf("format = %q", req.InputAudio.Format)
		}
		audio, err := base64.StdEncoding.DecodeString(req.InputAudio.Data)
		if err != nil {
			t.Fatalf("decode audio: %v", err)
		}
		if string(audio) != "audio bytes" {
			t.Fatalf("audio = %q", string(audio))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":" [музыка]\nПривет, это тест!!!\nСубтитры сделал Иван ","segments":[{"speaker":"speaker_0","start":0,"end":1.25,"text":"Привет, это тест!!!"}]}`))
	}))
	defer server.Close()

	transcriber, err := New("sk-or-v1-test", "qwen/qwen3-asr-flash-2026-02-10")
	if err != nil {
		t.Fatalf("new transcriber: %v", err)
	}
	transcriber.baseURL = server.URL
	transcriber.client = server.Client()

	got, err := transcriber.Transcribe(ctx, models.File{
		Content:          io.NopCloser(strings.NewReader("audio bytes")),
		OriginalFilename: "call.mp3",
		MimeType:         "audio/mpeg",
	})
	if err != nil {
		t.Fatalf("transcribe: %v", err)
	}
	if got.Text != "Привет, это тест!" {
		t.Fatalf("text = %q", got.Text)
	}
	if len(got.Segments) != 1 {
		t.Fatalf("segments len = %d, want 1", len(got.Segments))
	}
	if got.Segments[0].Speaker != "speaker_0" || got.Segments[0].Text != "Привет, это тест!" {
		t.Fatalf("segment = %#v", got.Segments[0])
	}
	if got.Segments[0].StartSeconds == nil || *got.Segments[0].StartSeconds != 0 {
		t.Fatalf("segment start = %v", got.Segments[0].StartSeconds)
	}
	if got.Segments[0].EndSeconds == nil || *got.Segments[0].EndSeconds != 1.25 {
		t.Fatalf("segment end = %v", got.Segments[0].EndSeconds)
	}
	if got.Language == nil || *got.Language != defaultLanguage {
		t.Fatalf("language = %v", got.Language)
	}
}

func TestTranscribeReturnsOpenRouterError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid api key","code":"invalid_credentials"}}`))
	}))
	defer server.Close()

	transcriber, err := New("sk-or-v1-test", "qwen/qwen3-asr-flash-2026-02-10")
	if err != nil {
		t.Fatalf("new transcriber: %v", err)
	}
	transcriber.baseURL = server.URL
	transcriber.client = server.Client()

	_, err = transcriber.Transcribe(context.Background(), models.File{
		Content:          io.NopCloser(strings.NewReader("audio bytes")),
		OriginalFilename: "call.wav",
		MimeType:         "audio/wav",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "status 401") || !strings.Contains(err.Error(), "invalid api key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTranscribeRejectsUnsupportedFormat(t *testing.T) {
	transcriber, err := New("sk-or-v1-test", "qwen/qwen3-asr-flash-2026-02-10")
	if err != nil {
		t.Fatalf("new transcriber: %v", err)
	}

	_, err = transcriber.Transcribe(context.Background(), models.File{
		Content:          io.NopCloser(strings.NewReader("audio bytes")),
		OriginalFilename: "call.txt",
		MimeType:         "text/plain",
	})
	if !errors.Is(err, models.ErrUnsupportedAudioType) {
		t.Fatalf("error = %v, want unsupported audio type", err)
	}
}
