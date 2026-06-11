package models

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type AnalysisInstructionScope string

const (
	AnalysisInstructionScopePersonal   AnalysisInstructionScope = "personal"
	AnalysisInstructionScopeCompany    AnalysisInstructionScope = "company"
	AnalysisInstructionScopeDepartment AnalysisInstructionScope = "department"
)

const (
	CompanyInstructionLimit         = 10
	DepartmentInstructionLimit      = 10
	DefaultPersonalInstructionLimit = 5
)

type AnalysisInstruction struct {
	ID                uuid.UUID
	Scope             AnalysisInstructionScope
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

type CreateAnalysisInstructionInput struct {
	Scope             AnalysisInstructionScope
	UserUUID          uuid.UUID
	CompanyUUID       uuid.NullUUID
	DepartmentUUID    uuid.NullUUID
	Title             string
	OriginalFilename  string
	MimeType          string
	SizeBytes         int64
	Content           io.Reader
	CreatedByUserUUID uuid.UUID
}

type ListAnalysisInstructionsInput struct {
	Scope          AnalysisInstructionScope
	UserUUID       uuid.UUID
	CompanyUUID    uuid.NullUUID
	DepartmentUUID uuid.NullUUID
}

type SaveInstructionInput struct {
	InstructionUUID  uuid.UUID
	Scope            AnalysisInstructionScope
	UserUUID         uuid.NullUUID
	CompanyUUID      uuid.NullUUID
	DepartmentUUID   uuid.NullUUID
	OriginalFilename string
	Content          io.Reader
	MimeType         string
}

type SavedInstructionFile struct {
	Path          string
	MimeType      string
	SizeBytes     int64
	ContentSHA256 string
}
