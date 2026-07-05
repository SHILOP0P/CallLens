package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) AnalyzeCall(ctx context.Context, input models.AnalyzeCallInput) (models.CallAnalysis, error) {
	if input.CallUUID == uuid.Nil || input.UserUUID == uuid.Nil {
		return models.CallAnalysis{}, models.ErrInvalidAnalysisInput
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

	analysis, err := s.createPendingAnalysis(ctx, call.ID)
	if err != nil {
		return models.CallAnalysis{}, fmt.Errorf("create analysis: %w", err)
	}

	if err = s.enqueueAnalyzeJob(ctx, call.ID); err != nil {
		return models.CallAnalysis{}, fmt.Errorf("enqueue analysis job: %w", err)
	}

	s.log.Info(ctx, "call analysis job enqueued", zap.String("call_id", call.ID.String()), zap.String("analysis_id", analysis.ID.String()))

	return analysis, nil
}

func (s *Service) ProcessAnalyzeCall(ctx context.Context, callID uuid.UUID) error {
	if callID == uuid.Nil {
		return models.ErrCallNotFound
	}

	if s.analyzer == nil {
		return models.ErrAnalyzerNotConfigured
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

	if _, err := s.callRepository.UpdateCallStatus(ctx, callID, models.CallStatusFailed); err != nil {
		return fmt.Errorf("mark call failed: %w", err)
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
		Provider:  s.analyzerProviderName(),
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) enqueueAnalyzeJob(ctx context.Context, callID uuid.UUID) error {
	if s.processingJobRepository == nil {
		return models.ErrProcessingJobNotFound
	}

	jobID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate analysis job uuid: %w", err)
	}

	now := time.Now().UTC()

	_, err = s.processingJobRepository.Enqueue(ctx, models.ProcessingJob{
		ID:          jobID,
		Type:        models.ProcessingJobTypeAnalyzeCall,
		EntityUUID:  callID,
		Status:      models.ProcessingJobStatusPending,
		Attempts:    0,
		MaxAttempts: s.processingJobMaxAttempts,
		AvailableAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return err
	}

	return nil
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
	if call.SkipCustomInstructions {
		return []models.AnalysisInstructionContent{}, nil
	}

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
	defer func() { _ = content.Close() }()

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
		summary = "Анализ завершен, но провайдер не вернул резюме."
	}
	summary = normalizeRussianAnalysisText(summary)
	payload["summary"] = summary

	payload["schema_version"] = float64(2)
	ensureArrayField(payload, "topics")
	ensureObjectFields(payload, "dialogue_tone", map[string]any{
		"overall":         "",
		"manager":         "",
		"client":          "",
		"evidence_quotes": []any{},
	})
	ensureArrayField(payload, "client_questions")
	ensureObjectFields(payload, "question_coverage", map[string]any{
		"status":               "unclear",
		"summary":              "",
		"unanswered_questions": []any{},
	})
	ensureObjectFields(payload, "manager_quality", map[string]any{
		"strengths":       []any{},
		"issues":          []any{},
		"recommendations": []any{},
	})
	ensureStringField(payload, "call_outcome")
	ensureArrayField(payload, "customer_objections")
	ensureArrayField(payload, "risks")
	ensureArrayField(payload, "next_steps")
	ensureObjectFields(payload, "next_step_quality", map[string]any{
		"has_next_step":          false,
		"specific":               false,
		"has_deadline":           false,
		"has_responsible_person": false,
	})
	ensureObjectFields(payload, "business_outcome", map[string]any{
		"status":      "unclear",
		"summary":     "",
		"lost_reason": "not_applicable",
	})
	ensureObjectFields(payload, "customer_signals", map[string]any{
		"intent":                 "unclear",
		"urgency":                "unclear",
		"budget_discussed":       false,
		"decision_maker_present": false,
	})
	normalizeBusinessOutcome(payload)
	normalizeCustomerSignals(payload)
	normalizeNextStepQuality(payload)
	normalizeIssueCodes(payload)
	ensureArrayField(payload, "evidence_quotes")
	ensureConfidenceField(payload)
	normalizeCriteriaAndScore(payload)

	if stringField(payload, "next_step") == "" {
		payload["next_step"] = firstStringFromArray(payload["next_steps"])
	}
	normalizePayloadRussianText(payload)
	summary = stringField(payload, "summary")

	if resultText == "" || isKnownEnglishAnalysisFallback(resultText) {
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

func normalizeCriteriaAndScore(payload map[string]any) {
	inputScore := normalizeScore(payload["score"], payload["score_scale"])
	payload["score_scale"] = float64(100)

	rawCriteria, _ := payload["criteria_results"].([]any)
	criteria := make([]any, 0, len(rawCriteria))
	pointsAwarded := 0.0
	pointsPossible := 0.0
	applicableCount := 0

	for _, raw := range rawCriteria {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		normalized := normalizeCriterionResult(item)
		criteria = append(criteria, normalized)

		status := stringField(normalized, "status")
		pointsMax := numberField(normalized, "points_max")
		if status == "not_applicable" || pointsMax <= 0 {
			continue
		}
		pointsAwarded += numberField(normalized, "points_awarded")
		pointsPossible += pointsMax
		applicableCount++
	}

	payload["criteria_results"] = criteria
	score := inputScore
	if pointsPossible > 0 {
		score = clampScore(math.Round(pointsAwarded / pointsPossible * 100))
	}
	payload["score"] = score
	payload["score_breakdown"] = map[string]any{
		"points_awarded":            pointsAwarded,
		"points_possible":           pointsPossible,
		"applicable_criteria_count": float64(applicableCount),
		"total_criteria_count":      float64(len(criteria)),
	}
}

func normalizeCriterionResult(item map[string]any) map[string]any {
	out := copyStringMap(item)
	code := stringField(out, "code")
	legacyInstructionCriterion := code == "" && (stringField(out, "instruction_title") != "" || stringField(out, "result") != "")
	if legacyInstructionCriterion {
		code = "custom_instruction_match"
	}
	if code == "" {
		code = "custom_instruction_match"
	}
	out["code"] = code

	criterion, known := analysisCriterionByCode(code)
	if stringField(out, "title") == "" {
		if known {
			out["title"] = criterion.Title
		} else if title := stringField(out, "instruction_title"); title != "" {
			out["title"] = title
		} else {
			out["title"] = code
		}
	}

	status := stringField(out, "status")
	if status == "" && stringField(out, "result") != "" {
		status = "unclear"
	}
	out["status"] = normalizeCriterionStatus(status)

	if _, ok := out["points_awarded"]; !ok {
		out["points_awarded"] = float64(0)
	} else {
		out["points_awarded"] = math.Max(0, numberField(out, "points_awarded"))
	}
	if _, ok := out["points_max"]; !ok {
		if legacyInstructionCriterion {
			out["points_max"] = float64(0)
		} else if known {
			out["points_max"] = float64(criterion.PointsMax)
		} else {
			out["points_max"] = float64(0)
		}
	} else {
		out["points_max"] = math.Max(0, numberField(out, "points_max"))
	}

	ensureArrayField(out, "evidence_quotes")
	ensureStringField(out, "issue")
	ensureStringField(out, "recommendation")
	return out
}

func normalizeCriterionStatus(status string) string {
	switch status {
	case "met", "partially_met", "missed", "not_applicable", "unclear":
		return status
	default:
		return "unclear"
	}
}

func normalizeScore(scoreValue, scaleValue any) float64 {
	score, ok := numericValue(scoreValue)
	if !ok {
		return 0
	}
	if scale, ok := numericValue(scaleValue); ok && scale > 0 {
		return clampScore(math.Round(score / scale * 100))
	}
	switch {
	case score <= 5:
		return clampScore(math.Round(score * 20))
	case score <= 10:
		return clampScore(math.Round(score * 10))
	default:
		return clampScore(math.Round(score))
	}
}

func clampScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func normalizeBusinessOutcome(payload map[string]any) {
	value, _ := payload["business_outcome"].(map[string]any)
	if !allowedString(value, "status", []string{"success", "follow_up_needed", "no_decision", "lost", "support_resolved", "not_call", "unclear"}) {
		value["status"] = "unclear"
	}
	if !allowedString(value, "lost_reason", []string{"price", "timing", "no_need", "competitor", "no_next_step", "unclear_value", "bad_fit", "not_applicable", "unclear"}) {
		value["lost_reason"] = "not_applicable"
	}
}

func normalizeCustomerSignals(payload map[string]any) {
	value, _ := payload["customer_signals"].(map[string]any)
	for _, key := range []string{"intent", "urgency"} {
		if !allowedString(value, key, []string{"high", "medium", "low", "unclear"}) {
			value[key] = "unclear"
		}
	}
	ensureBoolField(value, "budget_discussed")
	ensureBoolField(value, "decision_maker_present")
}

func normalizeNextStepQuality(payload map[string]any) {
	value, _ := payload["next_step_quality"].(map[string]any)
	for _, key := range []string{"has_next_step", "specific", "has_deadline", "has_responsible_person"} {
		ensureBoolField(value, key)
	}
}

func normalizeIssueCodes(payload map[string]any) {
	values, ok := payload["issue_codes"].([]any)
	if !ok {
		payload["issue_codes"] = []any{}
		return
	}
	out := make([]any, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			continue
		}
		text = strings.TrimSpace(text)
		if text != "" {
			out = append(out, text)
		}
	}
	payload["issue_codes"] = out
}

func normalizePayloadRussianText(value any) {
	normalizePayloadRussianTextValue("", value)
}

func normalizePayloadRussianTextValue(key string, value any) {
	switch typed := value.(type) {
	case map[string]any:
		for childKey, childValue := range typed {
			switch childTyped := childValue.(type) {
			case string:
				if isSchemaEnumKey(childKey) {
					continue
				}
				typed[childKey] = normalizeRussianAnalysisText(childTyped)
			default:
				normalizePayloadRussianTextValue(childKey, childValue)
			}
		}
	case []any:
		for i, item := range typed {
			switch childTyped := item.(type) {
			case string:
				if isSchemaEnumKey(key) {
					continue
				}
				typed[i] = normalizeRussianAnalysisText(childTyped)
			default:
				normalizePayloadRussianTextValue(key, item)
			}
		}
	}
}

func isSchemaEnumKey(key string) bool {
	switch key {
	case "answer_status", "confidence", "status", "code", "lost_reason", "intent", "urgency":
		return true
	default:
		return false
	}
}

func normalizeRussianAnalysisText(value string) string {
	trimmed := strings.TrimSpace(value)
	switch strings.ToLower(trimmed) {
	case "":
		return ""
	case "unclear":
		return "Неясно"
	case "not specified", "not provided", "none", "n/a":
		return "Не указано"
	case "no client questions were identified in the transcription.":
		return "В расшифровке не выявлены вопросы клиента."
	case "the transcription provided does not contain a sales or client call. it is a text about the history and new directions of advertising, including the use of human billboards. therefore, no analysis of a sales or client call can be provided.":
		return "Расшифровка не содержит продажного или клиентского звонка. Это текст об истории и новых направлениях рекламы, включая использование людей-рекламоносителей, поэтому анализ диалога с клиентом выполнить нельзя."
	default:
		return trimmed
	}
}

func isKnownEnglishAnalysisFallback(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "unclear",
		"not specified",
		"not provided",
		"no client questions were identified in the transcription.",
		"the transcription provided does not contain a sales or client call. it is a text about the history and new directions of advertising, including the use of human billboards. therefore, no analysis of a sales or client call can be provided.":
		return true
	default:
		return false
	}
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

func ensureArrayField(payload map[string]any, key string) {
	if _, ok := payload[key].([]any); ok {
		return
	}

	payload[key] = []any{}
}

func ensureObjectFields(payload map[string]any, key string, defaults map[string]any) {
	value, ok := payload[key].(map[string]any)
	if !ok {
		payload[key] = defaults
		return
	}

	for defaultKey, defaultValue := range defaults {
		if _, exists := value[defaultKey]; !exists {
			value[defaultKey] = defaultValue
		}
	}
}

func ensureStringField(payload map[string]any, key string) {
	if _, ok := payload[key].(string); ok {
		return
	}

	payload[key] = ""
}

func ensureNumberField(payload map[string]any, key string) {
	switch payload[key].(type) {
	case float64, int, int64:
		return
	default:
		payload[key] = 0
	}
}

func ensureBoolField(payload map[string]any, key string) {
	if _, ok := payload[key].(bool); ok {
		return
	}
	payload[key] = false
}

func ensureConfidenceField(payload map[string]any) {
	switch stringField(payload, "confidence") {
	case "low", "medium", "high":
		return
	default:
		payload["confidence"] = "low"
	}
}

func allowedString(payload map[string]any, key string, allowed []string) bool {
	value := stringField(payload, key)
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

func numberField(payload map[string]any, key string) float64 {
	value, _ := numericValue(payload[key])
	return value
}

func numericValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func copyStringMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
