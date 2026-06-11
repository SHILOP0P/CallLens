package models

import (
	"time"

	"github.com/google/uuid"
)

type AnalysisInstruction struct {
	ID                uuid.UUID
	Scope             string
	UserUUID          uuid.NullUUID
	CompanyUUID       uuid.NullUUID
	DepartmentUUID    uuid.NullUUID
	Title             string
	OriginalFilename  string
	FilePath          string
	MimeType          string
	SizeBytes         int64
	ContentSHA256     string
	SortOrder         int
	IsActive          bool
	CreatedByUserUUID uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
