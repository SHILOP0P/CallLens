package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	analyzerMocks "calllens/monolit/internal/analyzer/mocks"
	"calllens/monolit/internal/models"
	repositoryMocks "calllens/monolit/internal/repository/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestGetByCallUUID(t *testing.T) {
	callID := uuid.New()
	userID := uuid.New()
	callRepo := repositoryMocks.NewCallRepository(t)
	analysisRepo := repositoryMocks.NewAnalysisRepository(t)
	existing := models.CallAnalysis{ID: uuid.New(), CallUUID: callID}
	callRepo.EXPECT().GetByUUID(mock.Anything, callID, userID).Return(models.Call{ID: callID}, nil).Once()
	analysisRepo.EXPECT().GetByCallUUID(mock.Anything, callID).Return(existing, nil).Once()
	service := NewService(callRepo, nil, nil, analysisRepo, nil, nil, nil)

	got, err := service.GetByCallUUID(context.Background(), callID, userID)
	if err != nil || got.ID != existing.ID {
		t.Fatalf("GetByCallUUID = %+v, %v", got, err)
	}
	if _, err := service.GetByCallUUID(context.Background(), uuid.Nil, userID); !errors.Is(err, models.ErrInvalidAnalysisInput) {
		t.Fatalf("invalid input error = %v", err)
	}
}

func TestAnalyzeCallValidationAndStatus(t *testing.T) {
	service := NewService(nil, nil, nil, nil, nil, nil, nil)
	if _, err := service.AnalyzeCall(context.Background(), models.AnalyzeCallInput{}); !errors.Is(err, models.ErrInvalidAnalysisInput) {
		t.Fatalf("invalid input error = %v", err)
	}

	callID := uuid.New()
	userID := uuid.New()
	callRepo := repositoryMocks.NewCallRepository(t)
	transcriptionRepo := repositoryMocks.NewTranscriptionRepository(t)
	callRepo.EXPECT().GetByUUID(mock.Anything, callID, userID).Return(models.Call{ID: callID}, nil).Once()
	transcriptionRepo.EXPECT().GetByCallUUID(mock.Anything, callID).Return(models.Transcription{
		CallUUID: callID, Status: models.TranscriptionStatusProcessing,
	}, nil).Once()
	service = NewService(callRepo, transcriptionRepo, nil, repositoryMocks.NewAnalysisRepository(t), nil, nil, nil)
	if _, err := service.AnalyzeCall(context.Background(), models.AnalyzeCallInput{CallUUID: callID, UserUUID: userID}); !errors.Is(err, models.ErrInvalidAnalysisStatus) {
		t.Fatalf("status error = %v", err)
	}
}

func TestProcessAnalyzeCallValidation(t *testing.T) {
	service := NewService(nil, nil, nil, nil, nil, nil, nil)
	if err := service.ProcessAnalyzeCall(context.Background(), uuid.Nil); !errors.Is(err, models.ErrCallNotFound) {
		t.Fatalf("nil call error = %v", err)
	}
	if err := service.ProcessAnalyzeCall(context.Background(), uuid.New()); !errors.Is(err, models.ErrAnalyzerNotConfigured) {
		t.Fatalf("missing analyzer error = %v", err)
	}

	callID := uuid.New()
	callRepo := repositoryMocks.NewCallRepository(t)
	analyzer := analyzerMocks.NewAnalyzer(t)
	callRepo.EXPECT().GetByUUIDForProcessing(mock.Anything, callID).
		Return(models.Call{ID: callID, Status: models.CallStatusAnalyzed}, nil).Once()
	service = NewService(callRepo, nil, nil, nil, nil, analyzer, nil)
	if err := service.ProcessAnalyzeCall(context.Background(), callID); err != nil {
		t.Fatalf("already analyzed: %v", err)
	}

	callRepo = repositoryMocks.NewCallRepository(t)
	callRepo.EXPECT().GetByUUIDForProcessing(mock.Anything, callID).
		Return(models.Call{ID: callID, Status: models.CallStatusTranscribed}, nil).Once()
	service = NewService(callRepo, nil, nil, nil, nil, analyzerMocks.NewAnalyzer(t), nil)
	if err := service.ProcessAnalyzeCall(context.Background(), callID); !errors.Is(err, models.ErrInvalidAnalysisInput) {
		t.Fatalf("missing uploader error = %v", err)
	}
}

func TestMarkAnalyzeCallFailedCreatesMissingAnalysis(t *testing.T) {
	callID := uuid.New()
	callRepo := repositoryMocks.NewCallRepository(t)
	analysisRepo := repositoryMocks.NewAnalysisRepository(t)
	analysisRepo.EXPECT().GetByCallUUID(mock.Anything, callID).
		Return(models.CallAnalysis{}, models.ErrAnalysisNotFound).Once()
	analysisRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(value models.CallAnalysis) bool {
		return value.CallUUID == callID && value.Status == models.CallAnalysisStatusFailed
	})).Return(models.CallAnalysis{ID: uuid.New(), CallUUID: callID, Status: models.CallAnalysisStatusFailed}, nil).Once()
	callRepo.EXPECT().UpdateCallStatus(mock.Anything, callID, models.CallStatusFailed).
		Return(models.Call{ID: callID, Status: models.CallStatusFailed}, nil).Once()
	service := NewService(callRepo, nil, nil, analysisRepo, nil, nil, nil)
	if err := service.MarkAnalyzeCallFailed(context.Background(), callID, nil); err != nil {
		t.Fatalf("MarkAnalyzeCallFailed: %v", err)
	}
	if err := service.MarkAnalyzeCallFailed(context.Background(), uuid.Nil, nil); !errors.Is(err, models.ErrCallNotFound) {
		t.Fatalf("nil call error = %v", err)
	}
}

func TestNormalizationHelpers(t *testing.T) {
	text := "summary"
	result, err := normalizeAnalysisResult(models.AnalysisResult{ResultText: &text})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(result.ResultJSON, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["summary"] != text || payload["confidence"] != "low" {
		t.Fatalf("payload = %+v", payload)
	}
	if _, err := normalizeAnalysisResult(models.AnalysisResult{ResultJSON: []byte("{")}); err == nil {
		t.Fatal("expected invalid JSON error")
	}

	values := []any{1, " ", "next"}
	if firstStringFromArray(values) != "next" || firstStringFromArray("bad") != "" {
		t.Fatal("firstStringFromArray mismatch")
	}
	object := map[string]any{"existing": "value"}
	ensureObjectFields(object, "nested", map[string]any{"x": 1})
	ensureObjectFields(object, "nested", map[string]any{"y": 2})
	ensureStringField(object, "string")
	ensureNumberField(object, "number")
	ensureConfidenceField(object)
}

func TestNormalizeAnalysisResultRewritesKnownEnglishFallbacks(t *testing.T) {
	text := "The transcription provided does not contain a sales or client call. It is a text about the history and new directions of advertising, including the use of human billboards. Therefore, no analysis of a sales or client call can be provided."
	result, err := normalizeAnalysisResult(models.AnalysisResult{
		ResultJSON: []byte(`{
			"summary":"The transcription provided does not contain a sales or client call. It is a text about the history and new directions of advertising, including the use of human billboards. Therefore, no analysis of a sales or client call can be provided.",
			"dialogue_tone":{"overall":"unclear","manager":"unclear","client":"unclear","evidence_quotes":[]},
			"question_coverage":{"status":"unclear","summary":"No client questions were identified in the transcription.","unanswered_questions":[]},
			"call_outcome":"unclear",
			"next_steps":["unclear"],
			"confidence":"low"
		}`),
		ResultText: &text,
	})
	if err != nil {
		t.Fatal(err)
	}

	var payload map[string]any
	if err := json.Unmarshal(result.ResultJSON, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["summary"] != "Расшифровка не содержит продажного или клиентского звонка. Это текст об истории и новых направлениях рекламы, включая использование людей-рекламоносителей, поэтому анализ диалога с клиентом выполнить нельзя." {
		t.Fatalf("summary = %q", payload["summary"])
	}
	if result.ResultText == nil || strings.Contains(*result.ResultText, "The transcription") {
		t.Fatalf("result text was not normalized: %v", result.ResultText)
	}
	dialogueTone := payload["dialogue_tone"].(map[string]any)
	if dialogueTone["overall"] != "Неясно" {
		t.Fatalf("dialogue tone = %#v", dialogueTone)
	}
	questionCoverage := payload["question_coverage"].(map[string]any)
	if questionCoverage["status"] != "unclear" || questionCoverage["summary"] != "В расшифровке не выявлены вопросы клиента." {
		t.Fatalf("question coverage = %#v", questionCoverage)
	}
	if payload["call_outcome"] != "Неясно" {
		t.Fatalf("call outcome = %q", payload["call_outcome"])
	}
	nextSteps := payload["next_steps"].([]any)
	if len(nextSteps) != 1 || nextSteps[0] != "Неясно" {
		t.Fatalf("next steps = %#v", nextSteps)
	}
}

func TestServiceConfigurationAndInstructionSelection(t *testing.T) {
	instructionRepo := repositoryMocks.NewAnalysisInstructionRepository(t)
	service := NewService(nil, nil, instructionRepo, nil, nil, nil, nil)
	service.SetProcessingJobMaxAttempts(0)
	if service.processingJobMaxAttempts != models.DefaultProcessingJobMaxAttempts {
		t.Fatalf("max attempts = %d", service.processingJobMaxAttempts)
	}
	if service.analyzerProviderName() != "unknown" {
		t.Fatalf("provider = %q", service.analyzerProviderName())
	}
	if _, err := service.selectInstructions(context.Background(), models.Call{VisibilityScope: "invalid"}, uuid.New()); !errors.Is(err, models.ErrInvalidAnalysisInput) {
		t.Fatalf("selection error = %v", err)
	}
	instructionRepo.EXPECT().List(mock.Anything, mock.Anything).Return(nil, nil).Twice()
	if _, err := service.selectInstructions(context.Background(), models.Call{VisibilityScope: models.CallVisibilityScopePersonal}, uuid.New()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.selectInstructions(context.Background(), models.Call{VisibilityScope: models.CallVisibilityScopeCompany}, uuid.New()); err != nil {
		t.Fatal(err)
	}
}
