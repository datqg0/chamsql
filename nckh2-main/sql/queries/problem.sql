-- name: CreateProblem :one
INSERT INTO problems (
    title, slug, description, difficulty, topic_id, created_by,
    init_script, solution_query, supported_databases, order_matters,
    hints, sample_output, is_public
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetProblemByID :one
SELECT * FROM problems WHERE id = $1;

-- name: GetProblemBySlug :one
SELECT * FROM problems WHERE slug = $1 AND is_active = TRUE;

-- name: ListProblems :many
SELECT p.*, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.is_public = TRUE AND p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListProblemsByTopic :many
SELECT p.*, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.topic_id = $1 AND p.is_public = TRUE AND p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListProblemsByDifficulty :many
SELECT p.*, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.difficulty = $1 AND p.is_public = TRUE AND p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateProblem :one
UPDATE problems SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    difficulty = COALESCE(sqlc.narg('difficulty'), difficulty),
    topic_id = COALESCE(sqlc.narg('topic_id'), topic_id),
    init_script = COALESCE(sqlc.narg('init_script'), init_script),
    solution_query = COALESCE(sqlc.narg('solution_query'), solution_query),
    hints = COALESCE(sqlc.narg('hints'), hints),
    sample_output = COALESCE(sqlc.narg('sample_output'), sample_output),
    order_matters = COALESCE(sqlc.narg('order_matters'), order_matters),
    is_public = COALESCE(sqlc.narg('is_public'), is_public),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProblem :exec
UPDATE problems SET is_active = FALSE, updated_at = NOW() WHERE id = $1;

-- name: CountProblems :one
SELECT COUNT(*) FROM problems WHERE is_public = TRUE AND is_active = TRUE;

-- name: CountProblemsByDifficulty :many
SELECT difficulty, COUNT(*) as count
FROM problems
WHERE is_public = TRUE AND is_active = TRUE
GROUP BY difficulty;

-- name: GetProblemWithUserProgress :one
SELECT 
    p.*,
    t.name as topic_name,
    t.slug as topic_slug,
    up.is_solved,
    up.attempts,
    up.best_time_ms
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
LEFT JOIN user_progress up ON up.problem_id = p.id AND up.user_id = $2
WHERE p.slug = $1 AND p.is_active = TRUE;
