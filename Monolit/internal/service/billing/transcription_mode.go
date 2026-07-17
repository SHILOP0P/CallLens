package billing

import (
	"context"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func (s *Service) ResolveTranscriptionMode(ctx context.Context, userID uuid.UUID, companyID uuid.NullUUID) (models.TranscriptionMode, error) {
	if companyID.Valid {
		if _, err := s.activeBusinessSubscription(ctx, companyID.UUID); err != nil {
			return "", err
		}
		return models.TranscriptionModeDiarized, nil
	}
	subscription, err := s.activePersonalSubscription(ctx, userID)
	if err != nil {
		return "", err
	}
	return transcriptionModeForPlan(subscription.Plan), nil
}

func transcriptionModeForPlan(plan models.Plan) models.TranscriptionMode {
	if plan.Type == models.PlanTypeBusiness {
		return models.TranscriptionModeDiarized
	}
	switch plan.Code {
	case models.PlanCodePersonalPlus, models.PlanCodePersonalPro:
		return models.TranscriptionModeDiarized
	default:
		return models.TranscriptionModeStandard
	}
}
