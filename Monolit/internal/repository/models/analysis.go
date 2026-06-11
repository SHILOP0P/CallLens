package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CallAnalysis struct {
	ID           uuid.UUID
	CallUUID     uuid.UUID
	Status       CallAnalysisStatus
	Provider     string
	Model        sql.NullString
	ResultJSON   json.RawMessage
	ResultText   sql.NullString
	ErrorMessage sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CallAnalysisStatus string

const (
	CallAnalysisStatusPending    CallAnalysisStatus = "pending"
	CallAnalysisStatusProcessing CallAnalysisStatus = "processing"
	CallAnalysisStatusDone       CallAnalysisStatus = "done"
	CallAnalysisStatusFailed     CallAnalysisStatus = "failed"
)
