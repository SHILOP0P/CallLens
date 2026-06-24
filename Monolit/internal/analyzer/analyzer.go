package analyzer

import (
	"context"

	"calllens/monolit/internal/models"
)

type Analyzer interface {
	Provider() string
	Analyze(ctx context.Context, request models.AnalysisRequest) (models.AnalysisResult, error)
}
