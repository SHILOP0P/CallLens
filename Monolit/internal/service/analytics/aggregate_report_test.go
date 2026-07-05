package analytics

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAggregateReportGeneratorMarkdownAndXLSX(t *testing.T) {
	text := "fallback text"
	analysis := models.AggregateAnalysis{
		ID: uuid.New(), Scope: models.AggregateAnalysisScopePersonal,
		PeriodFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		PeriodTo:   time.Date(2026, 7, 7, 23, 59, 59, 0, time.UTC),
		Status:     models.AggregateAnalysisStatusDone, SourceCallsCount: 3,
		ResultJSON: []byte(`{"summary":"summary text","priority_actions":["call back"]}`),
		ResultText: &text,
	}
	md, err := generateAggregateReport(models.ReportFormatMD, AggregateReportData{Analysis: analysis, GeneratedAt: analysis.PeriodTo})
	require.NoError(t, err)
	require.Contains(t, string(md), "summary text")
	require.Contains(t, string(md), "call back")

	xlsx, err := generateAggregateReport(models.ReportFormatXLSX, AggregateReportData{Analysis: analysis, GeneratedAt: analysis.PeriodTo})
	require.NoError(t, err)
	require.NotEmpty(t, xlsx)
}

func TestAggregateReportGeneratorFallsBackOnMalformedJSON(t *testing.T) {
	text := "plain result"
	analysis := models.AggregateAnalysis{ID: uuid.New(), ResultJSON: []byte(`{bad`), ResultText: &text}
	content, err := generateAggregateReport(models.ReportFormatMD, AggregateReportData{Analysis: analysis, GeneratedAt: time.Now()})
	require.NoError(t, err)
	require.Contains(t, string(content), "plain result")
}

func TestCreateAggregateReportRequiresDoneAnalysis(t *testing.T) {
	userID := uuid.New()
	analysis := aggregateReportAnalysis(userID)
	analysis.Status = models.AggregateAnalysisStatusProcessing
	svc := NewService(&aggregateReportAnalyticsRepo{analysis: analysis})
	svc.SetReportRepository(&aggregateReportRepo{})
	svc.SetReportStorage(&aggregateReportStorage{})

	_, err := svc.CreateAggregateReport(context.Background(), models.CreateAggregateReportInput{
		AggregateAnalysisUUID: analysis.ID, UserUUID: userID, Format: models.ReportFormatMD,
	})
	require.ErrorIs(t, err, models.ErrInvalidAnalysisStatus)
}

func TestCreateDownloadDeleteAggregateReport(t *testing.T) {
	userID := uuid.New()
	analysis := aggregateReportAnalysis(userID)
	reports := &aggregateReportRepo{}
	storage := &aggregateReportStorage{files: map[string][]byte{}}
	svc := NewService(&aggregateReportAnalyticsRepo{analysis: analysis})
	svc.SetReportRepository(reports)
	svc.SetReportStorage(storage)

	report, err := svc.CreateAggregateReport(context.Background(), models.CreateAggregateReportInput{
		AggregateAnalysisUUID: analysis.ID, UserUUID: userID, Format: models.ReportFormatMD,
	})
	require.NoError(t, err)
	require.Equal(t, models.ReportStatusReady, report.Status)

	list, err := svc.ListAggregateReports(context.Background(), analysis.ID, userID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	file, err := svc.GetAggregateReportFile(context.Background(), report.ID, userID)
	require.NoError(t, err)
	body, _ := io.ReadAll(file.Content)
	require.Contains(t, string(body), "summary")

	require.NoError(t, svc.DeleteAggregateReport(context.Background(), report.ID, userID))
	_, err = svc.GetAggregateReportFile(context.Background(), report.ID, userID)
	require.ErrorIs(t, err, models.ErrAggregateReportNotFound)
}

func TestAggregateReportMissingStorageFile(t *testing.T) {
	userID := uuid.New()
	analysis := aggregateReportAnalysis(userID)
	reportID := uuid.New()
	path := "missing.md"
	reports := &aggregateReportRepo{reports: map[uuid.UUID]models.AggregateReportExport{reportID: {
		ID: reportID, AggregateAnalysisUUID: analysis.ID, RequestedByUserUUID: userID,
		Format: models.ReportFormatMD, Status: models.ReportStatusReady, StoragePath: &path, ExpiresAt: time.Now().Add(time.Hour),
	}}}
	svc := NewService(&aggregateReportAnalyticsRepo{analysis: analysis})
	svc.SetReportRepository(reports)
	svc.SetReportStorage(&aggregateReportStorage{files: map[string][]byte{}})

	_, err := svc.GetAggregateReportFile(context.Background(), reportID, userID)
	require.ErrorIs(t, err, models.ErrAggregateReportFileNotFound)
}

func aggregateReportAnalysis(userID uuid.UUID) models.AggregateAnalysis {
	return models.AggregateAnalysis{
		ID: uuid.New(), Scope: models.AggregateAnalysisScopePersonal,
		UserUUID: uuid.NullUUID{UUID: userID, Valid: true}, CreatedByUserUUID: userID,
		PeriodFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		PeriodTo:   time.Date(2026, 7, 7, 23, 59, 59, 0, time.UTC),
		Status:     models.AggregateAnalysisStatusDone, SourceCallsCount: 2,
		ResultJSON: []byte(`{"summary":"summary","priority_actions":["action"]}`),
	}
}

type aggregateReportAnalyticsRepo struct {
	analyticsRepoStub
	analysis models.AggregateAnalysis
}

func (r *aggregateReportAnalyticsRepo) GetAggregateAnalysisByUUID(_ context.Context, id uuid.UUID) (models.AggregateAnalysis, error) {
	if id != r.analysis.ID {
		return models.AggregateAnalysis{}, models.ErrAggregateAnalysisNotFound
	}
	return r.analysis, nil
}

type aggregateReportRepo struct {
	reports map[uuid.UUID]models.AggregateReportExport
}

func (r *aggregateReportRepo) ensure() {
	if r.reports == nil {
		r.reports = map[uuid.UUID]models.AggregateReportExport{}
	}
}

func (r *aggregateReportRepo) CreateAggregate(_ context.Context, report models.AggregateReportExport) (models.AggregateReportExport, error) {
	r.ensure()
	r.reports[report.ID] = report
	return report, nil
}

func (r *aggregateReportRepo) MarkAggregateReady(_ context.Context, input models.MarkAggregateReportReadyInput) (models.AggregateReportExport, error) {
	r.ensure()
	report := r.reports[input.ID]
	report.Status = models.ReportStatusReady
	report.StoragePath = &input.StoragePath
	report.FileName = input.FileName
	report.ContentType = input.ContentType
	report.SizeBytes = input.SizeBytes
	r.reports[input.ID] = report
	return report, nil
}

func (r *aggregateReportRepo) MarkAggregateFailed(_ context.Context, input models.MarkAggregateReportFailedInput) (models.AggregateReportExport, error) {
	r.ensure()
	report := r.reports[input.ID]
	report.Status = models.ReportStatusFailed
	report.ErrorMessage = &input.ErrorMessage
	r.reports[input.ID] = report
	return report, nil
}

func (r *aggregateReportRepo) GetAggregateByUUID(_ context.Context, id uuid.UUID) (models.AggregateReportExport, error) {
	r.ensure()
	report, ok := r.reports[id]
	if !ok {
		return models.AggregateReportExport{}, models.ErrAggregateReportNotFound
	}
	return report, nil
}

func (r *aggregateReportRepo) ListAggregateByAnalysisUUID(_ context.Context, analysisID uuid.UUID, _ time.Time) ([]models.AggregateReportExport, error) {
	r.ensure()
	out := []models.AggregateReportExport{}
	for _, report := range r.reports {
		if report.AggregateAnalysisUUID == analysisID {
			out = append(out, report)
		}
	}
	return out, nil
}

func (r *aggregateReportRepo) DeleteAggregate(_ context.Context, id uuid.UUID) error {
	r.ensure()
	if _, ok := r.reports[id]; !ok {
		return models.ErrAggregateReportNotFound
	}
	delete(r.reports, id)
	return nil
}

type aggregateReportStorage struct {
	files map[string][]byte
}

func (s *aggregateReportStorage) Save(_ context.Context, input models.SaveReportInput) (models.SavedReportFile, error) {
	if input.Content == nil || input.AggregateAnalysisUUID == uuid.Nil {
		return models.SavedReportFile{}, models.ErrInvalidReportInput
	}
	if s.files == nil {
		s.files = map[string][]byte{}
	}
	body, _ := io.ReadAll(input.Content)
	path := strings.Join([]string{"aggregate", input.AggregateAnalysisUUID.String(), input.ReportUUID.String()}, "/")
	s.files[path] = body
	return models.SavedReportFile{Path: path, MimeType: input.MimeType, SizeBytes: int64(len(body))}, nil
}

func (s *aggregateReportStorage) Open(_ context.Context, path string) (io.ReadCloser, error) {
	body, ok := s.files[path]
	if !ok {
		return nil, models.ErrReportFileNotFound
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}

func (s *aggregateReportStorage) Delete(_ context.Context, path string) error {
	if s.files == nil {
		return nil
	}
	delete(s.files, path)
	return nil
}
