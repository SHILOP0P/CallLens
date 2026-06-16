package invitation

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) requireActiveCompanySubscription(ctx context.Context, companyID uuid.UUID) error {
	if s.billingLimiter == nil {
		return nil
	}

	return s.billingLimiter.CanUseCompany(ctx, companyID)
}
