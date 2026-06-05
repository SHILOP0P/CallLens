package call

import (
	"bytes"
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/converter"
	model "calllens/monolit/internal/models"
	"errors"
	"io"
	"net/http"
	"path/filepath"
)

func (h *CallHandler) Create(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 100 << 20

	userID, ok := userIDFromRequest(r)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeInvalidMultipartForm, "failed to parse multipart form")
		return
	}

	title := r.FormValue("title")
	if title == "" {
		response.WriteError(w, http.StatusBadRequest, response.CodeCallTitleRequired, "title is required")
		return
	}
	file, fileHeader, err := r.FormFile("audio")
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeAudioFileRequired, "audio file is required")
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeAudioFileReadFailed, "failed to read file")
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
		response.WriteError(w, http.StatusBadRequest, response.CodeAudioFileExtensionRequired, "audio file extension is required")
		return
	}

	originalFilename := req.Audio.Filename
	//mimeType := req.Audio.Header.Get("Content-Type")
	sizeBytes := req.Audio.Size

	input := model.CreateCallInput{
		Title:              title,
		OriginalFilename:   originalFilename,
		MimeType:           detectedMimeType,
		SizeBytes:          sizeBytes,
		Content:            fileContent,
		UploadedByUserUUID: userID,
	}

	createdCall, err := h.service.CreateCall(r.Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrCallConvert) {
			response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToProcessCall, "failed to process call")
			return
		} else if errors.Is(err, model.ErrCallNotFound) {
			response.WriteError(w, http.StatusInternalServerError, response.CodeCallNotFound, "call not found")
			return
		} else if errors.Is(err, model.ErrUnsupportedAudioType) {
			response.WriteError(w, http.StatusBadRequest, response.CodeUnsupportedAudioType, "unsupported audio type")
			return
		} else if errors.Is(err, model.ErrInvalidCallOwner) {
			response.WriteError(w, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
			return
		} else {
			response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToCreateCall, "failed to create call")
			return
		}
	}

	resp, err := converter.CallModelToAPI(createdCall)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToCreateCall, "failed to create call")
		return
	}

	if err := response.WriteJSON(w, http.StatusCreated, resp); err != nil {
		return
	}
}
