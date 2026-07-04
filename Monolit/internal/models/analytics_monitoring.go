package models

import (
	"time"

	"github.com/google/uuid"
)

type AnalyticsOverviewInput struct {
	UserID          uuid.UUID
	VisibilityScope CallVisibilityScope
	CompanyUUID     uuid.NullUUID
	DepartmentUUID  uuid.NullUUID
	From            *time.Time
	To              *time.Time
}

type AnalyticsTopicCount struct {
	Title string
	Count int
}

type AnalyticsOverview struct {
	CallsTotal             int
	CallsNew               int
	CallsProcessing        int
	CallsTranscribed       int
	CallsAnalyzed          int
	CallsFailed            int
	AverageDurationSeconds *int
	AverageQualityScore    *float64
	QualityScoreScale      int
	TopTopics              []AnalyticsTopicCount
	RisksCount             *int
	RecommendationsCount   *int
	Charts                 AnalyticsCharts
}

type AnalyticsCharts struct {
	CallsByDay    []AnalyticsCountPoint
	AnalyzedByDay []AnalyticsCountPoint
	QualityByDay  []AnalyticsQualityPoint
	DurationByDay []AnalyticsDurationPoint
	RisksByDay    []AnalyticsCountPoint
}

type AnalyticsCountPoint struct {
	Date  string
	Count int
}

type AnalyticsQualityPoint struct {
	Date                string
	AverageQualityScore float64
}

type AnalyticsDurationPoint struct {
	Date                   string
	AverageDurationSeconds int
}

type ProcessingMonitoringInput struct {
	UserID      uuid.UUID
	UserRole    UserRole
	CompanyUUID uuid.NullUUID
	From        *time.Time
	To          *time.Time
}

type ProcessingQueueSummary struct {
	Pending int
	Running int
	Done    int
	Failed  int
	Retry   int
}

type FailedProcessingJob struct {
	ID         uuid.UUID
	Type       ProcessingJobType
	EntityUUID uuid.UUID
	Attempts   int
	LastError  *string
	UpdatedAt  time.Time
}

type ProcessingServicesStatus struct {
	Transcriber string
	Analyzer    string
	Storage     string
}

type ProcessingMonitoring struct {
	Queue                    ProcessingQueueSummary
	AverageProcessingSeconds *int
	LastFailedJobs           []FailedProcessingJob
	Services                 ProcessingServicesStatus
}
