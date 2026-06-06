package department

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestUpdateDepartmentMemberRoleSuccess() {
	companyID := uuid.New()
	departmentID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()

	s.service.On("UpdateDepartmentMemberRole", mock.Anything, models.UpdateDepartmentMemberRoleInput{
		CompanyUUID:    companyID,
		DepartmentUUID: departmentID,
		RequestUser:    requestUserID,
		UserUUID:       userID,
		Role:           models.DepartmentMemberRoleLeader,
	}).
		Return(models.DepartmentMember{DepartmentUUID: departmentID, UserUUID: userID, Role: models.DepartmentMemberRoleLeader, Status: models.MembershipStatusActive, CreatedAt: time.Now().UTC()}, nil).
		Once()

	rec, req := s.request(http.MethodPatch, "/api/v1/companies/"+companyID.String()+"/departments/"+departmentID.String()+"/members/"+userID.String()+"/role", `{"role":"department_leader"}`, requestUserID, map[string]string{
		"uuid":            companyID.String(),
		"department_uuid": departmentID.String(),
		"user_uuid":       userID.String(),
	})

	s.api.UpdateDepartmentMemberRole(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *APISuite) TestUpdateDepartmentMemberRoleMapsInvalidInput() {
	companyID := uuid.New()
	departmentID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()

	s.service.On("UpdateDepartmentMemberRole", mock.Anything, mock.Anything).
		Return(models.DepartmentMember{}, models.ErrInvalidDepartmentInput).
		Once()

	rec, req := s.request(http.MethodPatch, "/api/v1/companies/"+companyID.String()+"/departments/"+departmentID.String()+"/members/"+userID.String()+"/role", `{"role":"manager"}`, requestUserID, map[string]string{
		"uuid":            companyID.String(),
		"department_uuid": departmentID.String(),
		"user_uuid":       userID.String(),
	})

	s.api.UpdateDepartmentMemberRole(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidDepartmentInput)
}
