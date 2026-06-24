//go:build integration

package invitation

import (
	"context"
	"database/sql"
	"testing"

	companyRepo "calllens/monolit/internal/repository/company"
	departmentRepo "calllens/monolit/internal/repository/department"
	"calllens/monolit/internal/repository/repositorytest"
	userRepo "calllens/monolit/internal/repository/user"

	"github.com/stretchr/testify/suite"
)

type RepositorySuite struct {
	suite.Suite
	ctx                  context.Context
	db                   *sql.DB
	repository           *Repository
	userRepository       *userRepo.Repository
	companyRepository    *companyRepo.Repository
	departmentRepository *departmentRepo.Repository
}

func (s *RepositorySuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skip integration tests in short mode")
	}

	s.ctx = context.Background()
	s.db = repositorytest.OpenTestDB(s.T())
	repositorytest.RunMigrations(s.T(), s.db)
}

func (s *RepositorySuite) SetupTest() {
	repositorytest.TruncateTables(s.T(), s.db)
	s.repository = NewRepository(s.db)
	s.userRepository = userRepo.NewUserRepository(s.db)
	s.companyRepository = companyRepo.NewRepository(s.db)
	s.departmentRepository = departmentRepo.NewRepository(s.db)
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}
