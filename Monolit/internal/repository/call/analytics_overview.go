package call

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	model "calllens/monolit/internal/models"
)

func (r *Repository) GetAnalyticsOverview(ctx context.Context, input model.AnalyticsOverviewInput) (model.AnalyticsOverview, error) {
	where, args := buildListFilters(model.ListCallsInput{
		UserID:          input.UserID,
		VisibilityScope: input.VisibilityScope,
		CompanyUUID:     input.CompanyUUID,
		DepartmentUUID:  input.DepartmentUUID,
		From:            input.From,
		To:              input.To,
	})

	query := fmt.Sprintf(`
	SELECT COUNT(*)::int,
	       COUNT(*) FILTER (WHERE c.status = 'new')::int,
	       COUNT(*) FILTER (WHERE c.status = 'processing')::int,
	       COUNT(*) FILTER (WHERE c.status = 'transcribed')::int,
	       COUNT(*) FILTER (WHERE c.status = 'analyzed')::int,
	       COUNT(*) FILTER (WHERE c.status = 'failed')::int,
	       AVG(c.duration_seconds)::float8
	FROM calls c
	WHERE %s
	`, where)

	var overview model.AnalyticsOverview
	var averageDuration sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&overview.CallsTotal,
		&overview.CallsNew,
		&overview.CallsProcessing,
		&overview.CallsTranscribed,
		&overview.CallsAnalyzed,
		&overview.CallsFailed,
		&averageDuration,
	)
	if err != nil {
		return model.AnalyticsOverview{}, fmt.Errorf("get analytics overview: %w", err)
	}

	if averageDuration.Valid {
		rounded := int(math.Round(averageDuration.Float64))
		overview.AverageDurationSeconds = &rounded
	}
	overview.QualityScoreScale = 5
	overview.TopTopics = []model.AnalyticsTopicCount{}
	overview.ConversionReason = "deal data is not tracked"

	return overview, nil
}
