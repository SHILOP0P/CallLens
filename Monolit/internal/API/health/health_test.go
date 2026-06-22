package health

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected health response: code=%d body=%q", rec.Code, rec.Body.String())
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
