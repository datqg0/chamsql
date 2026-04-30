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

-- =============================================
-- ADMIN QUERIES (no is_public filter)
-- =============================================

-- name: ListProblemsAdmin :many
SELECT p.*, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListProblemsByTopicAdmin :many
SELECT p.*, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.topic_id = $1 AND p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListProblemsByDifficultyAdmin :many
SELECT p.*, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.difficulty = $1 AND p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountProblemsAdmin :one
SELECT COUNT(*) FROM problems WHERE is_active = TRUE;

-- name: ListProblemsByLecturer :many
SELECT p.*, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.created_by = $1 AND p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- =============================================
-- SEARCH QUERIES
-- =============================================

-- name: SearchProblems :many
SELECT p.*, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.is_active = TRUE AND p.is_public = TRUE
  AND (p.title ILIKE '%' || @search_query::text || '%'
       OR p.description ILIKE '%' || @search_query::text || '%')
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountSearchProblems :one
SELECT COUNT(*) FROM problems
WHERE is_active = TRUE AND is_public = TRUE
  AND (title ILIKE '%' || @search_query::text || '%'
       OR description ILIKE '%' || @search_query::text || '%');

-- name: SearchProblemsAdmin :many
SELECT p.*, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.is_active = TRUE
  AND (p.title ILIKE '%' || @search_query::text || '%'
       OR p.description ILIKE '%' || @search_query::text || '%')
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- =============================================
-- DASHBOARD / ANALYTICS QUERIES
-- =============================================

-- name: GetDailySubmissionStats :many
SELECT DATE(submitted_at) as submit_date,
       COUNT(*) as total_submissions,
       COUNT(CASE WHEN is_correct = TRUE THEN 1 END) as correct_count,
       AVG(execution_time_ms)::int as avg_execution_ms
FROM submissions
WHERE submitted_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(submitted_at)
ORDER BY submit_date DESC;

-- name: GetPassRatePerProblem :many
SELECT p.id, p.title, p.difficulty,
       COUNT(s.id) as total_submissions,
       COUNT(CASE WHEN s.is_correct = TRUE THEN 1 END) as correct_count,
       CASE WHEN COUNT(s.id) > 0
            THEN ROUND(COUNT(CASE WHEN s.is_correct = TRUE THEN 1 END)::numeric / COUNT(s.id) * 100, 2)
            ELSE 0 END as pass_rate
FROM problems p
LEFT JOIN submissions s ON s.problem_id = p.id
WHERE p.is_active = TRUE
GROUP BY p.id, p.title, p.difficulty
ORDER BY total_submissions DESC
LIMIT $1;

-- name: GetActiveUsersWeek :one
SELECT COUNT(DISTINCT user_id) as active_users
FROM submissions
WHERE submitted_at >= NOW() - INTERVAL '7 days';

-- name: GetAvgSolveTime :one
SELECT AVG(best_time_ms)::int as avg_solve_time_ms
FROM user_progress
WHERE is_solved = TRUE AND best_time_ms IS NOT NULL;

-- name: GetTopProblemsBySubmissions :many
SELECT p.id, p.title, p.slug, p.difficulty,
       COUNT(s.id) as submission_count,
       COUNT(DISTINCT s.user_id) as unique_users
FROM problems p
JOIN submissions s ON s.problem_id = p.id
WHERE p.is_active = TRUE
GROUP BY p.id, p.title, p.slug, p.difficulty
ORDER BY submission_count DESC
LIMIT $1;

-- name: GetUserPerformanceTimeline :many
SELECT DATE(s.submitted_at) as submit_date,
       AVG(s.execution_time_ms)::int as avg_time_ms,
       MIN(s.execution_time_ms) as best_time_ms,
       COUNT(*) as submission_count,
       COUNT(CASE WHEN s.is_correct THEN 1 END) as correct_count
FROM submissions s
WHERE s.user_id = $1
  AND ($2::bigint IS NULL OR s.problem_id = $2)
  AND s.submitted_at >= NOW() - INTERVAL '90 days'
GROUP BY DATE(s.submitted_at)
ORDER BY submit_date ASC;

-- name: GetSystemGradingStats :one
SELECT
    COUNT(*) as total_submissions,
    AVG(execution_time_ms)::int as avg_grading_time_ms,
    MIN(execution_time_ms) as min_grading_time_ms,
    MAX(execution_time_ms) as max_grading_time_ms,
    COUNT(CASE WHEN is_correct THEN 1 END) as total_correct,
    COUNT(DISTINCT user_id) as total_users,
    COUNT(DISTINCT problem_id) as total_problems_attempted
FROM submissions
WHERE submitted_at >= NOW() - INTERVAL '30 days';
