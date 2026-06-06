package call

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestListSuccess() {
	userID := uuid.New()

	s.service.On("List", mock.Anything, userID).
		Return([]models.Call{{ID: uuid.New(), Title: "call", Status: models.CallStatusNew, VisibilityScope: models.CallVisibilityScopePersonal, CreatedAt: time.Now().UTC()}}, nil).
		Once()

	rec, req := s.request(http.MethodGet, "/api/v1/calls", "", userID, nil)

	s.api.List(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *APISuite) TestListRequiresAuth() {
	rec, req := s.request(http.MethodGet, "/api/v1/calls", "", uuid.Nil, nil)

	s.api.List(rec, req)

	s.Require().Equal(http.StatusUnauthorized, rec.Code)
	s.requireErrorCode(rec, response.CodeUnauthorized)
}

func (s *APISuite) TestListMapsServiceError() {
	userID := uuid.New()

	s.service.On("List", mock.Anything, userID).Return(nil, errors.New("list failed")).Once()

	rec, req := s.request(http.MethodGet, "/api/v1/calls", "", userID, nil)

	s.api.List(rec, req)

	s.Require().Equal(http.StatusInternalServerError, rec.Code)
	s.requireErrorCode(rec, response.CodeFailedToListCalls)
}
