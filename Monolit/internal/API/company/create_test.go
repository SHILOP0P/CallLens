package company

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestCreateSuccess() {
	userID := uuid.New()
	companyID := uuid.New()

	s.service.On("CreateCompany", mock.Anything, models.CreateCompanyInput{
		Name:          "CallLens",
		ManagerUserID: userID,
	}).
		Return(models.Company{ID: companyID, Name: "CallLens", ManagerUserUUID: userID, MemberLimit: 1, CreatedAt: time.Now().UTC()}, nil).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/companies", `{"name":"CallLens"}`, userID, nil)

	s.api.Create(rec, req)

	s.Require().Equal(http.StatusCreated, rec.Code)
}

func (s *APISuite) TestCreateRequiresAuth() {
	rec, req := s.request(http.MethodPost, "/api/v1/companies", `{"name":"CallLens"}`, uuid.Nil, nil)

	s.api.Create(rec, req)

	s.Require().Equal(http.StatusUnauthorized, rec.Code)
	s.requireErrorCode(rec, response.CodeUnauthorized)
}

func (s *APISuite) TestCreateRejectsInvalidBody() {
	rec, req := s.request(http.MethodPost, "/api/v1/companies", `{`, uuid.New(), nil)

	s.api.Create(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidRequestBody)
}

func (s *APISuite) TestCreateMapsAlreadyManagedCompany() {
	userID := uuid.New()

	s.service.On("CreateCompany", mock.Anything, models.CreateCompanyInput{
		Name:          "CallLens",
		ManagerUserID: userID,
	}).
		Return(models.Company{}, models.ErrUserAlreadyManagesCompany).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/companies", `{"name":"CallLens"}`, userID, nil)

	s.api.Create(rec, req)

	s.Require().Equal(http.StatusConflict, rec.Code)
	s.requireErrorCode(rec, response.CodeUserAlreadyManagesCompany)
}
