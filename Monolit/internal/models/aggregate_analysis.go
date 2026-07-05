package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const DeepAnalysisWeeklyLimit = 2

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
	Scope            AggregateAnalysisScope
	PeriodFrom       time.Time
	PeriodTo         time.Time
	SourceCallsCount int
	Sources          []AggregateAnalysisSourceCall
	Metrics          AggregateAnalysisSourceMetrics
}

type AggregateAnalysisSourceMetrics struct {
	IncludedCalls int `json:"included_calls"`
	TotalCalls    int `json:"total_calls"`
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
