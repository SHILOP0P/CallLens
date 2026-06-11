package analyzer

import (
	"calllens/monolit/internal/models"
	"context"
)

type Analyzer interface {
	Provider() string
	Analyze(ctx context.Context, request models.AnalysisRequest) (models.AnalysisResult, error)
}
