package department

import (
	"calllens/monolit/internal/logger"
	repositoryMocks "calllens/monolit/internal/repository/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServiceSuite struct {
	suite.Suite
	ctx                  context.Context
	companyRepository    *repositoryMocks.CompanyRepository
	departmentRepository *repositoryMocks.DepartmentRepository
	service              *Service
}

func (s *ServiceSuite) SetupTest() {
	s.ctx = context.Background()
	s.companyRepository = repositoryMocks.NewCompanyRepository(s.T())
	s.departmentRepository = repositoryMocks.NewDepartmentRepository(s.T())
	s.service = NewService(s.companyRepository, s.departmentRepository, logger.NewNop())
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
