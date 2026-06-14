package billing

import (
	"calllens/monolit/internal/models"
	"time"

	"github.com/google/uuid"
)

func (s *RepositorySuite) TestListPlansIncludesDefaultPlans() {
	plans, err := s.repository.ListPlans(s.ctx)

	s.Require().NoError(err)
	s.Require().Len(plans, 6)
	s.Require().Equal(models.PlanCodePersonalStart, plans[0].Code)
	s.Require().Equal(120, plans[0].MonthlyMinutesLimit)
	s.Require().Equal(2, plans[0].ActiveInstructionLimit)
	s.Require().Equal(models.PlanCodeBusinessPro, plans[5].Code)
	s.Require().NotNil(plans[5].CompanyLimit)
	s.Require().Equal(3, *plans[5].CompanyLimit)
	s.Require().NotNil(plans[5].InstructionsPerDepartmentLimit)
	s.Require().Equal(10, *plans[5].InstructionsPerDepartmentLimit)
	s.Require().True(plans[5].APIAccessEnabled)
}

func (s *RepositorySuite) TestUpsertSubscriptionAndGetBusinessSubscriptionByCompanyOwner() {
	ownerID := s.createUser("owner@example.com")
	companyID := s.createCompany(ownerID)

	created, err := s.repository.UpsertSubscription(s.ctx, models.UpsertSubscriptionInput{
		PlanCode: models.PlanCodeBusinessPro,
		UserUUID: uuid.NullUUID{UUID: ownerID, Valid: true},
		Status:   models.SubscriptionStatusActive,
		StartsAt: time.Now().UTC().Add(-time.Hour),
	})
	s.Require().NoError(err)
	s.Require().Equal(models.PlanCodeBusinessPro, created.Plan.Code)

	byOwner, err := s.repository.GetActiveBusinessSubscriptionForOwner(s.ctx, ownerID)
	s.Require().NoError(err)
	s.Require().Equal(created.ID, byOwner.ID)

	byCompany, err := s.repository.GetActiveBusinessSubscription(s.ctx, companyID)
	s.Require().NoError(err)
	s.Require().Equal(created.ID, byCompany.ID)
}

func (s *RepositorySuite) TestAddUsageMinutesAccumulatesCurrentPeriod() {
	userID := s.createUser("usage@example.com")
	subscription, err := s.repository.UpsertSubscription(s.ctx, models.UpsertSubscriptionInput{
		PlanCode: models.PlanCodePersonalStart,
		UserUUID: uuid.NullUUID{UUID: userID, Valid: true},
		Status:   models.SubscriptionStatusActive,
		StartsAt: time.Now().UTC().Add(-time.Hour),
	})
	s.Require().NoError(err)

	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	_, err = s.repository.AddUsageMinutes(s.ctx, subscription.ID, now, 3)
	s.Require().NoError(err)
	_, err = s.repository.AddUsageMinutes(s.ctx, subscription.ID, now, 4)
	s.Require().NoError(err)

	usedMinutes, err := s.repository.CountUsedMinutes(s.ctx, subscription.ID, now)
	s.Require().NoError(err)
	s.Require().Equal(7, usedMinutes)
}

func (s *RepositorySuite) createUser(email string) uuid.UUID {
	id := uuid.New()
	_, err := s.db.ExecContext(
		s.ctx,
		`INSERT INTO users (user_uuid, email, password_hash, full_name, full_surname, nick_name, role, created_at)
		 VALUES ($1, $2, 'hash', 'Dmitry', 'Mukhachev', 'muxa', 'user', $3)`,
		id,
		email,
		time.Now().UTC(),
	)
	s.Require().NoError(err)

	return id
}

func (s *RepositorySuite) createCompany(ownerID uuid.UUID) uuid.UUID {
	companyID := uuid.New()
	_, err := s.db.ExecContext(
		s.ctx,
		`INSERT INTO companies (company_uuid, name, manager_user_uuid, member_limit, created_at)
		 VALUES ($1, 'CallLens', $2, 25, $3)`,
		companyID,
		ownerID,
		time.Now().UTC(),
	)
	s.Require().NoError(err)

	return companyID
}
