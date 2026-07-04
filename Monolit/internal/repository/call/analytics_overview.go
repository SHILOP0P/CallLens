package call

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

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

	if err := r.fillAnalysisAggregates(ctx, &overview, where, args); err != nil {
		return model.AnalyticsOverview{}, err
	}

	return overview, nil
}

type analyticsCallRow struct {
	Status          model.CallStatus
	DurationSeconds int
	CreatedAt       time.Time
	ResultJSON      sql.NullString
}

type analyticsAccumulator struct {
	callsByDay    map[string]int
	analyzedByDay map[string]int
	durationByDay map[string][]int
	qualityByDay  map[string][]float64
	risksByDay    map[string]int
	topicCounts   map[string]int

	qualityScores []float64
	risksCount    int
	recsCount     int
	analysisSeen  bool
}

func (r *Repository) fillAnalysisAggregates(ctx context.Context, overview *model.AnalyticsOverview, where string, args []any) error {
	query := fmt.Sprintf(`
	SELECT c.status,
	       c.duration_seconds,
	       c.created_at,
	       ca.result_json::text
	FROM calls c
	LEFT JOIN call_analyses ca
	  ON ca.call_uuid = c.call_uuid
	 AND ca.status = 'done'
	WHERE %s
	`, where)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("get analytics details: %w", err)
	}
	defer func() { _ = rows.Close() }()

	acc := analyticsAccumulator{
		callsByDay:    map[string]int{},
		analyzedByDay: map[string]int{},
		durationByDay: map[string][]int{},
		qualityByDay:  map[string][]float64{},
		risksByDay:    map[string]int{},
		topicCounts:   map[string]int{},
	}

	for rows.Next() {
		var row analyticsCallRow
		if err := rows.Scan(&row.Status, &row.DurationSeconds, &row.CreatedAt, &row.ResultJSON); err != nil {
			return fmt.Errorf("scan analytics details: %w", err)
		}
		acc.addCall(row)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan analytics details: %w", err)
	}

	acc.apply(overview)
	return nil
}

func (a *analyticsAccumulator) addCall(row analyticsCallRow) {
	day := row.CreatedAt.UTC().Format("2006-01-02")
	a.callsByDay[day]++
	if row.Status == model.CallStatusAnalyzed {
		a.analyzedByDay[day]++
	}
	if row.DurationSeconds > 0 {
		a.durationByDay[day] = append(a.durationByDay[day], row.DurationSeconds)
	}
	if !row.ResultJSON.Valid || strings.TrimSpace(row.ResultJSON.String) == "" {
		return
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(row.ResultJSON.String), &payload); err != nil {
		return
	}
	a.analysisSeen = true

	if score, ok := extractQualityScore(payload); ok {
		a.qualityScores = append(a.qualityScores, score)
		a.qualityByDay[day] = append(a.qualityByDay[day], score)
	}

	risks := countListValues(payload["risks"]) +
		countListValues(payload["customer_objections"]) +
		countNestedListValues(payload, "manager_quality", "issues")
	a.risksCount += risks
	a.risksByDay[day] += risks

	a.recsCount += countNestedListValues(payload, "manager_quality", "recommendations") +
		countListValues(payload["next_steps"]) +
		countListValues(payload["recommendations"])

	addTopics(a.topicCounts, payload["topics"])
	addTopics(a.topicCounts, payload["top_topics"])
}

func (a *analyticsAccumulator) apply(overview *model.AnalyticsOverview) {
	overview.Charts = model.AnalyticsCharts{
		CallsByDay:    countMapToPoints(a.callsByDay),
		AnalyzedByDay: countMapToPoints(a.analyzedByDay),
		QualityByDay:  averageFloatMapToQualityPoints(a.qualityByDay),
		DurationByDay: averageIntMapToDurationPoints(a.durationByDay),
		RisksByDay:    countMapToPoints(a.risksByDay),
	}

	if len(a.qualityScores) > 0 {
		average := roundFloat(averageFloat(a.qualityScores), 1)
		overview.AverageQualityScore = &average
	}
	if a.analysisSeen {
		risks := a.risksCount
		recs := a.recsCount
		overview.RisksCount = &risks
		overview.RecommendationsCount = &recs
	}
	overview.TopTopics = topicMapToCounts(a.topicCounts, 10)
}

func extractQualityScore(payload map[string]any) (float64, bool) {
	for _, key := range []string{"quality_score", "score", "overall_score", "manager_score"} {
		score, ok := numberValue(payload[key])
		if !ok || score <= 0 {
			continue
		}
		if score > 5 {
			score = score / 20
		}
		if score < 1 {
			score = 1
		}
		if score > 5 {
			score = 5
		}
		return roundFloat(score, 1), true
	}
	return 0, false
}

func numberValue(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case json.Number:
		n, err := v.Float64()
		return n, err == nil
	default:
		return 0, false
	}
}

func countNestedListValues(payload map[string]any, objectKey string, listKey string) int {
	object, ok := payload[objectKey].(map[string]any)
	if !ok {
		return 0
	}
	return countListValues(object[listKey])
}

func countListValues(value any) int {
	switch v := value.(type) {
	case []any:
		return len(v)
	case []string:
		return len(v)
	default:
		return 0
	}
}

func addTopics(counts map[string]int, value any) {
	switch topics := value.(type) {
	case []any:
		for _, item := range topics {
			switch topic := item.(type) {
			case string:
				addTopic(counts, topic)
			case map[string]any:
				if title, ok := topic["title"].(string); ok {
					addTopic(counts, title)
				}
			}
		}
	case []string:
		for _, topic := range topics {
			addTopic(counts, topic)
		}
	}
}

func addTopic(counts map[string]int, topic string) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return
	}
	counts[topic]++
}

func countMapToPoints(values map[string]int) []model.AnalyticsCountPoint {
	dates := sortedKeys(values)
	points := make([]model.AnalyticsCountPoint, 0, len(dates))
	for _, date := range dates {
		points = append(points, model.AnalyticsCountPoint{Date: date, Count: values[date]})
	}
	return points
}

func averageFloatMapToQualityPoints(values map[string][]float64) []model.AnalyticsQualityPoint {
	dates := sortedKeys(values)
	points := make([]model.AnalyticsQualityPoint, 0, len(dates))
	for _, date := range dates {
		points = append(points, model.AnalyticsQualityPoint{
			Date:                date,
			AverageQualityScore: roundFloat(averageFloat(values[date]), 1),
		})
	}
	return points
}

func averageIntMapToDurationPoints(values map[string][]int) []model.AnalyticsDurationPoint {
	dates := sortedKeys(values)
	points := make([]model.AnalyticsDurationPoint, 0, len(dates))
	for _, date := range dates {
		points = append(points, model.AnalyticsDurationPoint{
			Date:                   date,
			AverageDurationSeconds: averageInt(values[date]),
		})
	}
	return points
}

func topicMapToCounts(values map[string]int, limit int) []model.AnalyticsTopicCount {
	topics := make([]model.AnalyticsTopicCount, 0, len(values))
	for title, count := range values {
		topics = append(topics, model.AnalyticsTopicCount{Title: title, Count: count})
	}
	sort.Slice(topics, func(i, j int) bool {
		if topics[i].Count == topics[j].Count {
			return topics[i].Title < topics[j].Title
		}
		return topics[i].Count > topics[j].Count
	})
	if limit > 0 && len(topics) > limit {
		return topics[:limit]
	}
	return topics
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func averageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func averageInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	var sum int
	for _, value := range values {
		sum += value
	}
	return int(math.Round(float64(sum) / float64(len(values))))
}

func roundFloat(value float64, precision int) float64 {
	scale := math.Pow(10, float64(precision))
	return math.Round(value*scale) / scale
}
