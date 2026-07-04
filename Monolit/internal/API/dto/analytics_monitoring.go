package dto

type AnalyticsOverviewResponse struct {
	CallsTotal             int                  `json:"calls_total"`
	CallsNew               int                  `json:"calls_new"`
	CallsProcessing        int                  `json:"calls_processing"`
	CallsTranscribed       int                  `json:"calls_transcribed"`
	CallsAnalyzed          int                  `json:"calls_analyzed"`
	CallsFailed            int                  `json:"calls_failed"`
	AverageDurationSeconds *int                 `json:"average_duration_seconds"`
	AverageQualityScore    *float64             `json:"average_quality_score"`
	QualityScoreScale      int                  `json:"quality_score_scale"`
	TopTopics              []AnalyticsTopicItem `json:"top_topics"`
	RisksCount             *int                 `json:"risks_count"`
	RecommendationsCount   *int                 `json:"recommendations_count"`
	Charts                 AnalyticsCharts      `json:"charts"`
}

type AnalyticsTopicItem struct {
	Title string `json:"title"`
	Count int    `json:"count"`
}

type ProcessingMonitoringResponse struct {
	Queue                    ProcessingQueueResponse `json:"queue"`
	AverageProcessingSeconds *int                    `json:"average_processing_seconds"`
}

type ProcessingQueueResponse struct {
	Pending int `json:"pending"`
	Running int `json:"running"`
	Done    int `json:"done"`
	Failed  int `json:"failed"`
	Retry   int `json:"retry"`
}

type AnalyticsCharts struct {
	CallsByDay    []AnalyticsCountPoint    `json:"calls_by_day"`
	AnalyzedByDay []AnalyticsCountPoint    `json:"analyzed_by_day"`
	QualityByDay  []AnalyticsQualityPoint  `json:"quality_by_day"`
	DurationByDay []AnalyticsDurationPoint `json:"duration_by_day"`
	RisksByDay    []AnalyticsCountPoint    `json:"risks_by_day"`
}

type AnalyticsCountPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type AnalyticsQualityPoint struct {
	Date                string  `json:"date"`
	AverageQualityScore float64 `json:"average_quality_score"`
}

type AnalyticsDurationPoint struct {
	Date                   string `json:"date"`
	AverageDurationSeconds int    `json:"average_duration_seconds"`
}
