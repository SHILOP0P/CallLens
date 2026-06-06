package company

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestAddCompanyMemberSuccess() {
	companyID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()

	s.service.On("AddCompanyMember", mock.Anything, models.AddCompanyMemberInput{
		CompanyUUID: companyID,
		RequestUser: requestUserID,
		UserUUID:    userID,
		Role:        models.CompanyMemberRoleEmployee,
	}).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: userID, Role: models.CompanyMemberRoleEmployee, Status: models.MembershipStatusActive, CreatedAt: time.Now().UTC()}, nil).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/members", `{"user_uuid":"`+userID.String()+`","role":"employee"}`, requestUserID, map[string]string{
		"uuid": companyID.String(),
	})

	s.api.AddCompanyMember(rec, req)

	s.Require().Equal(http.StatusCreated, rec.Code)
}

func (s *APISuite) TestAddCompanyMemberRejectsInvalidCompanyUUID() {
	rec, req := s.request(http.MethodPost, "/api/v1/companies/bad/members", `{}`, uuid.New(), map[string]string{
		"uuid": "bad",
	})

	s.api.AddCompanyMember(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidCompanyInput)
}

func (s *APISuite) TestAddCompanyMemberRejectsInvalidUserUUID() {
	companyID := uuid.New()

	rec, req := s.request(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/members", `{"user_uuid":"bad","role":"employee"}`, uuid.New(), map[string]string{
		"uuid": companyID.String(),
	})

	s.api.AddCompanyMember(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidCompanyInput)
}

func (s *APISuite) TestAddCompanyMemberMapsForbidden() {
	companyID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()

	s.service.On("AddCompanyMember", mock.Anything, models.AddCompanyMemberInput{
		CompanyUUID: companyID,
		RequestUser: requestUserID,
		UserUUID:    userID,
		Role:        models.CompanyMemberRoleEmployee,
	}).
		Return(models.CompanyMember{}, models.ErrForbidden).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/members", `{"user_uuid":"`+userID.String()+`","role":"employee"}`, requestUserID, map[string]string{
		"uuid": companyID.String(),
	})

	s.api.AddCompanyMember(rec, req)

	s.Require().Equal(http.StatusForbidden, rec.Code)
	s.requireErrorCode(rec, response.CodeForbidden)
}
