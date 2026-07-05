package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func TestCreateDeepAnalysisReturnsExistingWithoutSpendingLimit(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	existing := models.AggregateAnalysis{ID: uuid.New(), Scope: models.AggregateAnalysisScopePersonal, Status: models.AggregateAnalysisStatusDone}
	repo := &analyticsRepoStub{reusable: &existing}
	svc := NewService(repo)
	svc.SetAnalyzer(&aggregateAnalyzerStub{})

	got, err := svc.CreateDeepAnalysis(ctx, models.CreateDeepAnalysisInput{
		UserID: userID, Scope: models.AggregateAnalysisScopePersonal, PeriodFrom: day(2026, 7, 1), PeriodTo: dayEnd(2026, 7, 7),
	})
	if err != nil {
		t.Fatalf("create deep analysis: %v", err)
	}
	if got.ID != existing.ID {
		t.Fatalf("returned analysis = %s, want %s", got.ID, existing.ID)
	}
	if repo.spent {
		t.Fatal("limit was spent for reusable analysis")
	}
}

func TestCreateDeepAnalysisForceSpendsLimitAndMarksDone(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := &analyticsRepoStub{sourceTotal: 1, sources: []models.AggregateAnalysisSourceCall{{CallUUID: uuid.New(), Title: "Call"}}}
	svc := NewService(repo)
	svc.SetAnalyzer(&aggregateAnalyzerStub{result: result("Готово")})

	got, err := svc.CreateDeepAnalysis(ctx, models.CreateDeepAnalysisInput{
		UserID: userID, Scope: models.AggregateAnalysisScopePersonal, PeriodFrom: day(2026, 7, 1), PeriodTo: dayEnd(2026, 7, 7), Force: true,
	})
	if err != nil {
		t.Fatalf("create deep analysis: %v", err)
	}
	if !repo.spent {
		t.Fatal("limit was not spent")
	}
	if !repo.markedProcessing || !repo.markedDone {
		t.Fatal("analysis lifecycle did not reach processing and done")
	}
	if got.Status != models.AggregateAnalysisStatusDone {
		t.Fatalf("status = %s", got.Status)
	}
}

func TestCreateDeepAnalysisMapsLimitNoCallsAndProviderError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	input := models.CreateDeepAnalysisInput{UserID: userID, Scope: models.AggregateAnalysisScopePersonal, PeriodFrom: day(2026, 7, 1), PeriodTo: dayEnd(2026, 7, 7), Force: true}

	noCallsRepo := &analyticsRepoStub{}
	noCallsSvc := NewService(noCallsRepo)
	noCallsSvc.SetAnalyzer(&aggregateAnalyzerStub{})
	_, err := noCallsSvc.CreateDeepAnalysis(ctx, input)
	if !errors.Is(err, models.ErrNoAnalyzedCallsForDeepAnalysis) {
		t.Fatalf("no calls err = %v", err)
	}

	limitRepo := &analyticsRepoStub{sourceTotal: 1, sources: []models.AggregateAnalysisSourceCall{{CallUUID: uuid.New()}}, spendErr: models.ErrDeepAnalysisLimitExceeded}
	limitSvc := NewService(limitRepo)
	limitSvc.SetAnalyzer(&aggregateAnalyzerStub{})
	_, err = limitSvc.CreateDeepAnalysis(ctx, input)
	if !errors.Is(err, models.ErrDeepAnalysisLimitExceeded) {
		t.Fatalf("limit err = %v", err)
	}

	providerRepo := &analyticsRepoStub{sourceTotal: 1, sources: []models.AggregateAnalysisSourceCall{{CallUUID: uuid.New()}}}
	providerSvc := NewService(providerRepo)
	providerSvc.SetAnalyzer(&aggregateAnalyzerStub{err: errors.New("provider failed")})
	_, err = providerSvc.CreateDeepAnalysis(ctx, input)
	if err == nil || !providerRepo.markedFailed {
		t.Fatalf("provider err = %v, markedFailed = %v", err, providerRepo.markedFailed)
	}
}

type analyticsRepoStub struct {
	reusable         *models.AggregateAnalysis
	sourceTotal      int
	sources          []models.AggregateAnalysisSourceCall
	spendErr         error
	spent            bool
	markedProcessing bool
	markedDone       bool
	markedFailed     bool
	created          models.AggregateAnalysis
}

func (r *analyticsRepoStub) GetAnalyticsOverview(context.Context, models.AnalyticsOverviewInput) (models.AnalyticsOverview, error) {
	panic("not implemented")
}

func (r *analyticsRepoStub) CreateAggregateAnalysis(_ context.Context, analysis models.AggregateAnalysis) (models.AggregateAnalysis, error) {
	r.created = analysis
	return analysis, nil
}

func (r *analyticsRepoStub) GetAggregateAnalysisByUUID(context.Context, uuid.UUID) (models.AggregateAnalysis, error) {
	panic("not implemented")
}

func (r *analyticsRepoStub) FindReusableAggregateAnalysis(context.Context, models.CreateDeepAnalysisInput) (models.AggregateAnalysis, error) {
	if r.reusable != nil {
		return *r.reusable, nil
	}
	return models.AggregateAnalysis{}, models.ErrAggregateAnalysisNotFound
}

func (r *analyticsRepoStub) ListAggregateAnalyses(context.Context, models.ListDeepAnalysesInput) (models.ListAggregateAnalysesResult, error) {
	panic("not implemented")
}

func (r *analyticsRepoStub) MarkAggregateAnalysisProcessing(_ context.Context, id uuid.UUID) (models.AggregateAnalysis, error) {
	r.markedProcessing = true
	r.created.ID = id
	r.created.Status = models.AggregateAnalysisStatusProcessing
	return r.created, nil
}

func (r *analyticsRepoStub) MarkAggregateAnalysisDone(_ context.Context, id uuid.UUID, result models.AnalysisResult, sourceCallsCount int) (models.AggregateAnalysis, error) {
	r.markedDone = true
	r.created.ID = id
	r.created.Status = models.AggregateAnalysisStatusDone
	r.created.ResultJSON = result.ResultJSON
	r.created.SourceCallsCount = sourceCallsCount
	return r.created, nil
}

func (r *analyticsRepoStub) MarkAggregateAnalysisFailed(_ context.Context, id uuid.UUID, message string) (models.AggregateAnalysis, error) {
	r.markedFailed = true
	r.created.ID = id
	r.created.Status = models.AggregateAnalysisStatusFailed
	r.created.ErrorMessage = &message
	return r.created, nil
}

func (r *analyticsRepoStub) ListAggregateAnalysisSourceCalls(context.Context, models.AnalyticsOverviewInput, int) ([]models.AggregateAnalysisSourceCall, int, error) {
	return r.sources, r.sourceTotal, nil
}

func (r *analyticsRepoStub) SpendDeepAnalysisUsage(context.Context, models.DeepAnalysisSubjectType, uuid.UUID, time.Time, time.Time) error {
	if r.spendErr != nil {
		return r.spendErr
	}
	r.spent = true
	return nil
}

type aggregateAnalyzerStub struct {
	result models.AnalysisResult
	err    error
}

func (a *aggregateAnalyzerStub) Provider() string { return "test" }

func (a *aggregateAnalyzerStub) Analyze(context.Context, models.AnalysisRequest) (models.AnalysisResult, error) {
	panic("not implemented")
}

func (a *aggregateAnalyzerStub) AnalyzeAggregate(context.Context, models.AggregateAnalysisRequest) (models.AnalysisResult, error) {
	if a.err != nil {
		return models.AnalysisResult{}, a.err
	}
	if len(a.result.ResultJSON) == 0 {
		return result("Готово"), nil
	}
	return a.result, nil
}

func result(summary string) models.AnalysisResult {
	raw, _ := json.Marshal(map[string]any{"summary": summary})
	text := summary
	return models.AnalysisResult{ResultJSON: raw, ResultText: &text}
}

func day(year int, month time.Month, date int) time.Time {
	return time.Date(year, month, date, 0, 0, 0, 0, time.UTC)
}

func dayEnd(year int, month time.Month, date int) time.Time {
	return day(year, month, date).AddDate(0, 0, 1).Add(-time.Nanosecond)
}
