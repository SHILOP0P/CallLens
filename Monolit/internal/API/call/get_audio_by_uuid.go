package call

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *CallHandler) GetAudioByUUID(w http.ResponseWriter, r *http.Request) {
	rawUUID := chi.URLParam(r, "uuid")

	callUUID, err := uuid.Parse(rawUUID)
	if err != nil {
		http.Error(w, "invalid call UUID", http.StatusBadRequest)
		return
	}

	audioFile, err := h.service.GetAudioByUUID(r.Context(), callUUID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "audio not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error getting audio file", http.StatusInternalServerError)
		return
	}

	defer audioFile.Content.Close()

	contentType := audioFile.MimeType

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{
		"filename": audioFile.OriginalFilename,
	}))

	if audioFile.SizeBytes > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(audioFile.SizeBytes, 10))
	}

	if _, err := io.Copy(w, audioFile.Content); err != nil {
		return
	}
}
