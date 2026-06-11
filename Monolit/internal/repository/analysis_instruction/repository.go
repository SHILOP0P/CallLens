package analysis_instruction

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const analysisInstructionReturningColumns = `
	instruction_uuid,
	scope,
	user_uuid,
	company_uuid,
	department_uuid,
	title,
	original_filename,
	file_path,
	mime_type,
	size_bytes,
	content_sha256,
	sort_order,
	is_active,
	created_by_user_uuid,
	created_at,
	updated_at
`
