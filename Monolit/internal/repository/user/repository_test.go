package user

import (
	"calllens/monolit/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

func testUser() models.User {
	post := "manager"

	return models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: "hash",
		FullName:     "Dmitry",
		FullSurname:  "Mukhachev",
		Username:     "muxa",
		Role:         models.UserRoleUser,
		Post:         &post,
		CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
	}
}

func (s *RepositorySuite) TestCreateUserAndGetByUUID() {
	input := testUser()

	created, err := s.repository.CreateUser(s.ctx, input)
	s.Require().NoError(err)
	s.Require().Equal(input.ID, created.ID)
	s.Require().Equal(input.Email, created.Email)
	s.Require().Equal(input.PasswordHash, created.PasswordHash)
	s.Require().Equal(input.FullName, created.FullName)
	s.Require().Equal(input.FullSurname, created.FullSurname)
	s.Require().Equal(input.Username, created.Username)
	s.Require().Equal(input.Role, created.Role)
	s.Require().NotNil(created.Post)
	s.Require().Equal(*input.Post, *created.Post)

	got, err := s.repository.GetUserByUUID(s.ctx, input.ID)
	s.Require().NoError(err)
	s.Require().Equal(created, got)
}

func (s *RepositorySuite) TestGetUserByEmailIsCaseInsensitive() {
	input := testUser()
	_, err := s.repository.CreateUser(s.ctx, input)
	s.Require().NoError(err)

	got, err := s.repository.GetUserByEmail(s.ctx, strings.ToUpper(input.Email))

	s.Require().NoError(err)
	s.Require().Equal(input.ID, got.ID)
	s.Require().Equal(input.Email, got.Email)
}

func (s *RepositorySuite) TestGetUserByUUIDNotFound() {
	_, err := s.repository.GetUserByUUID(s.ctx, uuid.New())

	s.Require().ErrorIs(err, models.ErrUserNotFound)
}

func (s *RepositorySuite) TestGetUserByEmailNotFound() {
	_, err := s.repository.GetUserByEmail(s.ctx, "missing@example.com")

	s.Require().ErrorIs(err, models.ErrUserNotFound)
}

func (s *RepositorySuite) TestCreateUserRejectsDuplicateEmailCaseInsensitive() {
	input := testUser()
	_, err := s.repository.CreateUser(s.ctx, input)
	s.Require().NoError(err)

	duplicate := testUser()
	duplicate.ID = uuid.New()
	duplicate.Email = strings.ToUpper(input.Email)

	_, err = s.repository.CreateUser(s.ctx, duplicate)

	s.Require().Error(err)
}
