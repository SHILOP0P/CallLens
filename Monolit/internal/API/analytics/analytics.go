package analytics

import (
	"net/http"
	"time"

	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/models"
	"calllens/monolit/internal/service"

	"github.com/google/uuid"
)

type Handler struct {
	service service.AnalyticsService
}

func NewHandler(service service.AnalyticsService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	input, err := parseOverviewInput(r, userID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeInvalidCallFilter, "invalid analytics filter")
		return
	}

	overview, err := h.service.GetOverview(r.Context(), input)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToGetAnalyticsOverview, "failed to get analytics overview")
		return
	}

	if err := response.WriteJSON(w, http.StatusOK, overviewToAPI(overview)); err != nil {
		return
	}
}

func parseOverviewInput(r *http.Request, userID uuid.UUID) (models.AnalyticsOverviewInput, error) {
	query := r.URL.Query()
	input := models.AnalyticsOverviewInput{UserID: userID}

	if scope := query.Get("scope"); scope != "" {
		parsed := models.CallVisibilityScope(scope)
		if !isValidCallVisibilityScope(parsed) {
			return models.AnalyticsOverviewInput{}, models.ErrInvalidCallFilter
		}
		input.VisibilityScope = parsed
	}

	var err error
	if input.CompanyUUID, err = parseOptionalUUID(query.Get("company_uuid")); err != nil {
		return models.AnalyticsOverviewInput{}, err
	}
	if input.DepartmentUUID, err = parseOptionalUUID(query.Get("department_uuid")); err != nil {
		return models.AnalyticsOverviewInput{}, err
	}
	if input.From, err = parseOptionalISOTime(query.Get("from"), false); err != nil {
		return models.AnalyticsOverviewInput{}, err
	}
	if input.To, err = parseOptionalISOTime(query.Get("to"), true); err != nil {
		return models.AnalyticsOverviewInput{}, err
	}
	if input.From != nil && input.To != nil && input.From.After(*input.To) {
		return models.AnalyticsOverviewInput{}, models.ErrInvalidCallFilter
	}

	return input, nil
}

func overviewToAPI(overview models.AnalyticsOverview) dto.AnalyticsOverviewResponse {
	topics := make([]dto.AnalyticsTopicItem, len(overview.TopTopics))
	for i, topic := range overview.TopTopics {
		topics[i] = dto.AnalyticsTopicItem{Title: topic.Title, Count: topic.Count}
	}

	return dto.AnalyticsOverviewResponse{
		CallsTotal:             overview.CallsTotal,
		CallsNew:               overview.CallsNew,
		CallsProcessing:        overview.CallsProcessing,
		CallsTranscribed:       overview.CallsTranscribed,
		CallsAnalyzed:          overview.CallsAnalyzed,
		CallsFailed:            overview.CallsFailed,
		AverageDurationSeconds: overview.AverageDurationSeconds,
		AverageQualityScore:    overview.AverageQualityScore,
		QualityScoreScale:      overview.QualityScoreScale,
		TopTopics:              topics,
		RisksCount:             overview.RisksCount,
		RecommendationsCount:   overview.RecommendationsCount,
		Charts:                 analyticsChartsToAPI(overview.Charts),
	}
}

func analyticsChartsToAPI(charts models.AnalyticsCharts) dto.AnalyticsCharts {
	return dto.AnalyticsCharts{
		CallsByDay:    countPointsToAPI(charts.CallsByDay),
		AnalyzedByDay: countPointsToAPI(charts.AnalyzedByDay),
		QualityByDay:  qualityPointsToAPI(charts.QualityByDay),
		DurationByDay: durationPointsToAPI(charts.DurationByDay),
		RisksByDay:    countPointsToAPI(charts.RisksByDay),
	}
}

func countPointsToAPI(points []models.AnalyticsCountPoint) []dto.AnalyticsCountPoint {
	resp := make([]dto.AnalyticsCountPoint, len(points))
	for i, point := range points {
		resp[i] = dto.AnalyticsCountPoint{Date: point.Date, Count: point.Count}
	}
	return resp
}

func qualityPointsToAPI(points []models.AnalyticsQualityPoint) []dto.AnalyticsQualityPoint {
	resp := make([]dto.AnalyticsQualityPoint, len(points))
	for i, point := range points {
		resp[i] = dto.AnalyticsQualityPoint{Date: point.Date, AverageQualityScore: point.AverageQualityScore}
	}
	return resp
}

func durationPointsToAPI(points []models.AnalyticsDurationPoint) []dto.AnalyticsDurationPoint {
	resp := make([]dto.AnalyticsDurationPoint, len(points))
	for i, point := range points {
		resp[i] = dto.AnalyticsDurationPoint{Date: point.Date, AverageDurationSeconds: point.AverageDurationSeconds}
	}
	return resp
}

func parseOptionalUUID(value string) (uuid.NullUUID, error) {
	if value == "" {
		return uuid.NullUUID{}, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.NullUUID{}, models.ErrInvalidCallFilter
	}

	return uuid.NullUUID{UUID: parsed, Valid: true}, nil
}

func parseOptionalISOTime(value string, endOfDate bool) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		utc := parsed.UTC()
		return &utc, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, models.ErrInvalidCallFilter
	}
	parsed = parsed.UTC()
	if endOfDate {
		parsed = parsed.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}

	return &parsed, nil
}

func isValidCallVisibilityScope(scope models.CallVisibilityScope) bool {
	switch scope {
	case models.CallVisibilityScopePersonal,
		models.CallVisibilityScopeCompany,
		models.CallVisibilityScopeDepartment:
		return true
	default:
		return false
	}
}
