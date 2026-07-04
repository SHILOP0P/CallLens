//go:build integration

package call

import (
	"time"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func (s *RepositorySuite) TestGetAnalyticsOverviewAggregatesVisibleFilteredCalls() {
	company, manager := s.createCompanyWithManager()
	uploader := s.createUser(uuid.NewString() + "@example.com")
	outsider := s.createUser(uuid.NewString() + "@example.com")
	baseTime := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	analyzed := testCall(uploader.ID)
	analyzed.Status = models.CallStatusAnalyzed
	analyzed.VisibilityScope = models.CallVisibilityScopeCompany
	analyzed.CompanyUUID = uuid.NullUUID{UUID: company.ID, Valid: true}
	analyzed.DurationSeconds = 60
	analyzed.CreatedAt = baseTime
	_, err := s.repository.CreateCall(s.ctx, analyzed)
	s.Require().NoError(err)
	s.insertDoneAnalysis(analyzed.ID, `{
		"quality_score": 4.5,
		"topics": ["Интеграция", "Договор"],
		"risks": ["Нет бюджета"],
		"manager_quality": {
			"issues": ["Перебивал"],
			"recommendations": ["Уточнять сроки"]
		},
		"next_steps": ["Отправить КП"]
	}`)

	failed := testCall(uploader.ID)
	failed.Status = models.CallStatusFailed
	failed.VisibilityScope = models.CallVisibilityScopeCompany
	failed.CompanyUUID = uuid.NullUUID{UUID: company.ID, Valid: true}
	failed.DurationSeconds = 120
	failed.CreatedAt = baseTime.Add(time.Hour)
	_, err = s.repository.CreateCall(s.ctx, failed)
	s.Require().NoError(err)

	outOfRange := testCall(uploader.ID)
	outOfRange.Status = models.CallStatusNew
	outOfRange.VisibilityScope = models.CallVisibilityScopeCompany
	outOfRange.CompanyUUID = uuid.NullUUID{UUID: company.ID, Valid: true}
	outOfRange.CreatedAt = baseTime.Add(48 * time.Hour)
	_, err = s.repository.CreateCall(s.ctx, outOfRange)
	s.Require().NoError(err)

	from := baseTime.Add(-time.Minute)
	to := baseTime.Add(2 * time.Hour)
	overview, err := s.repository.GetAnalyticsOverview(s.ctx, models.AnalyticsOverviewInput{
		UserID:          manager.ID,
		VisibilityScope: models.CallVisibilityScopeCompany,
		CompanyUUID:     uuid.NullUUID{UUID: company.ID, Valid: true},
		From:            &from,
		To:              &to,
	})
	s.Require().NoError(err)
	s.Require().Equal(2, overview.CallsTotal)
	s.Require().Equal(1, overview.CallsAnalyzed)
	s.Require().Equal(1, overview.CallsFailed)
	s.Require().NotNil(overview.AverageDurationSeconds)
	s.Require().Equal(90, *overview.AverageDurationSeconds)
	s.Require().NotNil(overview.AverageQualityScore)
	s.Require().Equal(4.5, *overview.AverageQualityScore)
	s.Require().NotEmpty(overview.TopTopics)
	s.Require().NotNil(overview.RisksCount)
	s.Require().Equal(2, *overview.RisksCount)
	s.Require().NotNil(overview.RecommendationsCount)
	s.Require().Equal(2, *overview.RecommendationsCount)
	s.Require().Len(overview.Charts.CallsByDay, 1)
	s.Require().Len(overview.Charts.AnalyzedByDay, 1)
	s.Require().Len(overview.Charts.QualityByDay, 1)
	s.Require().Len(overview.Charts.DurationByDay, 1)
	s.Require().Len(overview.Charts.RisksByDay, 1)

	outsiderOverview, err := s.repository.GetAnalyticsOverview(s.ctx, models.AnalyticsOverviewInput{
		UserID: outsider.ID,
		From:   &from,
		To:     &to,
	})
	s.Require().NoError(err)
	s.Require().Zero(outsiderOverview.CallsTotal)
}

func (s *RepositorySuite) insertDoneAnalysis(callID uuid.UUID, resultJSON string) {
	s.T().Helper()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO call_analyses (
			analysis_uuid,
			call_uuid,
			status,
			provider,
			result_json,
			result_text,
			created_at,
			updated_at
		) VALUES ($1, $2, 'done', 'test', $3::jsonb, 'summary', NOW(), NOW())
	`, uuid.New(), callID, resultJSON)
	s.Require().NoError(err)
}
