//go:build integration

package admin

import (
	"time"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func (s *RepositorySuite) TestChangeAdminUserRoleInvalidatesAccessAndAudits() {
	actor := s.createUser(models.UserRoleAdmin)
	target := s.createUser(models.UserRoleUser)
	sessionID := uuid.New()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO refresh_sessions (session_uuid,user_uuid,refresh_token_hash,access_version,created_at,expires_at)
		VALUES ($1,$2,$3,1,$4,$5)
	`, sessionID, target.ID, uuid.NewString(), time.Now().UTC(), time.Now().UTC().Add(time.Hour))
	s.Require().NoError(err)
	reason := "support promotion"
	before := models.UserRoleUser
	updated, err := s.repository.ChangeAdminUserRole(s.ctx, models.ChangeAdminUserRoleInput{
		ActorUserUUID: actor.ID, TargetUserUUID: target.ID, ExpectedRole: before, Role: models.UserRoleHelper,
		Metadata: models.AdminMutationMetadata{Reason: reason},
	})
	s.Require().NoError(err)
	s.Require().Equal(models.UserRoleHelper, updated.Role)
	var version int64
	s.Require().NoError(s.db.QueryRowContext(s.ctx, `SELECT access_version FROM refresh_sessions WHERE session_uuid=$1`, sessionID).Scan(&version))
	s.Require().Equal(int64(2), version)
	var action string
	var beforeJSON, afterJSON []byte
	s.Require().NoError(s.db.QueryRowContext(s.ctx, `SELECT action,before_data,after_data FROM admin_audit_logs`).Scan(&action, &beforeJSON, &afterJSON))
	s.Require().Equal("user.role_changed", action)
	s.Require().JSONEq(`{"role":"user"}`, string(beforeJSON))
	s.Require().JSONEq(`{"role":"helper"}`, string(afterJSON))
}

func (s *RepositorySuite) TestAdminCannotRevokeAdminSession() {
	actor := s.createUser(models.UserRoleAdmin)
	target := s.createUser(models.UserRoleAdmin)
	reason := "incident"
	err := s.repository.RevokeAllAdminUserSessions(s.ctx, models.AdminSessionMutationInput{
		ActorUserUUID: actor.ID, TargetUserUUID: target.ID, Metadata: models.AdminMutationMetadata{Reason: reason},
	})
	s.Require().ErrorIs(err, models.ErrAdminSessionManagementForbidden)
}
