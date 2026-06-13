package models

type TranscriptionResult struct {
	Text     string
	Segments []TranscriptionSegment
	Language *string
}
