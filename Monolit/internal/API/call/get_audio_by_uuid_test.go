package call

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestGetAudioByUUIDSuccess() {
	callID := uuid.New()
	userID := uuid.New()

	s.service.On("GetAudioByUUID", mock.Anything, callID, userID).
		Return(models.File{
			Content:          io.NopCloser(strings.NewReader("audio")),
			OriginalFilename: "call.wav",
			MimeType:         "audio/wav",
			SizeBytes:        5,
		}, nil).
		Once()

	rec, req := s.request(http.MethodGet, "/api/v1/calls/"+callID.String()+"/audio", "", userID, map[string]string{"uuid": callID.String()})

	s.api.GetAudioByUUID(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
	s.Require().Equal("audio/wav", rec.Header().Get("Content-Type"))
	s.Require().Equal("5", rec.Header().Get("Content-Length"))
}

func (s *APISuite) TestGetAudioByUUIDDefaultsContentType() {
	callID := uuid.New()
	userID := uuid.New()

	s.service.On("GetAudioByUUID", mock.Anything, callID, userID).
		Return(models.File{Content: io.NopCloser(strings.NewReader("audio")), OriginalFilename: "call.bin"}, nil).
		Once()

	rec, req := s.request(http.MethodGet, "/api/v1/calls/"+callID.String()+"/audio", "", userID, map[string]string{"uuid": callID.String()})

	s.api.GetAudioByUUID(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
	s.Require().Equal("application/octet-stream", rec.Header().Get("Content-Type"))
}

func (s *APISuite) TestGetAudioByUUIDMapsNotFound() {
	callID := uuid.New()
	userID := uuid.New()

	s.service.On("GetAudioByUUID", mock.Anything, callID, userID).Return(models.File{}, os.ErrNotExist).Once()

	rec, req := s.request(http.MethodGet, "/api/v1/calls/"+callID.String()+"/audio", "", userID, map[string]string{"uuid": callID.String()})

	s.api.GetAudioByUUID(rec, req)

	s.Require().Equal(http.StatusNotFound, rec.Code)
	s.requireErrorCode(rec, response.CodeAudioNotFound)
}

func (s *APISuite) TestGetAudioByUUIDMapsUnexpectedError() {
	callID := uuid.New()
	userID := uuid.New()

	s.service.On("GetAudioByUUID", mock.Anything, callID, userID).Return(models.File{}, errors.New("open failed")).Once()

	rec, req := s.request(http.MethodGet, "/api/v1/calls/"+callID.String()+"/audio", "", userID, map[string]string{"uuid": callID.String()})

	s.api.GetAudioByUUID(rec, req)

	s.Require().Equal(http.StatusInternalServerError, rec.Code)
	s.requireErrorCode(rec, response.CodeFailedToGetAudio)
}
