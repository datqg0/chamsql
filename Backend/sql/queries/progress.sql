-- name: UpsertProgress :one
INSERT INTO user_progress (user_id, problem_id, attempts, first_attempted_at, last_attempted_at)
VALUES ($1, $2, 1, NOW(), NOW())
ON CONFLICT (user_id, problem_id) 
DO UPDATE SET 
    attempts = user_progress.attempts + 1,
    last_attempted_at = NOW()
RETURNING *;

-- name: MarkProblemSolved :one
UPDATE user_progress SET 
    is_solved = TRUE,
    best_time_ms = CASE 
        WHEN best_time_ms IS NULL THEN $3
        WHEN $3 < best_time_ms THEN $3
        ELSE best_time_ms
    END,
    solved_at = COALESCE(solved_at, NOW())
WHERE user_id = $1 AND problem_id = $2
RETURNING *;

-- name: GetUserProgress :one
SELECT * FROM user_progress
WHERE user_id = $1 AND problem_id = $2;

-- name: GetUserStats :one
SELECT 
    COUNT(*) as total_attempted,
    COUNT(*) FILTER (WHERE is_solved = TRUE) as total_solved,
    SUM(attempts) as total_submissions
FROM user_progress
WHERE user_id = $1;

-- name: GetUserStatsByDifficulty :many
SELECT 
    p.difficulty,
    COUNT(*) FILTER (WHERE up.is_solved = TRUE) as solved,
    COUNT(DISTINCT p.id) as total
FROM problems p
LEFT JOIN user_progress up ON up.problem_id = p.id AND up.user_id = $1
WHERE p.is_public = TRUE AND p.is_active = TRUE
GROUP BY p.difficulty;

-- name: ListSolvedProblems :many
SELECT up.*, p.title, p.slug, p.difficulty
FROM user_progress up
JOIN problems p ON p.id = up.problem_id
WHERE up.user_id = $1 AND up.is_solved = TRUE
ORDER BY up.solved_at DESC
LIMIT $2 OFFSET $3;

-- name: ListRecentAttempts :many
SELECT up.*, p.title, p.slug, p.difficulty
FROM user_progress up
JOIN problems p ON p.id = up.problem_id
WHERE up.user_id = $1
ORDER BY up.last_attempted_at DESC
LIMIT $2;
