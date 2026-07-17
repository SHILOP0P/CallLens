package openrouter

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"calllens/monolit/internal/models"
)

func TestNewRequiresAPIKey(t *testing.T) {
	_, err := New("", "openai/whisper-large-v3-turbo")
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
		if req.Model != "openai/whisper-large-v3-turbo" {
			t.Fatalf("model = %q", req.Model)
		}
		if req.Language != defaultLanguage {
			t.Fatalf("language = %q", req.Language)
		}
		if req.Temperature == nil || *req.Temperature != 0 {
			t.Fatalf("temperature = %v", req.Temperature)
		}
		if req.ResponseFormat != "verbose_json" {
			t.Fatalf("response format = %q", req.ResponseFormat)
		}
		if len(req.TimestampGranularities) != 1 || req.TimestampGranularities[0] != "segment" {
			t.Fatalf("timestamp granularities = %#v", req.TimestampGranularities)
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

	transcriber, err := New("sk-or-v1-test", "openai/whisper-large-v3-turbo")
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

	transcriber, err := New("sk-or-v1-test", "openai/whisper-large-v3-turbo")
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
	transcriber, err := New("sk-or-v1-test", "openai/whisper-large-v3-turbo")
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

func TestIsVideo(t *testing.T) {
	for _, file := range []models.File{
		{MimeType: "video/mp4"},
		{MimeType: "video/webm; codecs=vp9"},
		{OriginalFilename: "meeting.MOV"},
	} {
		if !isVideo(file) {
			t.Fatalf("expected video: %+v", file)
		}
	}
	if isVideo(models.File{MimeType: "audio/mp4", OriginalFilename: "call.m4a"}) {
		t.Fatal("m4a audio detected as video")
	}
}

func TestTranscriptionAudioExtractsVideoTrack(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg is not installed")
	}

	videoPath := filepath.Join(t.TempDir(), "meeting.mp4")
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=c=black:s=32x32:d=0.2", "-f", "lavfi", "-i", "sine=frequency=440:duration=0.2", "-shortest", "-c:v", "mpeg4", "-c:a", "aac", videoPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create fixture video: %v: %s", err, output)
	}

	video, err := os.Open(videoPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = video.Close() }()

	audio, format, err := transcriptionAudio(context.Background(), models.File{
		Content:          video,
		OriginalFilename: "meeting.mp4",
		MimeType:         "video/mp4",
	})
	if err != nil {
		t.Fatalf("extract audio: %v", err)
	}
	if format != "wav" || len(audio) < 44 || string(audio[:4]) != "RIFF" {
		t.Fatalf("unexpected extracted audio: format=%q size=%d", format, len(audio))
	}
	if got, want := binary.LittleEndian.Uint32(audio[4:8]), uint32(len(audio)-8); got != want {
		t.Fatalf("RIFF size = %d, want %d", got, want)
	}
	dataSize, ok := wavDataSize(audio)
	if !ok || dataSize == 0 {
		t.Fatalf("WAV data chunk was not found: size=%d", len(audio))
	}
}

func wavDataSize(audio []byte) (uint32, bool) {
	for offset := 12; offset+8 <= len(audio); {
		chunkSize := binary.LittleEndian.Uint32(audio[offset+4 : offset+8])
		if string(audio[offset:offset+4]) == "data" {
			return chunkSize, int(chunkSize) <= len(audio)-offset-8
		}
		next := offset + 8 + int(chunkSize)
		if chunkSize%2 != 0 {
			next++
		}
		if next <= offset || next > len(audio) {
			return 0, false
		}
		offset = next
	}
	return 0, false
}

func TestProviderEndpointAndAudioFormats(t *testing.T) {
	transcriber, err := New(" key ", " model ")
	if err != nil {
		t.Fatal(err)
	}
	if transcriber.Provider() != providerName || transcriber.endpoint() != defaultBaseURL+transcribePath {
		t.Fatalf("provider=%q endpoint=%q", transcriber.Provider(), transcriber.endpoint())
	}
	transcriber.baseURL = "https://example.com/"
	if transcriber.endpoint() != "https://example.com"+transcribePath {
		t.Fatalf("endpoint = %q", transcriber.endpoint())
	}

	cases := []struct {
		file models.File
		want string
	}{
		{file: models.File{MimeType: "audio/mpeg; codecs=test"}, want: "mp3"},
		{file: models.File{MimeType: "audio/x-wav"}, want: "wav"},
		{file: models.File{MimeType: "audio/mp4"}, want: "m4a"},
		{file: models.File{MimeType: "audio/flac"}, want: "flac"},
		{file: models.File{MimeType: "audio/ogg"}, want: "ogg"},
		{file: models.File{MimeType: "audio/webm"}, want: "webm"},
		{file: models.File{MimeType: "audio/aac"}, want: "aac"},
		{file: models.File{OriginalFilename: "call.MP3"}, want: "mp3"},
		{file: models.File{Path: "stored/call.m4a"}, want: "m4a"},
	}
	for _, tt := range cases {
		got, err := audioFormat(tt.file)
		if err != nil || got != tt.want {
			t.Fatalf("audioFormat(%+v) = %q, %v; want %q", tt.file, got, err, tt.want)
		}
	}
	if got := supportedFormatFromExt("call.txt"); got != "" {
		t.Fatalf("unsupported extension = %q", got)
	}
}

func TestNormalizeSegmentsAndText(t *testing.T) {
	start, end := 1.0, 2.0
	segments := normalizeSegments([]transcriptionSegment{
		{Speaker: " A ", Start: &start, End: &end, Text: " Hello!!! "},
		{Speaker: "B", Text: "[music]"},
	})
	if len(segments) != 1 || segments[0].Speaker != "A" || segments[0].StartSeconds != &start {
		t.Fatalf("segments = %+v", segments)
	}
	text := textFromSegments([]models.TranscriptionSegment{
		{Speaker: " A ", Text: "hello"},
		{Text: "world"},
		{Speaker: "B", Text: " "},
	})
	if text != "A: hello\nworld" {
		t.Fatalf("text = %q", text)
	}
}

func TestTranscribeValidationAndResponseErrors(t *testing.T) {
	transcriber, _ := New("key", "model")
	if _, err := transcriber.Transcribe(context.Background(), models.File{}); !errors.Is(err, models.ErrUnsupportedAudioType) {
		t.Fatalf("nil content error = %v", err)
	}
	if _, err := transcriber.Transcribe(context.Background(), models.File{
		Content: io.NopCloser(strings.NewReader("")), OriginalFilename: "call.mp3",
	}); !errors.Is(err, models.ErrUnsupportedAudioType) {
		t.Fatalf("empty content error = %v", err)
	}
	if _, err := transcriber.Transcribe(context.Background(), models.File{
		Content: errReadCloser{}, OriginalFilename: "call.mp3",
	}); err == nil || !strings.Contains(err.Error(), "read audio content") {
		t.Fatalf("read error = %v", err)
	}

	for _, responseBody := range []string{"{", `{"text":"[music]"}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(responseBody))
		}))
		transcriber.baseURL = server.URL
		transcriber.client = server.Client()
		_, err := transcriber.Transcribe(context.Background(), models.File{
			Content: io.NopCloser(strings.NewReader("audio")), OriginalFilename: "call.mp3",
		})
		server.Close()
		if err == nil {
			t.Fatalf("response %q unexpectedly succeeded", responseBody)
		}
	}
}

func TestTranscribeBuildsTextFromSegments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"segments":[{"speaker":"A","text":"hello"},{"text":"world"}]}`))
	}))
	defer server.Close()

	transcriber, _ := New("key", "model")
	transcriber.baseURL = server.URL
	transcriber.client = server.Client()
	result, err := transcriber.Transcribe(context.Background(), models.File{
		Content: io.NopCloser(strings.NewReader("audio")), OriginalFilename: "call.mp3",
	})
	if err != nil || result.Text != "A: hello\nworld" {
		t.Fatalf("result = %+v, err=%v", result, err)
	}
}

func TestDecodeErrorFallbacks(t *testing.T) {
	err := decodeError(&http.Response{
		StatusCode: http.StatusBadGateway,
		Body:       io.NopCloser(strings.NewReader("plain error")),
	})
	if !strings.Contains(err.Error(), "plain error") {
		t.Fatalf("plain error = %v", err)
	}

	err = decodeError(&http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Body:       io.NopCloser(strings.NewReader("")),
	})
	if !strings.Contains(err.Error(), http.StatusText(http.StatusServiceUnavailable)) {
		t.Fatalf("empty error = %v", err)
	}

	err = decodeError(&http.Response{StatusCode: http.StatusBadGateway, Body: errReadCloser{}})
	if !strings.Contains(err.Error(), "read error response") {
		t.Fatalf("read error = %v", err)
	}
}

type errReadCloser struct{}

func (errReadCloser) Read([]byte) (int, error) { return 0, fmt.Errorf("read failed") }
func (errReadCloser) Close() error             { return nil }
