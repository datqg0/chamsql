-- name: CreateSubmission :one
INSERT INTO submissions (
    user_id, problem_id, code, database_type, status,
    execution_time_ms, expected_output, actual_output, error_message, is_correct
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetSubmissionByID :one
SELECT s.*, p.title as problem_title, p.slug as problem_slug
FROM submissions s
JOIN problems p ON p.id = s.problem_id
WHERE s.id = $1;

-- name: ListUserSubmissions :many
SELECT s.*, p.title as problem_title, p.slug as problem_slug
FROM submissions s
JOIN problems p ON p.id = s.problem_id
WHERE s.user_id = $1
ORDER BY s.submitted_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUserSubmissionsForProblem :many
SELECT * FROM submissions
WHERE user_id = $1 AND problem_id = $2
ORDER BY submitted_at DESC
LIMIT $3;

-- name: CountUserSubmissions :one
SELECT COUNT(*) FROM submissions WHERE user_id = $1;

-- name: CountCorrectSubmissions :one
SELECT COUNT(*) FROM submissions WHERE user_id = $1 AND is_correct = TRUE;

-- name: GetLatestSubmission :one
SELECT * FROM submissions
WHERE user_id = $1 AND problem_id = $2
ORDER BY submitted_at DESC
LIMIT 1;
