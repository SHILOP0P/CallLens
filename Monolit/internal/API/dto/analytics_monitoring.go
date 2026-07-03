package dto

import "time"

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
	ConversionToDeal       *float64             `json:"conversion_to_deal"`
	ConversionReason       string               `json:"conversion_reason"`
}

type AnalyticsTopicItem struct {
	Title string `json:"title"`
	Count int    `json:"count"`
}

type ProcessingMonitoringResponse struct {
	Queue                    ProcessingQueueResponse       `json:"queue"`
	AverageProcessingSeconds *int                          `json:"average_processing_seconds"`
	LastFailedJobs           []FailedProcessingJobResponse `json:"last_failed_jobs"`
	Services                 ProcessingServicesResponse    `json:"services"`
}

type ProcessingQueueResponse struct {
	Pending int `json:"pending"`
	Running int `json:"running"`
	Done    int `json:"done"`
	Failed  int `json:"failed"`
	Retry   int `json:"retry"`
}

type FailedProcessingJobResponse struct {
	JobUUID    string    `json:"job_uuid"`
	Type       string    `json:"type"`
	EntityUUID string    `json:"entity_uuid"`
	Attempts   int       `json:"attempts"`
	LastError  *string   `json:"last_error"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ProcessingServicesResponse struct {
	Transcriber string `json:"transcriber"`
	Analyzer    string `json:"analyzer"`
	Storage     string `json:"storage"`
}
