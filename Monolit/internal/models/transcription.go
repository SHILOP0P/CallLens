package models

import (
	"time"

	"github.com/google/uuid"
)

type Transcription struct {
	ID           uuid.UUID
	CallUUID     uuid.UUID
	Status       TranscriptionStatus
	Text         *string
	Language     *string
	Provider     string
	ErrorMessage *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TranscriptionStatus string

const (
	TranscriptionStatusProcessing  TranscriptionStatus = "processing"
	TranscriptionStatusTranscribed TranscriptionStatus = "transcribed"
	TranscriptionStatusFailed      TranscriptionStatus = "failed"
)
