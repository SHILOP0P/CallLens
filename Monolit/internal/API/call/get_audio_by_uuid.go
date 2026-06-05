package call

import (
	"calllens/monolit/internal/API/response"
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
	userID, ok := userIDFromRequest(r)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	rawUUID := chi.URLParam(r, "uuid")

	callUUID, err := uuid.Parse(rawUUID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeInvalidCallUUID, "invalid call UUID")
		return
	}

	audioFile, err := h.service.GetAudioByUUID(r.Context(), callUUID, userID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			response.WriteError(w, http.StatusNotFound, response.CodeAudioNotFound, "audio not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToGetAudio, "error getting audio file")
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
