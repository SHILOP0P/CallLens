package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	if err := WriteJSON(rec, http.StatusCreated, map[string]string{"status": "ok"}); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	if rec.Code != http.StatusCreated || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected response: code=%d content-type=%q", rec.Code, rec.Header().Get("Content-Type"))
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["status"] != "ok" {
		t.Fatalf("unexpected body: %s, err=%v", rec.Body.String(), err)
	}
}

func TestWriteJSONNilAndEncodingError(t *testing.T) {
	rec := httptest.NewRecorder()
	if err := WriteJSON(rec, http.StatusAccepted, nil); err != nil || rec.Code != http.StatusAccepted {
		t.Fatalf("nil payload response: code=%d err=%v", rec.Code, err)
	}

	err := WriteJSON(httptest.NewRecorder(), http.StatusOK, func() {})
	if err == nil {
		t.Fatal("expected JSON encoding error")
	}
}

func TestWriteJSONWriteError(t *testing.T) {
	writer := &failingWriter{header: make(http.Header)}
	if err := WriteJSON(writer, http.StatusOK, map[string]bool{"ok": true}); !errors.Is(err, errWrite) {
		t.Fatalf("WriteJSON error = %v", err)
	}
}

func TestWriteErrorAndNoContent(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadRequest, CodeInvalidUserInput, "bad input")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	WriteNoContent(rec)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

var errWrite = errors.New("write failed")

type failingWriter struct {
	header http.Header
}

func (w *failingWriter) Header() http.Header        { return w.header }
func (w *failingWriter) WriteHeader(statusCode int) {}
func (w *failingWriter) Write([]byte) (int, error)  { return 0, errWrite }
