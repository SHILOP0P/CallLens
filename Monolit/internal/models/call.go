package models

import (
	"time"

	"github.com/google/uuid"
)

type Call struct {
	ID               uuid.UUID
	Title            string
	Status           CallStatus
	AudioPath        string
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
	DurationSeconds  int
	CreatedAt        time.Time
}

type CallStatus string

const (
	CallStatusNew        CallStatus = "new"
	CallStatusProcessing CallStatus = "processing"
	CallStatusDone       CallStatus = "done"
	CallStatusFailed     CallStatus = "failed"
)
