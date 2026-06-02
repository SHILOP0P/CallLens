package models

import (
	"io"

	"github.com/google/uuid"
)

type SaveInput struct {
	CallID           uuid.UUID
	OriginalFilename string
	Content          io.Reader
	SizeBytes        int64
	MimeType         string
}

type SavedFile struct {
	Path             string
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
}

type File struct {
	Content          io.ReadCloser
	Path             string
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
}
