package processing_job

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const processingJobReturningColumns = `
	job_uuid,
	job_type,
	entity_uuid,
	status,
	attempts,
	max_attempts,
	available_at,
	locked_at,
	locked_by,
	last_error,
	created_at,
	updated_at
`
