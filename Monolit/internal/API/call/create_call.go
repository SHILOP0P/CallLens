package call

import (
	"bytes"
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/converter"
	model "calllens/monolit/internal/models"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
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

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}

	detectedMimeType := http.DetectContentType(buffer[:n])
	fileContent := io.MultiReader(bytes.NewReader(buffer[:n]), file)

	req := dto.CreateCallRequest{
		Title: title,
		Audio: fileHeader,
	}

	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		http.Error(w, "audio file extension is required", http.StatusBadRequest)
		return
	}

	originalFilename := req.Audio.Filename
	//mimeType := req.Audio.Header.Get("Content-Type")
	sizeBytes := req.Audio.Size

	input := model.CreateCallInput{
		Title:            title,
		OriginalFilename: originalFilename,
		MimeType:         detectedMimeType,
		SizeBytes:        sizeBytes,
		Content:          fileContent,
	}

	createdCall, err := h.service.CreateCall(r.Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrCallConvert) {
			http.Error(w, "failed to process call", http.StatusInternalServerError)
			return
		} else if errors.Is(err, model.ErrCallNotFound) {
			http.Error(w, "call not found", http.StatusInternalServerError)
			return
		} else if errors.Is(err, model.ErrUnsupportedAudioType) {
			http.Error(w, "unsupported audio type", http.StatusBadRequest)
			return
		} else {
			http.Error(w, "failed to create call", http.StatusInternalServerError)
			return
		}
	}

	response, err := converter.CallModelToAPI(createdCall)
	if err != nil {
		http.Error(w, "failed to create call", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
