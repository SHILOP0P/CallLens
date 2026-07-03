package models

import (
	"time"

	"github.com/google/uuid"
)

type SearchType string

const (
	SearchTypeCalls        SearchType = "calls"
	SearchTypeCompanies    SearchType = "companies"
	SearchTypeReports      SearchType = "reports"
	SearchTypeInstructions SearchType = "instructions"
)

type SearchInput struct {
	UserUUID uuid.UUID
	Query    string
	Types    []SearchType
	Limit    int
}

type SearchResult struct {
	Calls        []SearchCallResult
	Companies    []SearchCompanyResult
	Reports      []SearchReportResult
	Instructions []SearchInstructionResult
}

type SearchCallResult struct {
	ID        uuid.UUID
	Title     string
	Status    CallStatus
	CreatedAt time.Time
}

type SearchCompanyResult struct {
	ID   uuid.UUID
	Name string
}

type SearchReportResult struct {
	ID       uuid.UUID
	CallUUID uuid.UUID
	FileName string
	Status   ReportStatus
}

type SearchInstructionResult struct {
	ID    uuid.UUID
	Title string
	Scope AnalysisInstructionScope
}
