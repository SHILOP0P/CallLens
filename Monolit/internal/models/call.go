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
	CreatedAt          time.Time
}

type CallStatus string

const (
	CallStatusNew        CallStatus = "new"
	CallStatusProcessing CallStatus = "processing"
	CallStatusDone       CallStatus = "done"
	CallStatusFailed     CallStatus = "failed"
)

type CreateCallInput struct {
	Title              string
	OriginalFilename   string
	MimeType           string
	SizeBytes          int64
	Content            io.Reader
	UploadedByUserUUID uuid.UUID
}
