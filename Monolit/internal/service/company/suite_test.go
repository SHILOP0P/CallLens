package company

import (
	"calllens/monolit/internal/logger"
	repositoryMocks "calllens/monolit/internal/repository/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServiceSuite struct {
	suite.Suite
	ctx        context.Context
	repository *repositoryMocks.CompanyRepository
	service    *Service
}

func (s *ServiceSuite) SetupTest() {
	s.ctx = context.Background()
	s.repository = repositoryMocks.NewCompanyRepository(s.T())
	s.service = NewService(s.repository, logger.NewNop())
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
