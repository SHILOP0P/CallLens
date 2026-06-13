package analysis

import (
	"calllens/monolit/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) AnalyzeCall(ctx context.Context, input models.AnalyzeCallInput) (models.CallAnalysis, error) {
	if input.CallUUID == uuid.Nil || input.UserUUID == uuid.Nil {
		return models.CallAnalysis{}, models.ErrInvalidAnalysisInput
	}

	if s.analyzer == nil {
		return models.CallAnalysis{}, models.ErrAnalyzerNotConfigured
	}

	call, err := s.callRepository.GetByUUID(ctx, input.CallUUID, input.UserUUID)
	if err != nil {
		return models.CallAnalysis{}, fmt.Errorf("get call: %w", err)
	}

	return s.analyzeCall(ctx, call, input.UserUUID, analyzeCallOptions{
		markAttemptFailed: true,
	})
}

func (s *Service) ProcessAnalyzeCall(ctx context.Context, callID uuid.UUID) error {
	if callID == uuid.Nil {
		return models.ErrCallNotFound
	}

	call, err := s.callRepository.GetByUUIDForProcessing(ctx, callID)
	if err != nil {
		return fmt.Errorf("get call for analysis processing: %w", err)
	}

	if call.Status == models.CallStatusAnalyzed {
		s.log.Info(ctx, "call already analyzed", zap.String("call_id", call.ID.String()))
		return nil
	}

	if !call.UploadedByUserUUID.Valid {
		return models.ErrInvalidAnalysisInput
	}

	_, err = s.analyzeCall(ctx, call, call.UploadedByUserUUID.UUID, analyzeCallOptions{
		markAttemptFailed: false,
	})
	return err
}

func (s *Service) MarkAnalyzeCallFailed(ctx context.Context, callID uuid.UUID, cause error) error {
	if callID == uuid.Nil {
		return models.ErrCallNotFound
	}

	errorMessage := "analysis failed"
	if cause != nil && strings.TrimSpace(cause.Error()) != "" {
		errorMessage = cause.Error()
	}

	analysis, err := s.analysisRepository.GetByCallUUID(ctx, callID)
	if err != nil {
		if !errors.Is(err, models.ErrAnalysisNotFound) {
			return fmt.Errorf("get analysis for failure: %w", err)
		}

		failedAnalysis, createErr := s.createFailedAnalysis(ctx, callID, errorMessage)
		if createErr != nil {
			return fmt.Errorf("create failed analysis: %w", createErr)
		}
		analysis = failedAnalysis
	} else {
		failedAnalysis, markErr := s.analysisRepository.MarkFailed(ctx, analysis.ID, errorMessage)
		if markErr != nil {
			return fmt.Errorf("mark analysis failed: %w", markErr)
		}
		analysis = failedAnalysis
	}

	s.log.Error(ctx, "call analysis permanently failed", zap.String("call_id", callID.String()), zap.String("analysis_id", analysis.ID.String()), zap.Error(cause))

	return nil
}

type analyzeCallOptions struct {
	markAttemptFailed bool
}

func (s *Service) analyzeCall(ctx context.Context, call models.Call, userID uuid.UUID, opts analyzeCallOptions) (models.CallAnalysis, error) {
	transcription, err := s.transcriptionRepository.GetByCallUUID(ctx, call.ID)
	if err != nil {
		return models.CallAnalysis{}, fmt.Errorf("get transcription: %w", err)
	}

	if transcription.Status != models.TranscriptionStatusTranscribed || transcription.Text == nil {
		return models.CallAnalysis{}, models.ErrInvalidAnalysisStatus
	}

	instructions, err := s.loadInstructions(ctx, call, userID)
	if err != nil {
		return models.CallAnalysis{}, fmt.Errorf("load instructions: %w", err)
	}

	analysis, err := s.createPendingAnalysis(ctx, call.ID)
	if err != nil {
		return models.CallAnalysis{}, fmt.Errorf("create analysis: %w", err)
	}

	analysis, err = s.analysisRepository.MarkProcessing(ctx, analysis.ID)
	if err != nil {
		return models.CallAnalysis{}, fmt.Errorf("mark analysis processing: %w", err)
	}

	result, err := s.analyzer.Analyze(ctx, models.AnalysisRequest{
		CallUUID:      call.ID,
		Transcription: *transcription.Text,
		Instructions:  instructions,
	})
	if err != nil {
		if opts.markAttemptFailed {
			failedAnalysis, markErr := s.analysisRepository.MarkFailed(ctx, analysis.ID, err.Error())
			if markErr != nil {
				return models.CallAnalysis{}, fmt.Errorf("mark analysis failed: %w", markErr)
			}
			analysis = failedAnalysis
		}
		s.log.Error(ctx, "call analysis failed", zap.String("call_id", call.ID.String()), zap.Error(err))
		return analysis, fmt.Errorf("analyze call: %w", err)
	}

	result, err = normalizeAnalysisResult(result)
	if err != nil {
		if opts.markAttemptFailed {
			failedAnalysis, markErr := s.analysisRepository.MarkFailed(ctx, analysis.ID, err.Error())
			if markErr != nil {
				return models.CallAnalysis{}, fmt.Errorf("mark analysis failed: %w", markErr)
			}
			analysis = failedAnalysis
		}
		s.log.Error(ctx, "call analysis result is invalid", zap.String("call_id", call.ID.String()), zap.Error(err))
		return analysis, fmt.Errorf("normalize analysis result: %w", err)
	}

	analysis, err = s.analysisRepository.MarkDone(ctx, analysis.ID, result)
	if err != nil {
		return models.CallAnalysis{}, fmt.Errorf("mark analysis done: %w", err)
	}

	if _, err = s.callRepository.UpdateCallStatus(ctx, call.ID, models.CallStatusAnalyzed); err != nil {
		return models.CallAnalysis{}, fmt.Errorf("mark call analyzed: %w", err)
	}

	s.log.Info(ctx, "call analyzed", zap.String("call_id", call.ID.String()), zap.String("provider", s.analyzer.Provider()))

	return analysis, nil
}

func (s *Service) createPendingAnalysis(ctx context.Context, callID uuid.UUID) (models.CallAnalysis, error) {
	analysisID, err := uuid.NewV7()
	if err != nil {
		return models.CallAnalysis{}, err
	}

	now := time.Now().UTC()

	return s.analysisRepository.Create(ctx, models.CallAnalysis{
		ID:        analysisID,
		CallUUID:  callID,
		Status:    models.CallAnalysisStatusPending,
		Provider:  s.analyzer.Provider(),
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) createFailedAnalysis(ctx context.Context, callID uuid.UUID, errorMessage string) (models.CallAnalysis, error) {
	analysisID, err := uuid.NewV7()
	if err != nil {
		return models.CallAnalysis{}, err
	}

	now := time.Now().UTC()

	return s.analysisRepository.Create(ctx, models.CallAnalysis{
		ID:           analysisID,
		CallUUID:     callID,
		Status:       models.CallAnalysisStatusFailed,
		Provider:     s.analyzerProviderName(),
		ErrorMessage: &errorMessage,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (s *Service) analyzerProviderName() string {
	if s.analyzer == nil {
		return "unknown"
	}

	return s.analyzer.Provider()
}

func (s *Service) loadInstructions(ctx context.Context, call models.Call, userID uuid.UUID) ([]models.AnalysisInstructionContent, error) {
	instructions, err := s.selectInstructions(ctx, call, userID)
	if err != nil {
		return nil, err
	}

	contents := make([]models.AnalysisInstructionContent, 0, len(instructions))
	for _, instruction := range instructions {
		content, err := s.readInstructionContent(ctx, instruction)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return contents, nil
}

func (s *Service) selectInstructions(ctx context.Context, call models.Call, userID uuid.UUID) ([]models.AnalysisInstruction, error) {
	switch call.VisibilityScope {
	case models.CallVisibilityScopePersonal:
		return s.instructionRepository.List(ctx, models.ListAnalysisInstructionsInput{
			Scope:    models.AnalysisInstructionScopePersonal,
			UserUUID: userID,
		})
	case models.CallVisibilityScopeCompany:
		return s.instructionRepository.List(ctx, models.ListAnalysisInstructionsInput{
			Scope:       models.AnalysisInstructionScopeCompany,
			CompanyUUID: call.CompanyUUID,
		})
	case models.CallVisibilityScopeDepartment:
		companyInstructions, err := s.instructionRepository.List(ctx, models.ListAnalysisInstructionsInput{
			Scope:       models.AnalysisInstructionScopeCompany,
			CompanyUUID: call.CompanyUUID,
		})
		if err != nil {
			return nil, err
		}

		departmentInstructions, err := s.instructionRepository.List(ctx, models.ListAnalysisInstructionsInput{
			Scope:          models.AnalysisInstructionScopeDepartment,
			CompanyUUID:    call.CompanyUUID,
			DepartmentUUID: call.DepartmentUUID,
		})
		if err != nil {
			return nil, err
		}

		return append(companyInstructions, departmentInstructions...), nil
	default:
		return nil, models.ErrInvalidAnalysisInput
	}
}

func (s *Service) readInstructionContent(ctx context.Context, instruction models.AnalysisInstruction) (models.AnalysisInstructionContent, error) {
	content, err := s.instructionStorage.Open(ctx, instruction.FilePath)
	if err != nil {
		return models.AnalysisInstructionContent{}, err
	}
	defer content.Close()

	data, err := io.ReadAll(content)
	if err != nil {
		return models.AnalysisInstructionContent{}, err
	}

	return models.AnalysisInstructionContent{
		ID:      instruction.ID,
		Scope:   instruction.Scope,
		Title:   instruction.Title,
		Content: string(data),
	}, nil
}

func normalizeAnalysisResult(result models.AnalysisResult) (models.AnalysisResult, error) {
	payload := map[string]any{}

	if len(result.ResultJSON) > 0 {
		if err := json.Unmarshal(result.ResultJSON, &payload); err != nil {
			return models.AnalysisResult{}, fmt.Errorf("decode analysis result json: %w", err)
		}
	}

	resultText := ""
	if result.ResultText != nil {
		resultText = strings.TrimSpace(*result.ResultText)
	}

	summary := stringField(payload, "summary")
	if summary == "" {
		summary = resultText
	}
	if summary == "" {
		summary = "Analysis completed, but provider returned no summary."
	}
	payload["summary"] = summary

	if _, ok := payload["topics"]; !ok {
		payload["topics"] = []any{}
	}

	if stringField(payload, "next_step") == "" {
		payload["next_step"] = firstStringFromArray(payload["next_steps"])
	}

	if resultText == "" {
		resultText = summary
		result.ResultText = &resultText
	}

	resultJSON, err := json.Marshal(payload)
	if err != nil {
		return models.AnalysisResult{}, fmt.Errorf("encode normalized analysis result json: %w", err)
	}

	result.ResultJSON = resultJSON

	return result, nil
}

func stringField(payload map[string]any, key string) string {
	value, ok := payload[key].(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(value)
}

func firstStringFromArray(value any) string {
	values, ok := value.([]any)
	if !ok {
		return ""
	}

	for _, item := range values {
		text, ok := item.(string)
		if !ok {
			continue
		}
		text = strings.TrimSpace(text)
		if text != "" {
			return text
		}
	}

	return ""
}
