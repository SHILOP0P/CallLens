package models

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type ReportFormat string
type ReportStatus string

const (
	ReportFormatPDF  ReportFormat = "pdf"
	ReportFormatDOCX ReportFormat = "docx"
	ReportFormatMD   ReportFormat = "md"
	ReportFormatXLSX ReportFormat = "xlsx"
)

const (
	ReportStatusPending ReportStatus = "pending"
	ReportStatusReady   ReportStatus = "ready"
	ReportStatusFailed  ReportStatus = "failed"
)

type ReportExport struct {
	ID                  uuid.UUID
	CallUUID            uuid.UUID
	AnalysisUUID        uuid.UUID
	RequestedByUserUUID uuid.UUID
	Format              ReportFormat
	Status              ReportStatus
	StoragePath         *string
	FileName            string
	ContentType         string
	SizeBytes           int64
	ErrorMessage        *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ExpiresAt           time.Time
}

type CreateReportInput struct {
	CallUUID uuid.UUID
	UserUUID uuid.UUID
	Format   ReportFormat
}

type CreateReportExportInput struct {
	Report ReportExport
}

type MarkReportReadyInput struct {
	ID          uuid.UUID
	StoragePath string
	FileName    string
	ContentType string
	SizeBytes   int64
}

type MarkReportFailedInput struct {
	ID           uuid.UUID
	ErrorMessage string
}

type SaveReportInput struct {
	ReportUUID uuid.UUID
	CallUUID   uuid.UUID
	Format     ReportFormat
	FileName   string
	MimeType   string
	Content    io.Reader
}

type SavedReportFile struct {
	Path      string
	MimeType  string
	SizeBytes int64
}

type ReportFile struct {
	Report  ReportExport
	Content io.ReadCloser
}
