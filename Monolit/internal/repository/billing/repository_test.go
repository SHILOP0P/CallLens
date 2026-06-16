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

func (s *RepositorySuite) TestUserBusinessSubscriptionDoesNotCoverCompany() {
	managerID := s.createUser("manager-business-user@example.com")
	companyID := s.createCompany(managerID)

	created, err := s.repository.UpsertSubscription(s.ctx, models.UpsertSubscriptionInput{
		PlanCode: models.PlanCodeBusinessPro,
		UserUUID: uuid.NullUUID{UUID: managerID, Valid: true},
		Status:   models.SubscriptionStatusActive,
		StartsAt: time.Now().UTC().Add(-time.Hour),
	})
	s.Require().NoError(err)
	s.Require().Equal(models.PlanCodeBusinessPro, created.Plan.Code)

	_, err = s.repository.GetActiveBusinessSubscription(s.ctx, companyID)
	s.Require().ErrorIs(err, models.ErrSubscriptionNotFound)
}

func (s *RepositorySuite) TestGetBestActiveBusinessSubscriptionForManager() {
	managerID := s.createUser("best-business-manager@example.com")
	firstCompanyID := s.createCompany(managerID)
	secondCompanyID := s.createCompany(managerID)

	start, err := s.repository.ActivateCompanySubscription(s.ctx, models.ActivateCompanySubscriptionInput{
		CompanyUUID: firstCompanyID,
		PlanCode:    models.PlanCodeBusinessStart,
	}, time.Now().UTC().Add(-time.Hour))
	s.Require().NoError(err)

	pro, err := s.repository.ActivateCompanySubscription(s.ctx, models.ActivateCompanySubscriptionInput{
		CompanyUUID: secondCompanyID,
		PlanCode:    models.PlanCodeBusinessPro,
	}, time.Now().UTC().Add(-time.Hour))
	s.Require().NoError(err)

	best, err := s.repository.GetBestActiveBusinessSubscriptionForManager(s.ctx, managerID)
	s.Require().NoError(err)
	s.Require().Equal(pro.ID, best.ID)
	s.Require().NotEqual(start.ID, best.ID)
}

func (s *RepositorySuite) TestActivateAndCancelCompanySubscription() {
	ownerID := s.createUser("company-subscription-owner@example.com")
	companyID := s.createCompany(ownerID)

	startsAt := time.Now().UTC().Add(-time.Hour)
	created, err := s.repository.ActivateCompanySubscription(s.ctx, models.ActivateCompanySubscriptionInput{
		CompanyUUID: companyID,
		PlanCode:    models.PlanCodeBusinessPlus,
	}, startsAt)
	s.Require().NoError(err)
	s.Require().Equal(models.PlanCodeBusinessPlus, created.Plan.Code)
	s.Require().Equal(models.SubscriptionStatusActive, created.Status)
	s.Require().True(created.CompanyUUID.Valid)
	s.Require().Equal(companyID, created.CompanyUUID.UUID)

	updated, err := s.repository.ActivateCompanySubscription(s.ctx, models.ActivateCompanySubscriptionInput{
		CompanyUUID: companyID,
		PlanCode:    models.PlanCodeBusinessPro,
	}, startsAt.Add(time.Minute))
	s.Require().NoError(err)
	s.Require().Equal(created.ID, updated.ID)
	s.Require().Equal(models.PlanCodeBusinessPro, updated.Plan.Code)

	active, err := s.repository.GetActiveBusinessSubscription(s.ctx, companyID)
	s.Require().NoError(err)
	s.Require().Equal(updated.ID, active.ID)

	canceled, err := s.repository.CancelCompanySubscription(s.ctx, companyID, time.Now().UTC())
	s.Require().NoError(err)
	s.Require().Equal(updated.ID, canceled.ID)
	s.Require().Equal(models.SubscriptionStatusCanceled, canceled.Status)
	s.Require().NotNil(canceled.EndsAt)

	_, err = s.repository.GetActiveBusinessSubscription(s.ctx, companyID)
	s.Require().ErrorIs(err, models.ErrSubscriptionNotFound)
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
		`INSERT INTO users (user_uuid, email, password_hash, full_name, full_surname, username, role, created_at)
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
