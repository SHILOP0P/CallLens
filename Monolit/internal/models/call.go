package models

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type Call struct {
	ID                 uuid.UUID
	Title              string
	Status             CallStatus
	AudioPath          string
	OriginalFilename   string
	MimeType           string
	SizeBytes          int64
	DurationSeconds    int
	UploadedByUserUUID uuid.NullUUID
	CompanyUUID        uuid.NullUUID
	DepartmentUUID     uuid.NullUUID
	VisibilityScope    CallVisibilityScope
	CreatedAt          time.Time
}

type CallStatus string
type CallVisibilityScope string

const (
	CallStatusNew         CallStatus = "new"
	CallStatusProcessing  CallStatus = "processing"
	CallStatusTranscribed CallStatus = "transcribed"
	CallStatusAnalyzed    CallStatus = "analyzed"
	CallStatusFailed      CallStatus = "failed"
)

const (
	CallVisibilityScopePersonal   CallVisibilityScope = "personal"
	CallVisibilityScopeCompany    CallVisibilityScope = "company"
	CallVisibilityScopeDepartment CallVisibilityScope = "department"
)

type CreateCallInput struct {
	Title              string
	OriginalFilename   string
	MimeType           string
	SizeBytes          int64
	Content            io.Reader
	UploadedByUserUUID uuid.UUID
	CompanyUUID        uuid.NullUUID
	DepartmentUUID     uuid.NullUUID
	VisibilityScope    CallVisibilityScope
}

type UpdateCallStatusInput struct {
	CallUUID uuid.UUID
	Status   CallStatus
}
