package analysis

import (
	"calllens/monolit/internal/models"
	"context"
	"fmt"
	"io"
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

	transcription, err := s.transcriptionRepository.GetByCallUUID(ctx, call.ID)
	if err != nil {
		return models.CallAnalysis{}, fmt.Errorf("get transcription: %w", err)
	}

	if transcription.Status != models.TranscriptionStatusTranscribed || transcription.Text == nil {
		return models.CallAnalysis{}, models.ErrInvalidAnalysisStatus
	}

	instructions, err := s.loadInstructions(ctx, call, input.UserUUID)
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
		failedAnalysis, markErr := s.analysisRepository.MarkFailed(ctx, analysis.ID, err.Error())
		if markErr != nil {
			return models.CallAnalysis{}, fmt.Errorf("mark analysis failed: %w", markErr)
		}
		s.log.Error(ctx, "call analysis failed", zap.String("call_id", call.ID.String()), zap.Error(err))
		return failedAnalysis, fmt.Errorf("analyze call: %w", err)
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
