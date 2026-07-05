package analytics

import (
	"context"
	"errors"
	"time"

	"calllens/monolit/internal/analyzer"
	"calllens/monolit/internal/models"
	"calllens/monolit/internal/repository"

	"github.com/google/uuid"
)

type Service struct {
	analyticsRepository  repository.AnalyticsRepository
	callFolderRepository repository.CallFolderRepository
	companyRepository    repository.CompanyRepository
	departmentRepository repository.DepartmentRepository
	analyzer             analyzer.Analyzer
	now                  func() time.Time
}

func NewService(analyticsRepository repository.AnalyticsRepository) *Service {
	return &Service{analyticsRepository: analyticsRepository, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) SetCallFolderRepository(repository repository.CallFolderRepository) {
	s.callFolderRepository = repository
}

func (s *Service) SetCompanyRepository(repository repository.CompanyRepository) {
	s.companyRepository = repository
}

func (s *Service) SetDepartmentRepository(repository repository.DepartmentRepository) {
	s.departmentRepository = repository
}

func (s *Service) SetAnalyzer(analyzer analyzer.Analyzer) {
	s.analyzer = analyzer
}

func (s *Service) GetOverview(ctx context.Context, input models.AnalyticsOverviewInput) (models.AnalyticsOverview, error) {
	if input.FolderUUID.Valid {
		if s.callFolderRepository == nil {
			return models.AnalyticsOverview{}, models.ErrCallFolderNotFound
		}
		if _, err := s.callFolderRepository.GetVisibleByUUID(ctx, input.FolderUUID.UUID, input.UserID); err != nil {
			return models.AnalyticsOverview{}, err
		}
	}
	return s.analyticsRepository.GetAnalyticsOverview(ctx, input)
}

func (s *Service) CreateDeepAnalysis(ctx context.Context, input models.CreateDeepAnalysisInput) (models.AggregateAnalysis, error) {
	if err := s.normalizeCreateInput(ctx, &input); err != nil {
		return models.AggregateAnalysis{}, err
	}
	if err := s.authorizeCreate(ctx, input); err != nil {
		return models.AggregateAnalysis{}, err
	}
	if !input.Force {
		existing, err := s.analyticsRepository.FindReusableAggregateAnalysis(ctx, input)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, models.ErrAggregateAnalysisNotFound) {
			return models.AggregateAnalysis{}, err
		}
	}
	overviewInput := analyticsInputFromDeepInput(input)
	sources, total, err := s.analyticsRepository.ListAggregateAnalysisSourceCalls(ctx, overviewInput, 100)
	if err != nil {
		return models.AggregateAnalysis{}, err
	}
	if total == 0 {
		return models.AggregateAnalysis{}, models.ErrNoAnalyzedCallsForDeepAnalysis
	}
	subjectType, subjectID, err := s.limitSubject(ctx, input)
	if err != nil {
		return models.AggregateAnalysis{}, err
	}
	periodStart, periodEnd := utcWeek(input.PeriodFrom)
	if err := s.analyticsRepository.SpendDeepAnalysisUsage(ctx, subjectType, subjectID, periodStart, periodEnd); err != nil {
		return models.AggregateAnalysis{}, err
	}

	now := s.now()
	analysis := models.AggregateAnalysis{
		ID:                uuid.New(),
		Scope:             input.Scope,
		CompanyUUID:       input.CompanyUUID,
		DepartmentUUID:    input.DepartmentUUID,
		FolderUUID:        input.FolderUUID,
		PeriodFrom:        input.PeriodFrom,
		PeriodTo:          input.PeriodTo,
		Status:            models.AggregateAnalysisStatusPending,
		Provider:          s.analyzer.Provider(),
		SourceCallsCount:  total,
		CreatedByUserUUID: input.UserID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if input.Scope == models.AggregateAnalysisScopePersonal {
		analysis.UserUUID = uuid.NullUUID{UUID: input.UserID, Valid: true}
	}
	created, err := s.analyticsRepository.CreateAggregateAnalysis(ctx, analysis)
	if err != nil {
		return models.AggregateAnalysis{}, err
	}
	processing, err := s.analyticsRepository.MarkAggregateAnalysisProcessing(ctx, created.ID)
	if err != nil {
		return models.AggregateAnalysis{}, err
	}
	result, err := s.analyzer.AnalyzeAggregate(ctx, models.AggregateAnalysisRequest{
		Scope: input.Scope, PeriodFrom: input.PeriodFrom, PeriodTo: input.PeriodTo, SourceCallsCount: total,
		Sources: sources, Metrics: models.AggregateAnalysisSourceMetrics{IncludedCalls: len(sources), TotalCalls: total},
	})
	if err != nil {
		failed, markErr := s.analyticsRepository.MarkAggregateAnalysisFailed(ctx, processing.ID, err.Error())
		if markErr != nil {
			return models.AggregateAnalysis{}, markErr
		}
		return failed, err
	}
	return s.analyticsRepository.MarkAggregateAnalysisDone(ctx, processing.ID, result, total)
}

func (s *Service) ListDeepAnalyses(ctx context.Context, input models.ListDeepAnalysesInput) (models.ListAggregateAnalysesResult, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 || input.Offset < 0 {
		return models.ListAggregateAnalysesResult{}, models.ErrInvalidDeepAnalysisInput
	}
	if input.Scope != "" && !validAggregateScope(input.Scope) {
		return models.ListAggregateAnalysesResult{}, models.ErrInvalidDeepAnalysisInput
	}
	if input.Status != "" && !validAggregateStatus(input.Status) {
		return models.ListAggregateAnalysesResult{}, models.ErrInvalidDeepAnalysisInput
	}
	return s.analyticsRepository.ListAggregateAnalyses(ctx, input)
}

func (s *Service) GetDeepAnalysis(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.AggregateAnalysis, error) {
	analysis, err := s.analyticsRepository.GetAggregateAnalysisByUUID(ctx, id)
	if err != nil {
		return models.AggregateAnalysis{}, err
	}
	if err := s.authorizeRead(ctx, analysis, userID); err != nil {
		return models.AggregateAnalysis{}, models.ErrAggregateAnalysisNotFound
	}
	return analysis, nil
}

func (s *Service) normalizeCreateInput(ctx context.Context, input *models.CreateDeepAnalysisInput) error {
	if s.analyzer == nil || !validAggregateScope(input.Scope) || input.UserID == uuid.Nil || input.PeriodFrom.IsZero() || input.PeriodTo.IsZero() || input.PeriodFrom.After(input.PeriodTo) {
		return models.ErrInvalidDeepAnalysisInput
	}
	input.PeriodFrom = input.PeriodFrom.UTC()
	input.PeriodTo = input.PeriodTo.UTC()
	if input.FolderUUID.Valid {
		if s.callFolderRepository == nil {
			return models.ErrInvalidDeepAnalysisInput
		}
		folder, err := s.callFolderRepository.GetByUUID(ctx, input.FolderUUID.UUID)
		if err != nil {
			return err
		}
		input.Scope = models.AggregateAnalysisScopeFolder
		input.CompanyUUID = folder.CompanyUUID
		input.DepartmentUUID = folder.DepartmentUUID
		return nil
	}
	switch input.Scope {
	case models.AggregateAnalysisScopePersonal:
		if input.CompanyUUID.Valid || input.DepartmentUUID.Valid {
			return models.ErrInvalidDeepAnalysisInput
		}
	case models.AggregateAnalysisScopeCompany:
		if !input.CompanyUUID.Valid || input.DepartmentUUID.Valid {
			return models.ErrInvalidDeepAnalysisInput
		}
	case models.AggregateAnalysisScopeDepartment:
		if !input.CompanyUUID.Valid || !input.DepartmentUUID.Valid {
			return models.ErrInvalidDeepAnalysisInput
		}
	case models.AggregateAnalysisScopeFolder:
		return models.ErrInvalidDeepAnalysisInput
	}
	return nil
}

func (s *Service) authorizeCreate(ctx context.Context, input models.CreateDeepAnalysisInput) error {
	switch input.Scope {
	case models.AggregateAnalysisScopePersonal:
		return nil
	case models.AggregateAnalysisScopeCompany:
		return s.requireCompanyManager(ctx, input.CompanyUUID.UUID, input.UserID)
	case models.AggregateAnalysisScopeDepartment:
		return s.requireDepartmentLeaderOrCompanyManager(ctx, input.CompanyUUID.UUID, input.DepartmentUUID.UUID, input.UserID)
	case models.AggregateAnalysisScopeFolder:
		folder, err := s.callFolderRepository.GetByUUID(ctx, input.FolderUUID.UUID)
		if err != nil {
			return err
		}
		switch folder.Scope {
		case models.CallFolderScopePersonal:
			if folder.UserUUID.Valid && folder.UserUUID.UUID == input.UserID {
				return nil
			}
			return models.ErrForbidden
		case models.CallFolderScopeCompany:
			return s.requireCompanyManager(ctx, folder.CompanyUUID.UUID, input.UserID)
		case models.CallFolderScopeDepartment:
			return s.requireDepartmentLeaderOrCompanyManager(ctx, folder.CompanyUUID.UUID, folder.DepartmentUUID.UUID, input.UserID)
		}
	}
	return models.ErrInvalidDeepAnalysisInput
}

func (s *Service) authorizeRead(ctx context.Context, analysis models.AggregateAnalysis, userID uuid.UUID) error {
	if analysis.UserUUID.Valid && analysis.UserUUID.UUID == userID {
		return nil
	}
	if analysis.CompanyUUID.Valid {
		if member, err := s.companyRepository.GetCompanyMember(ctx, analysis.CompanyUUID.UUID, userID); err == nil && member.Status == models.MembershipStatusActive {
			return nil
		}
	}
	if analysis.DepartmentUUID.Valid {
		if _, err := s.departmentRepository.GetDepartmentMember(ctx, analysis.CompanyUUID.UUID, analysis.DepartmentUUID.UUID, userID); err == nil {
			return nil
		}
	}
	return models.ErrForbidden
}

func (s *Service) requireCompanyManager(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) error {
	member, err := s.companyRepository.GetCompanyMember(ctx, companyID, userID)
	if err != nil || member.Role != models.CompanyMemberRoleManager || member.Status != models.MembershipStatusActive {
		return models.ErrForbidden
	}
	return nil
}

func (s *Service) requireDepartmentLeaderOrCompanyManager(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID, userID uuid.UUID) error {
	if err := s.requireCompanyManager(ctx, companyID, userID); err == nil {
		return nil
	}
	member, err := s.departmentRepository.GetDepartmentMember(ctx, companyID, departmentID, userID)
	if err != nil || member.Role != models.DepartmentMemberRoleLeader || member.Status != models.MembershipStatusActive {
		return models.ErrForbidden
	}
	return nil
}

func (s *Service) limitSubject(ctx context.Context, input models.CreateDeepAnalysisInput) (models.DeepAnalysisSubjectType, uuid.UUID, error) {
	if input.Scope == models.AggregateAnalysisScopePersonal {
		return models.DeepAnalysisSubjectTypeUser, input.UserID, nil
	}
	if input.Scope == models.AggregateAnalysisScopeFolder {
		folder, err := s.callFolderRepository.GetByUUID(ctx, input.FolderUUID.UUID)
		if err != nil {
			return "", uuid.Nil, err
		}
		if folder.Scope == models.CallFolderScopePersonal {
			return models.DeepAnalysisSubjectTypeUser, input.UserID, nil
		}
		return models.DeepAnalysisSubjectTypeCompany, folder.CompanyUUID.UUID, nil
	}
	return models.DeepAnalysisSubjectTypeCompany, input.CompanyUUID.UUID, nil
}

func analyticsInputFromDeepInput(input models.CreateDeepAnalysisInput) models.AnalyticsOverviewInput {
	visibilityScope := models.CallVisibilityScope(input.Scope)
	if input.Scope == models.AggregateAnalysisScopeFolder {
		visibilityScope = ""
	}
	return models.AnalyticsOverviewInput{
		UserID: input.UserID, VisibilityScope: visibilityScope, CompanyUUID: input.CompanyUUID,
		DepartmentUUID: input.DepartmentUUID, From: &input.PeriodFrom, To: &input.PeriodTo, FolderUUID: input.FolderUUID,
	}
}

func utcWeek(t time.Time) (time.Time, time.Time) {
	d := t.UTC()
	weekday := int(d.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
	return start, start.AddDate(0, 0, 7).Add(-time.Nanosecond)
}

func validAggregateScope(scope models.AggregateAnalysisScope) bool {
	switch scope {
	case models.AggregateAnalysisScopePersonal, models.AggregateAnalysisScopeCompany, models.AggregateAnalysisScopeDepartment, models.AggregateAnalysisScopeFolder:
		return true
	default:
		return false
	}
}

func validAggregateStatus(status models.AggregateAnalysisStatus) bool {
	switch status {
	case models.AggregateAnalysisStatusPending, models.AggregateAnalysisStatusProcessing, models.AggregateAnalysisStatusDone, models.AggregateAnalysisStatusFailed:
		return true
	default:
		return false
	}
}
