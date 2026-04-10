-- PDF Upload Queries

-- name: CreatePDFUpload :one
INSERT INTO pdf_uploads (
    lecturer_id, file_path, file_name, original_filename, status
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPDFUploadByID :one
SELECT * FROM pdf_uploads WHERE id = $1;

-- name: GetPDFUploadsByLecturer :many
SELECT * FROM pdf_uploads 
WHERE lecturer_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdatePDFUploadStatus :one
UPDATE pdf_uploads 
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePDFUploadWithExtraction :one
UPDATE pdf_uploads 
SET 
    status = $2,
    extraction_result = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePDFUploadError :one
UPDATE pdf_uploads 
SET 
    status = 'failed',
    error_message = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- AI Generated Content Queries

-- name: CreateAIGeneratedContent :one
INSERT INTO ai_generated_content (
    pdf_upload_id, problem_number, content_type, original_content,
    ai_generated_content, confidence_score, ai_provider
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAIGeneratedContentByProblem :many
SELECT * FROM ai_generated_content
WHERE pdf_upload_id = $1 AND problem_number = $2
ORDER BY created_at DESC;

-- name: GetAIGeneratedContentByType :many
SELECT * FROM ai_generated_content
WHERE pdf_upload_id = $1 AND content_type = $2
ORDER BY problem_number;

-- name: UpdateAIGeneratedContentApproval :one
UPDATE ai_generated_content
SET 
    is_approved = $2,
    lecturer_notes = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- Problem Review Queue Queries

-- name: CreateProblemReviewQueue :one
INSERT INTO problem_review_queue (
    pdf_upload_id, problem_number, problem_draft, status
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProblemReviewQueueByID :one
SELECT * FROM problem_review_queue WHERE id = $1;

-- name: GetProblemReviewQueueByPDF :many
SELECT * FROM problem_review_queue
WHERE pdf_upload_id = $1
ORDER BY problem_number;

-- name: GetProblemReviewQueueByStatus :many
SELECT * FROM problem_review_queue
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateProblemReviewStatus :one
UPDATE problem_review_queue
SET 
    status = $2,
    reviewer_id = $3,
    review_notes = $4,
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateProblemReviewDraft :one
UPDATE problem_review_queue
SET 
    problem_draft = $2,
    edits_made = $3,
    status = 'editing',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- Test Case Templates Queries

-- name: CreateTestCaseTemplate :one
INSERT INTO test_case_templates (
    problem_id, test_case_number, description, schema_sql,
    test_data_sql, expected_output, is_public, difficulty, created_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetTestCaseTemplatesByProblem :many
SELECT * FROM test_case_templates
WHERE problem_id = $1
ORDER BY test_case_number;

-- name: GetPublicTestCaseTemplates :many
SELECT * FROM test_case_templates
WHERE problem_id = $1 AND is_public = TRUE
ORDER BY test_case_number;

-- name: UpdateTestCaseValidation :one
UPDATE test_case_templates
SET 
    is_validated = $2,
    validation_status = $3,
    validation_error = $4
WHERE id = $1
RETURNING *;

-- Excel Export Queries

-- name: CreateExcelExport :one
INSERT INTO excel_exports (
    exam_id, export_type, file_path, file_name, created_by, row_count
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetExcelExportsByExam :many
SELECT * FROM excel_exports
WHERE exam_id = $1
ORDER BY created_at DESC;

-- name: GetLatestExcelExport :one
SELECT * FROM excel_exports
WHERE exam_id = $1 AND export_type = $2
ORDER BY created_at DESC
LIMIT 1;
