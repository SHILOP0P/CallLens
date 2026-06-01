package call

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/converter"
	model "calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func (h *CallHandler) Create(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 100 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "audio file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	req := dto.CreateCallRequest{
		Title: title,
		Audio: fileHeader,
	}

	callUUID := uuid.New()

	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		http.Error(w, "audio file extension is required", http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		http.Error(w, "failed to create upload dir", http.StatusInternalServerError)
		return
	}

	audioPath := filepath.Join(h.uploadDir, fmt.Sprintf("%s%s", callUUID, ext))

	dst, err := os.Create(audioPath)
	if err != nil {
		http.Error(w, "failed to create audio file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "failed to save audio file", http.StatusInternalServerError)
		return
	}

	originalFilename := req.Audio.Filename
	mimeType := req.Audio.Header.Get("Content-Type")
	sizeBytes := req.Audio.Size
	now := time.Now().UTC()

	call, err := converter.CreateAPIToModel(callUUID, title, model.CallStatusNew, audioPath,
		originalFilename, mimeType, sizeBytes, now)
	if err != nil {
		http.Error(w, "failed to create call", http.StatusInternalServerError)
		return
	}

	createdCall, err := h.service.CreateCall(r.Context(), call)
	if err != nil {
		if errors.Is(err, model.ErrCallConvert) {
			http.Error(w, "failed to process call", http.StatusInternalServerError)
			return
		} else if errors.Is(err, model.ErrCallNotFound) {
			http.Error(w, "call not found", http.StatusInternalServerError)
			return
		} else {
			http.Error(w, "failed to create call", http.StatusInternalServerError)
			return
		}
	}

	response, err := converter.CallModelToAPI(createdCall)
	if err != nil {
		http.Error(w, "failed to create call", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
