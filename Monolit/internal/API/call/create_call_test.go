package call

import (
	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func (s *APISuite) TestParseCallPlacementPersonal() {
	companyID, departmentID, scope, err := parseCallPlacement("", "")

	s.Require().NoError(err)
	s.Require().False(companyID.Valid)
	s.Require().False(departmentID.Valid)
	s.Require().Equal(models.CallVisibilityScopePersonal, scope)
}

func (s *APISuite) TestParseCallPlacementCompany() {
	companyUUID := uuid.New()

	companyID, departmentID, scope, err := parseCallPlacement(companyUUID.String(), "")

	s.Require().NoError(err)
	s.Require().True(companyID.Valid)
	s.Require().Equal(companyUUID, companyID.UUID)
	s.Require().False(departmentID.Valid)
	s.Require().Equal(models.CallVisibilityScopeCompany, scope)
}

func (s *APISuite) TestParseCallPlacementDepartment() {
	companyUUID := uuid.New()
	departmentUUID := uuid.New()

	companyID, departmentID, scope, err := parseCallPlacement(companyUUID.String(), departmentUUID.String())

	s.Require().NoError(err)
	s.Require().True(companyID.Valid)
	s.Require().True(departmentID.Valid)
	s.Require().Equal(companyUUID, companyID.UUID)
	s.Require().Equal(departmentUUID, departmentID.UUID)
	s.Require().Equal(models.CallVisibilityScopeDepartment, scope)
}

func (s *APISuite) TestParseCallPlacementRejectsDepartmentWithoutCompany() {
	_, _, _, err := parseCallPlacement("", uuid.New().String())

	s.Require().ErrorIs(err, models.ErrInvalidCallPlacement)
}

func (s *APISuite) TestParseCallPlacementRejectsInvalidUUID() {
	_, _, _, err := parseCallPlacement("bad uuid", "")

	s.Require().ErrorIs(err, models.ErrInvalidCallPlacement)
}
