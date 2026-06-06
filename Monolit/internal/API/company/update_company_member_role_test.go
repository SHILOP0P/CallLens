package company

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestUpdateCompanyMemberRoleSuccess() {
	companyID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()

	s.service.On("UpdateCompanyMemberRole", mock.Anything, models.UpdateCompanyMemberRoleInput{
		CompanyUUID: companyID,
		RequestUser: requestUserID,
		UserUUID:    userID,
		Role:        models.CompanyMemberRoleEmployee,
	}).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: userID, Role: models.CompanyMemberRoleEmployee, Status: models.MembershipStatusActive, CreatedAt: time.Now().UTC()}, nil).
		Once()

	rec, req := s.request(http.MethodPatch, "/api/v1/companies/"+companyID.String()+"/members/"+userID.String()+"/role", `{"role":"employee"}`, requestUserID, map[string]string{
		"uuid":      companyID.String(),
		"user_uuid": userID.String(),
	})

	s.api.UpdateCompanyMemberRole(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *APISuite) TestUpdateCompanyMemberRoleMapsInvalidInput() {
	companyID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()

	s.service.On("UpdateCompanyMemberRole", mock.Anything, mock.Anything).
		Return(models.CompanyMember{}, models.ErrInvalidCompanyInput).
		Once()

	rec, req := s.request(http.MethodPatch, "/api/v1/companies/"+companyID.String()+"/members/"+userID.String()+"/role", `{"role":"company_manager"}`, requestUserID, map[string]string{
		"uuid":      companyID.String(),
		"user_uuid": userID.String(),
	})

	s.api.UpdateCompanyMemberRole(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidCompanyInput)
}
