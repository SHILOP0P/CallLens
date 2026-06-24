package company

import (
	"net/http"

	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestGetCompanyMembersOverviewSuccess() {
	companyID := uuid.New()
	userID := uuid.New()

	s.service.On("GetCompanyMembersOverview", mock.Anything, companyID, userID).
		Return(models.CompanyMembersOverview{CompanyUUID: companyID}, nil).
		Once()

	rec, req := s.request(http.MethodGet, "/api/v1/companies/"+companyID.String()+"/members", "", userID, map[string]string{
		"uuid": companyID.String(),
	})

	s.api.GetCompanyMembersOverview(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *APISuite) TestGetCompanyMembersOverviewRequiresAuth() {
	companyID := uuid.New()

	rec, req := s.request(http.MethodGet, "/api/v1/companies/"+companyID.String()+"/members", "", uuid.Nil, map[string]string{
		"uuid": companyID.String(),
	})

	s.api.GetCompanyMembersOverview(rec, req)

	s.Require().Equal(http.StatusUnauthorized, rec.Code)
	s.requireErrorCode(rec, response.CodeUnauthorized)
}

func (s *APISuite) TestGetCompanyMembersOverviewRejectsInvalidCompanyUUID() {
	rec, req := s.request(http.MethodGet, "/api/v1/companies/bad/members", "", uuid.New(), map[string]string{
		"uuid": "bad",
	})

	s.api.GetCompanyMembersOverview(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidCompanyInput)
}

func (s *APISuite) TestGetCompanyMembersOverviewMapsCompanyNotFound() {
	companyID := uuid.New()
	userID := uuid.New()

	s.service.On("GetCompanyMembersOverview", mock.Anything, companyID, userID).
		Return(models.CompanyMembersOverview{}, models.ErrCompanyNotFound).
		Once()

	rec, req := s.request(http.MethodGet, "/api/v1/companies/"+companyID.String()+"/members", "", userID, map[string]string{
		"uuid": companyID.String(),
	})

	s.api.GetCompanyMembersOverview(rec, req)

	s.Require().Equal(http.StatusNotFound, rec.Code)
	s.requireErrorCode(rec, response.CodeCompanyNotFound)
}

func (s *APISuite) TestGetCompanyMembersOverviewMapsForbidden() {
	companyID := uuid.New()
	userID := uuid.New()

	s.service.On("GetCompanyMembersOverview", mock.Anything, companyID, userID).
		Return(models.CompanyMembersOverview{}, models.ErrForbidden).
		Once()

	rec, req := s.request(http.MethodGet, "/api/v1/companies/"+companyID.String()+"/members", "", userID, map[string]string{
		"uuid": companyID.String(),
	})

	s.api.GetCompanyMembersOverview(rec, req)

	s.Require().Equal(http.StatusForbidden, rec.Code)
	s.requireErrorCode(rec, response.CodeForbidden)
}
