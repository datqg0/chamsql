-- name: CreateProblemTestCase :one
INSERT INTO problem_test_cases (
    problem_id, name, description, init_script, solution_query, weight, is_hidden
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetProblemTestCaseByID :one
SELECT * FROM problem_test_cases WHERE id = $1;

-- name: ListProblemTestCases :many
SELECT * FROM problem_test_cases
WHERE problem_id = $1
ORDER BY created_at ASC;

-- name: UpdateProblemTestCase :one
UPDATE problem_test_cases SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    init_script = COALESCE(sqlc.narg('init_script'), init_script),
    solution_query = COALESCE(sqlc.narg('solution_query'), solution_query),
    weight = COALESCE(sqlc.narg('weight'), weight),
    is_hidden = COALESCE(sqlc.narg('is_hidden'), is_hidden),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProblemTestCase :exec
DELETE FROM problem_test_cases WHERE id = $1;

-- name: DeleteAllProblemTestCases :exec
DELETE FROM problem_test_cases WHERE problem_id = $1;

-- name: CreateSubmissionTestResult :one
INSERT INTO submission_test_results (
    submission_id, test_case_id, status, execution_time_ms, actual_output, error_message, is_correct
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListSubmissionTestResults :many
SELECT tr.*, tc.name as test_case_name, tc.is_hidden
FROM submission_test_results tr
JOIN problem_test_cases tc ON tc.id = tr.test_case_id
WHERE tr.submission_id = $1
ORDER BY tc.created_at ASC;

-- name: UpdateSubmissionScore :exec
UPDATE submissions SET
    score = $2,
    total_test_cases = $3,
    passed_test_cases = $4
WHERE id = $1;
