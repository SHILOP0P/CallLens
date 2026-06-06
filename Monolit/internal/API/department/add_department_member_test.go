package department

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestAddDepartmentMemberSuccess() {
	companyID := uuid.New()
	departmentID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()

	s.service.On("AddDepartmentMember", mock.Anything, models.AddDepartmentMemberInput{
		CompanyUUID:    companyID,
		DepartmentUUID: departmentID,
		RequestUser:    requestUserID,
		UserUUID:       userID,
		Role:           models.DepartmentMemberRoleEmployee,
	}).
		Return(models.DepartmentMember{DepartmentUUID: departmentID, UserUUID: userID, Role: models.DepartmentMemberRoleEmployee, Status: models.MembershipStatusActive, CreatedAt: time.Now().UTC()}, nil).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/departments/"+departmentID.String()+"/members", `{"user_uuid":"`+userID.String()+`","role":"employee"}`, requestUserID, map[string]string{
		"uuid":            companyID.String(),
		"department_uuid": departmentID.String(),
	})

	s.api.AddDepartmentMember(rec, req)

	s.Require().Equal(http.StatusCreated, rec.Code)
}

func (s *APISuite) TestAddDepartmentMemberRejectsInvalidUserUUID() {
	companyID := uuid.New()
	departmentID := uuid.New()

	rec, req := s.request(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/departments/"+departmentID.String()+"/members", `{"user_uuid":"bad","role":"employee"}`, uuid.New(), map[string]string{
		"uuid":            companyID.String(),
		"department_uuid": departmentID.String(),
	})

	s.api.AddDepartmentMember(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidDepartmentInput)
}
