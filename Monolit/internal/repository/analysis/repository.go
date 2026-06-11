package analysis

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const analysisReturningColumns = `
	analysis_uuid,
	call_uuid,
	status,
	provider,
	model,
	result_json,
	result_text,
	error_message,
	created_at,
	updated_at
`
