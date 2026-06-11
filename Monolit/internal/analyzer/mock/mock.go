package mock

import (
	"calllens/monolit/internal/models"
	"context"
	"encoding/json"
)

type Analyzer struct {
	model string
}

func New(model string) *Analyzer {
	return &Analyzer{model: model}
}

func (a *Analyzer) Provider() string {
	return "mock"
}

func (a *Analyzer) Analyze(ctx context.Context, request models.AnalysisRequest) (models.AnalysisResult, error) {
	select {
	case <-ctx.Done():
		return models.AnalysisResult{}, ctx.Err()
	default:
	}

	payload := map[string]any{
		"summary":            "Mock call analysis",
		"call_uuid":          request.CallUUID.String(),
		"transcription_size": len(request.Transcription),
		"instruction_count":  len(request.Instructions),
	}

	resultJSON, err := json.Marshal(payload)
	if err != nil {
		return models.AnalysisResult{}, err
	}

	resultText := "Mock call analysis: transcription and instructions were accepted."
	model := stringPtr(a.model)
	if a.model == "" {
		model = nil
	}

	return models.AnalysisResult{
		ResultJSON: resultJSON,
		ResultText: &resultText,
		Model:      model,
	}, nil
}

func stringPtr(value string) *string {
	return &value
}
