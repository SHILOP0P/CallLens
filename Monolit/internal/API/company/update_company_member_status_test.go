package company

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestUpdateCompanyMemberStatusSuccess() {
	companyID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()

	s.service.On("UpdateCompanyMemberStatus", mock.Anything, models.UpdateCompanyMemberStatusInput{
		CompanyUUID: companyID,
		RequestUser: requestUserID,
		UserUUID:    userID,
		Status:      models.MembershipStatusSuspended,
	}).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: userID, Status: models.MembershipStatusSuspended, CreatedAt: time.Now().UTC()}, nil).
		Once()

	rec, req := s.request(http.MethodPatch, "/api/v1/companies/"+companyID.String()+"/members/"+userID.String()+"/status", `{"status":"suspended"}`, requestUserID, map[string]string{
		"uuid":      companyID.String(),
		"user_uuid": userID.String(),
	})

	s.api.UpdateCompanyMemberStatus(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *APISuite) TestUpdateCompanyMemberStatusRejectsInvalidBody() {
	companyID := uuid.New()
	userID := uuid.New()

	rec, req := s.request(http.MethodPatch, "/api/v1/companies/"+companyID.String()+"/members/"+userID.String()+"/status", `{`, uuid.New(), map[string]string{
		"uuid":      companyID.String(),
		"user_uuid": userID.String(),
	})

	s.api.UpdateCompanyMemberStatus(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidRequestBody)
}
