package company

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"

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
