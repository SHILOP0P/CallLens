//go:build integration

package admin

import (
	"time"

	"calllens/monolit/internal/models"
)

func (s *RepositorySuite) TestGrantExtendAndCancelPersonalSubscription() {
	actor := s.createUser(models.UserRoleAdmin)
	target := s.createUser(models.UserRoleUser)
	now := time.Now().UTC().Truncate(time.Second)
	reason := "manual payment"
	granted, err := s.repository.GrantAdminSubscription(s.ctx, models.GrantAdminSubscriptionInput{
		ActorUserUUID: actor.ID, UserUUID: target.ID, PlanCode: models.PlanCodePersonalPlus,
		StartsAt: now, EndsAt: now.Add(30 * 24 * time.Hour), Metadata: models.AdminMutationMetadata{Reason: reason},
	})
	s.Require().NoError(err)
	s.Require().Equal(models.SubscriptionStatusActive, granted.Status)
	s.Require().Equal(models.PlanCodePersonalPlus, granted.PlanCode)

	extendedEnd := now.Add(60 * 24 * time.Hour)
	extended, err := s.repository.GrantAdminSubscription(s.ctx, models.GrantAdminSubscriptionInput{
		ActorUserUUID: actor.ID, UserUUID: target.ID, PlanCode: models.PlanCodePersonalPlus,
		StartsAt: now, EndsAt: extendedEnd, Metadata: models.AdminMutationMetadata{Reason: reason},
	})
	s.Require().NoError(err)
	s.Require().Equal(granted.ID, extended.ID)
	s.Require().NotNil(extended.EndsAt)
	s.Require().True(extended.EndsAt.Equal(extendedEnd))

	canceled, err := s.repository.CancelAdminSubscription(s.ctx, models.CancelAdminSubscriptionInput{ActorUserUUID: actor.ID, UserUUID: target.ID, Metadata: models.AdminMutationMetadata{Reason: reason}})
	s.Require().NoError(err)
	s.Require().Equal(models.SubscriptionStatusCanceled, canceled.Status)
	var count int
	s.Require().NoError(s.db.QueryRowContext(s.ctx, `SELECT COUNT(*) FROM admin_audit_logs WHERE action IN ('subscription.granted','subscription.extended','subscription.canceled')`).Scan(&count))
	s.Require().Equal(3, count)
}
