package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DeepAnalysisWeeklyLimit is deliberately elevated while deep analysis is being tested.
const DeepAnalysisWeeklyLimit = 100

type AggregateAnalysisScope string

const (
	AggregateAnalysisScopePersonal   AggregateAnalysisScope = "personal"
	AggregateAnalysisScopeCompany    AggregateAnalysisScope = "company"
	AggregateAnalysisScopeDepartment AggregateAnalysisScope = "department"
	AggregateAnalysisScopeFolder     AggregateAnalysisScope = "folder"
)

type AggregateAnalysisStatus string

const (
	AggregateAnalysisStatusPending    AggregateAnalysisStatus = "pending"
	AggregateAnalysisStatusProcessing AggregateAnalysisStatus = "processing"
	AggregateAnalysisStatusDone       AggregateAnalysisStatus = "done"
	AggregateAnalysisStatusFailed     AggregateAnalysisStatus = "failed"
)

type DeepAnalysisSubjectType string

const (
	DeepAnalysisSubjectTypeUser    DeepAnalysisSubjectType = "user"
	DeepAnalysisSubjectTypeCompany DeepAnalysisSubjectType = "company"
)

type AggregateAnalysis struct {
	ID                uuid.UUID
	Scope             AggregateAnalysisScope
	UserUUID          uuid.NullUUID
	CompanyUUID       uuid.NullUUID
	DepartmentUUID    uuid.NullUUID
	FolderUUID        uuid.NullUUID
	PeriodFrom        time.Time
	PeriodTo          time.Time
	Status            AggregateAnalysisStatus
	Provider          string
	Model             *string
	SourceCallsCount  int
	ResultJSON        json.RawMessage
	ResultText        *string
	ErrorMessage      *string
	CreatedByUserUUID uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CreateDeepAnalysisInput struct {
	UserID         uuid.UUID
	Scope          AggregateAnalysisScope
	CompanyUUID    uuid.NullUUID
	DepartmentUUID uuid.NullUUID
	FolderUUID     uuid.NullUUID
	PeriodFrom     time.Time
	PeriodTo       time.Time
	Force          bool
}

type ListDeepAnalysesInput struct {
	UserID         uuid.UUID
	Scope          AggregateAnalysisScope
	CompanyUUID    uuid.NullUUID
	DepartmentUUID uuid.NullUUID
	FolderUUID     uuid.NullUUID
	From           *time.Time
	To             *time.Time
	Status         AggregateAnalysisStatus
	Limit          int
	Offset         int
}

type ListAggregateAnalysesResult struct {
	Items  []AggregateAnalysis
	Total  int
	Limit  int
	Offset int
}

type AggregateAnalysisRequest struct {
	Scope            AggregateAnalysisScope         `json:"scope"`
	PeriodFrom       time.Time                      `json:"period_from"`
	PeriodTo         time.Time                      `json:"period_to"`
	SourceCallsCount int                            `json:"source_calls_count"`
	Sources          []AggregateAnalysisSourceCall  `json:"representative_calls"`
	Metrics          AggregateAnalysisSourceMetrics `json:"metrics"`
	Dataset          AggregateAnalysisSourceDataset `json:"dataset"`
}

type AggregateAnalysisSourceMetrics struct {
	IncludedCalls       int    `json:"included_calls"`
	TotalCalls          int    `json:"total_calls"`
	AggregatedCalls     int    `json:"aggregated_calls"`
	RepresentativeCalls int    `json:"representative_calls"`
	SourceSetHash       string `json:"source_set_hash"`
}

type AggregateAnalysisSourceDataset struct {
	SourceSummary      AggregateAnalysisSourceSummary     `json:"source_summary"`
	ScoreSummary       AggregateAnalysisScoreSummary      `json:"score_summary"`
	IssueCoverage      []AggregateAnalysisFrequency       `json:"issue_coverage"`
	WeakCriteria       []AggregateAnalysisCriterionMetric `json:"weak_criteria"`
	BusinessOutcomes   []AggregateAnalysisFrequency       `json:"business_outcomes"`
	LostReasons        []AggregateAnalysisFrequency       `json:"lost_reasons"`
	CustomerObjections []AggregateAnalysisFrequency       `json:"customer_objections"`
	Risks              []AggregateAnalysisFrequency       `json:"risks"`
	Topics             []AggregateAnalysisFrequency       `json:"topics"`
	NextStepSummary    AggregateAnalysisNextStepSummary   `json:"next_step_summary"`
	AttentionCalls     []AggregateAnalysisCallEvidence    `json:"attention_calls"`
	StrongCalls        []AggregateAnalysisCallEvidence    `json:"strong_calls"`
}

type AggregateAnalysisSourceSummary struct {
	AnalyzedCalls        int    `json:"analyzed_calls"`
	IncludedInStatistics int    `json:"included_in_statistics"`
	RepresentativeCalls  int    `json:"representative_calls"`
	AllAnalyzedCallsUsed bool   `json:"all_analyzed_calls_used"`
	SourceSetHash        string `json:"source_set_hash"`
}

type AggregateAnalysisScoreSummary struct {
	CallsWithScore int      `json:"calls_with_score"`
	Average        *float64 `json:"average,omitempty"`
	Min            *float64 `json:"min,omitempty"`
	Max            *float64 `json:"max,omitempty"`
	LowCount       int      `json:"low_count"`
	MediumCount    int      `json:"medium_count"`
	HighCount      int      `json:"high_count"`
}

type AggregateAnalysisFrequency struct {
	Code            string   `json:"code"`
	Title           string   `json:"title"`
	Count           int      `json:"count"`
	Share           float64  `json:"share"`
	SampleCallUUIDs []string `json:"sample_call_uuids"`
}

type AggregateAnalysisCriterionMetric struct {
	Code               string   `json:"code"`
	Title              string   `json:"title"`
	ApplicableCalls    int      `json:"applicable_calls"`
	WeakCalls          int      `json:"weak_calls"`
	WeakShare          float64  `json:"weak_share"`
	AveragePointsShare *float64 `json:"average_points_share,omitempty"`
	MissedCalls        int      `json:"missed_calls"`
	PartiallyMetCalls  int      `json:"partially_met_calls"`
	UnclearCalls       int      `json:"unclear_calls"`
	SampleCallUUIDs    []string `json:"sample_call_uuids"`
}

type AggregateAnalysisNextStepSummary struct {
	CallsWithNextStep         int     `json:"calls_with_next_step"`
	CallsWithSpecificNextStep int     `json:"calls_with_specific_next_step"`
	CallsMissingNextStep      int     `json:"calls_missing_next_step"`
	CallsMissingSpecificStep  int     `json:"calls_missing_specific_step"`
	MissingNextStepShare      float64 `json:"missing_next_step_share"`
	MissingSpecificStepShare  float64 `json:"missing_specific_step_share"`
}

type AggregateAnalysisCallEvidence struct {
	CallUUID   uuid.UUID `json:"call_uuid"`
	CreatedAt  time.Time `json:"created_at"`
	Title      string    `json:"title"`
	Score      *float64  `json:"score,omitempty"`
	Summary    string    `json:"summary,omitempty"`
	IssueCodes []string  `json:"issue_codes,omitempty"`
}

type AggregateAnalysisSourceCall struct {
	CallUUID           uuid.UUID `json:"call_uuid"`
	CreatedAt          time.Time `json:"created_at"`
	Title              string    `json:"title"`
	Score              *float64  `json:"score,omitempty"`
	Summary            string    `json:"summary,omitempty"`
	Topics             any       `json:"topics,omitempty"`
	CriteriaResults    any       `json:"criteria_results,omitempty"`
	BusinessOutcome    any       `json:"business_outcome,omitempty"`
	CustomerSignals    any       `json:"customer_signals,omitempty"`
	IssueCodes         any       `json:"issue_codes,omitempty"`
	Risks              any       `json:"risks,omitempty"`
	CustomerObjections any       `json:"customer_objections,omitempty"`
	NextStepQuality    any       `json:"next_step_quality,omitempty"`
}
