package report

import (
	"calllens/monolit/internal/models"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateGeneratesOnlyRequestedFormat(t *testing.T) {
	ctx := context.Background()
	callID := uuid.New()
	userID := uuid.New()
	analysisID := uuid.New()
	now := time.Date(2026, 6, 16, 10, 0, 0, 0, time.UTC)

	reports := &fakeReportRepository{}
	storage := &fakeReportStorage{}
	svc := NewService(
		&fakeCallRepository{call: testCall(callID)},
		&fakeAnalysisRepository{analysis: testAnalysis(callID, analysisID)},
		&fakeTranscriptionRepository{text: "Менеджер: Добрый день"},
		reports,
		storage,
	)
	svc.now = func() time.Time { return now }
	svc.SetBillingLimiter(&fakeBillingLimiter{
		subscription: models.Subscription{Plan: models.Plan{ExportEnabled: true}},
	})

	report, err := svc.Create(ctx, models.CreateReportInput{
		CallUUID: callID,
		UserUUID: userID,
		Format:   models.ReportFormatMD,
	})

	require.NoError(t, err)
	require.Equal(t, models.ReportStatusReady, report.Status)
	require.Equal(t, models.ReportFormatMD, storage.saved.Format)
	require.True(t, strings.HasSuffix(report.FileName, ".md"))
	require.Contains(t, storage.content, "# Отчет по звонку")
	require.Contains(t, storage.content, "## Вопросы клиента и ответы менеджера")
	require.Contains(t, storage.content, "## Качество менеджера")
	require.Len(t, reports.items, 1)
}

func TestCreateRejectsPersonalPlanWithoutExport(t *testing.T) {
	ctx := context.Background()
	callID := uuid.New()
	userID := uuid.New()

	svc := NewService(
		&fakeCallRepository{call: testCall(callID)},
		&fakeAnalysisRepository{analysis: testAnalysis(callID, uuid.New())},
		nil,
		&fakeReportRepository{},
		&fakeReportStorage{},
	)
	svc.SetBillingLimiter(&fakeBillingLimiter{
		subscription: models.Subscription{Plan: models.Plan{ExportEnabled: false}},
	})

	_, err := svc.Create(ctx, models.CreateReportInput{
		CallUUID: callID,
		UserUUID: userID,
		Format:   models.ReportFormatMD,
	})

	require.ErrorIs(t, err, models.ErrExportAccessDenied)
}

func TestDeleteRemovesStorageFileAndRepositoryRow(t *testing.T) {
	ctx := context.Background()
	reportID := uuid.New()
	callID := uuid.New()
	userID := uuid.New()
	path := "calls/report.md"

	reports := &fakeReportRepository{
		items: map[uuid.UUID]models.ReportExport{
			reportID: {
				ID:          reportID,
				CallUUID:    callID,
				Status:      models.ReportStatusReady,
				StoragePath: &path,
			},
		},
	}
	storage := &fakeReportStorage{}
	svc := NewService(&fakeCallRepository{call: testCall(callID)}, nil, nil, reports, storage)

	err := svc.Delete(ctx, reportID, userID)

	require.NoError(t, err)
	require.Equal(t, path, storage.deletedPath)
	require.Empty(t, reports.items)
}

func testCall(id uuid.UUID) models.Call {
	return models.Call{
		ID:              id,
		Title:           "Тестовый звонок",
		Status:          models.CallStatusAnalyzed,
		VisibilityScope: models.CallVisibilityScopePersonal,
		CreatedAt:       time.Date(2026, 6, 16, 9, 0, 0, 0, time.UTC),
	}
}

func testAnalysis(callID uuid.UUID, analysisID uuid.UUID) models.CallAnalysis {
	return models.CallAnalysis{
		ID:         analysisID,
		CallUUID:   callID,
		Status:     models.CallAnalysisStatusDone,
		Provider:   "test",
		ResultJSON: []byte(`{"summary":"Клиент забронировал квест.","topics":["Бронирование"],"dialogue_tone":{"overall":"нейтральный","manager":"вежливый","client":"нейтральный","evidence_quotes":["Да, да."]},"client_questions":[{"question":"Где находится квест?","manager_answer":"Рядом с метро.","answer_status":"answered","evidence_quotes":["Рядом с метро"]}],"question_coverage":{"status":"answered","summary":"Все вопросы закрыты.","unanswered_questions":[]},"manager_quality":{"strengths":["Менеджер был вежлив"],"issues":[],"recommendations":[]},"call_outcome":"Звонок завершился успешно.","score":90,"criteria_results":[{"instruction_title":"Приветствие","result":"Выполнено","evidence_quotes":["Здравствуйте"]}],"customer_objections":[],"risks":[],"next_steps":["Подтвердить бронь"],"next_step":"Подтвердить бронь","evidence_quotes":["Да, да."],"confidence":"high"}`),
		CreatedAt:  time.Date(2026, 6, 16, 9, 5, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 6, 16, 9, 6, 0, 0, time.UTC),
	}
}

type fakeCallRepository struct {
	call models.Call
}

func (f *fakeCallRepository) CreateCall(context.Context, models.Call) (models.Call, error) {
	return models.Call{}, nil
}
func (f *fakeCallRepository) CreateCallWithProcessingJob(context.Context, models.Call, models.ProcessingJob) (models.Call, error) {
	return models.Call{}, nil
}
func (f *fakeCallRepository) List(context.Context, uuid.UUID) ([]models.Call, error) { return nil, nil }
func (f *fakeCallRepository) GetByUUID(context.Context, uuid.UUID, uuid.UUID) (models.Call, error) {
	return f.call, nil
}
func (f *fakeCallRepository) GetByUUIDForProcessing(context.Context, uuid.UUID) (models.Call, error) {
	return models.Call{}, nil
}
func (f *fakeCallRepository) UpdateCallTitle(context.Context, uuid.UUID, uuid.UUID, string) (models.Call, error) {
	return models.Call{}, nil
}
func (f *fakeCallRepository) UpdateCallStatus(context.Context, uuid.UUID, models.CallStatus) (models.Call, error) {
	return models.Call{}, nil
}
func (f *fakeCallRepository) DeleteCall(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (f *fakeCallRepository) TakeNextForProcessing(context.Context) (models.Call, error) {
	return models.Call{}, nil
}

type fakeAnalysisRepository struct {
	analysis models.CallAnalysis
}

func (f *fakeAnalysisRepository) Create(context.Context, models.CallAnalysis) (models.CallAnalysis, error) {
	return models.CallAnalysis{}, nil
}
func (f *fakeAnalysisRepository) GetByCallUUID(context.Context, uuid.UUID) (models.CallAnalysis, error) {
	return f.analysis, nil
}
func (f *fakeAnalysisRepository) MarkProcessing(context.Context, uuid.UUID) (models.CallAnalysis, error) {
	return models.CallAnalysis{}, nil
}
func (f *fakeAnalysisRepository) MarkDone(context.Context, uuid.UUID, models.AnalysisResult) (models.CallAnalysis, error) {
	return models.CallAnalysis{}, nil
}
func (f *fakeAnalysisRepository) MarkFailed(context.Context, uuid.UUID, string) (models.CallAnalysis, error) {
	return models.CallAnalysis{}, nil
}

type fakeTranscriptionRepository struct {
	text string
}

func (f *fakeTranscriptionRepository) Create(context.Context, models.Transcription) (models.Transcription, error) {
	return models.Transcription{}, nil
}
func (f *fakeTranscriptionRepository) GetByCallUUID(context.Context, uuid.UUID) (models.Transcription, error) {
	return models.Transcription{Text: &f.text}, nil
}
func (f *fakeTranscriptionRepository) MarkTranscribed(context.Context, uuid.UUID, string, []models.TranscriptionSegment, *string) (models.Transcription, error) {
	return models.Transcription{}, nil
}
func (f *fakeTranscriptionRepository) MarkFailed(context.Context, uuid.UUID, string) (models.Transcription, error) {
	return models.Transcription{}, nil
}

type fakeReportRepository struct {
	items map[uuid.UUID]models.ReportExport
}

func (f *fakeReportRepository) Create(_ context.Context, report models.ReportExport) (models.ReportExport, error) {
	if f.items == nil {
		f.items = make(map[uuid.UUID]models.ReportExport)
	}
	f.items[report.ID] = report
	return report, nil
}
func (f *fakeReportRepository) MarkReady(_ context.Context, input models.MarkReportReadyInput) (models.ReportExport, error) {
	report := f.items[input.ID]
	report.Status = models.ReportStatusReady
	report.StoragePath = &input.StoragePath
	report.FileName = input.FileName
	report.ContentType = input.ContentType
	report.SizeBytes = input.SizeBytes
	f.items[input.ID] = report
	return report, nil
}
func (f *fakeReportRepository) MarkFailed(_ context.Context, input models.MarkReportFailedInput) (models.ReportExport, error) {
	report := f.items[input.ID]
	report.Status = models.ReportStatusFailed
	report.ErrorMessage = &input.ErrorMessage
	f.items[input.ID] = report
	return report, nil
}
func (f *fakeReportRepository) GetByUUID(_ context.Context, id uuid.UUID) (models.ReportExport, error) {
	return f.items[id], nil
}
func (f *fakeReportRepository) ListByCallUUID(context.Context, uuid.UUID, time.Time) ([]models.ReportExport, error) {
	return nil, nil
}
func (f *fakeReportRepository) ListExpiredReady(context.Context, time.Time, int) ([]models.ReportExport, error) {
	return nil, nil
}
func (f *fakeReportRepository) Delete(_ context.Context, id uuid.UUID) error {
	delete(f.items, id)
	return nil
}

type fakeReportStorage struct {
	saved       models.SaveReportInput
	content     string
	deletedPath string
}

func (f *fakeReportStorage) Save(_ context.Context, input models.SaveReportInput) (models.SavedReportFile, error) {
	f.saved = input
	content, _ := io.ReadAll(input.Content)
	f.content = string(content)
	return models.SavedReportFile{
		Path:      "reports/" + input.ReportUUID.String() + fileExtension(input.Format),
		MimeType:  input.MimeType,
		SizeBytes: int64(len(content)),
	}, nil
}
func (f *fakeReportStorage) Open(context.Context, string) (io.ReadCloser, error) { return nil, nil }
func (f *fakeReportStorage) Delete(_ context.Context, path string) error {
	f.deletedPath = path
	return nil
}

type fakeBillingLimiter struct {
	subscription models.Subscription
}

func (f *fakeBillingLimiter) CanExportReports(context.Context, uuid.UUID) error { return nil }
func (f *fakeBillingLimiter) GetPersonalSubscription(context.Context, uuid.UUID) (models.Subscription, error) {
	return f.subscription, nil
}
