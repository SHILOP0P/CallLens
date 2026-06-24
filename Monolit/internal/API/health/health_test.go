package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected health response: code=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestHealthHandlerProbes(t *testing.T) {
	dir := t.TempDir()
	handler := NewHandler(
		WritableDirectoryCheck("uploads", dir),
		BinaryCheck("go", "go"),
	)

	liveRecorder := httptest.NewRecorder()
	handler.Live(liveRecorder, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if liveRecorder.Code != http.StatusOK || liveRecorder.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected live response: code=%d body=%q", liveRecorder.Code, liveRecorder.Body.String())
	}

	startupRecorder := httptest.NewRecorder()
	handler.Startup(startupRecorder, httptest.NewRequest(http.MethodGet, "/health/startup", nil))
	if startupRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected startup status: %d", startupRecorder.Code)
	}

	readyRecorder := httptest.NewRecorder()
	handler.Ready(readyRecorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if readyRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected ready status: code=%d body=%q", readyRecorder.Code, readyRecorder.Body.String())
	}
}

func TestReadyReportsFailedCheck(t *testing.T) {
	handler := NewHandler(Check{
		Name: "broken",
		Run: func(ctx context.Context) error {
			return errors.New("not ready")
		},
	})

	rec := httptest.NewRecorder()
	handler.Ready(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected ready status: code=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestWritableDirectoryCheckRejectsFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "file-*")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	check := WritableDirectoryCheck("uploads", file.Name())
	if err := check.Run(context.Background()); err == nil {
		t.Fatal("expected file path to fail writable directory check")
	}
}

func TestHealthHandlesWriteError(t *testing.T) {
	writer := &failingWriter{header: make(http.Header)}
	Health(writer, httptest.NewRequest(http.MethodGet, "/health", nil))
	if writer.writes < 2 {
		t.Fatalf("writes = %d, want response and fallback error", writer.writes)
	}
}

type failingWriter struct {
	header http.Header
	writes int
}

func (w *failingWriter) Header() http.Header        { return w.header }
func (w *failingWriter) WriteHeader(statusCode int) {}
func (w *failingWriter) Write([]byte) (int, error) {
	w.writes++
	return 0, errors.New("write failed")
}
