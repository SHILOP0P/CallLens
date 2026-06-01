package models

import (
	"time"

	"github.com/google/uuid"
)

type Call struct {
	ID               uuid.UUID
	Title            string
	Status           string
	AudioPath        string
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
	DurationSeconds  int
	CreatedAt        time.Time
}

const (
	CallStatusNew        string = "new"
	CallStatusProcessing string = "processing"
	CallStatusDone       string = "done"
	CallStatusFailed     string = "failed"
)
