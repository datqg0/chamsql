-- =============================================
-- EXAMS
-- =============================================

-- name: CreateExam :one
INSERT INTO exams (
    title, description, created_by, start_time, end_time, duration_minutes,
    allowed_databases, allow_ai_assistance, shuffle_problems, 
    show_result_immediately, max_attempts, is_public, status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetExamByID :one
SELECT e.*, u.full_name as creator_name
FROM exams e
JOIN users u ON u.id = e.created_by
WHERE e.id = $1;

-- name: ListExams :many
SELECT e.*, u.full_name as creator_name,
    (SELECT COUNT(*) FROM exam_problems WHERE exam_id = e.id) as problem_count,
    (SELECT COUNT(*) FROM exam_participants WHERE exam_id = e.id) as participant_count
FROM exams e
JOIN users u ON u.id = e.created_by
ORDER BY e.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListExamsByLecturer :many
SELECT e.*, 
    (SELECT COUNT(*) FROM exam_problems WHERE exam_id = e.id) as problem_count,
    (SELECT COUNT(*) FROM exam_participants WHERE exam_id = e.id) as participant_count
FROM exams e
WHERE e.created_by = $1
ORDER BY e.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPublicExams :many
SELECT e.*, u.full_name as creator_name,
    (SELECT COUNT(*) FROM exam_problems WHERE exam_id = e.id) as problem_count
FROM exams e
JOIN users u ON u.id = e.created_by
WHERE e.is_public = TRUE AND e.status = 'published'
ORDER BY e.start_time DESC
LIMIT $1 OFFSET $2;

-- name: UpdateExam :one
UPDATE exams SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    start_time = COALESCE(sqlc.narg('start_time'), start_time),
    end_time = COALESCE(sqlc.narg('end_time'), end_time),
    duration_minutes = COALESCE(sqlc.narg('duration_minutes'), duration_minutes),
    allow_ai_assistance = COALESCE(sqlc.narg('allow_ai_assistance'), allow_ai_assistance),
    shuffle_problems = COALESCE(sqlc.narg('shuffle_problems'), shuffle_problems),
    show_result_immediately = COALESCE(sqlc.narg('show_result_immediately'), show_result_immediately),
    max_attempts = COALESCE(sqlc.narg('max_attempts'), max_attempts),
    is_public = COALESCE(sqlc.narg('is_public'), is_public),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateExamStatus :one
UPDATE exams SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteExam :exec
DELETE FROM exams WHERE id = $1;

-- =============================================
-- EXAM PROBLEMS
-- =============================================

-- name: AddProblemToExam :one
INSERT INTO exam_problems (exam_id, problem_id, points, sort_order)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListExamProblems :many
SELECT ep.*, p.title, p.slug, p.difficulty, p.description
FROM exam_problems ep
JOIN problems p ON p.id = ep.problem_id
WHERE ep.exam_id = $1
ORDER BY ep.sort_order ASC;

-- name: RemoveProblemFromExam :exec
DELETE FROM exam_problems WHERE exam_id = $1 AND problem_id = $2;

-- name: UpdateExamProblemPoints :one
UPDATE exam_problems SET points = $3
WHERE exam_id = $1 AND problem_id = $2
RETURNING *;

-- =============================================
-- EXAM PARTICIPANTS
-- =============================================

-- name: AddParticipant :one
INSERT INTO exam_participants (exam_id, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetParticipant :one
SELECT ep.*, u.full_name, u.email, u.student_id
FROM exam_participants ep
JOIN users u ON u.id = ep.user_id
WHERE ep.exam_id = $1 AND ep.user_id = $2;

-- name: ListExamParticipants :many
SELECT ep.*, u.full_name, u.email, u.student_id
FROM exam_participants ep
JOIN users u ON u.id = ep.user_id
WHERE ep.exam_id = $1
ORDER BY u.full_name ASC;

-- name: StartExam :one
UPDATE exam_participants SET 
    status = 'in_progress',
    started_at = NOW()
WHERE exam_id = $1 AND user_id = $2
RETURNING *;

-- name: SubmitExam :one
UPDATE exam_participants SET 
    status = 'submitted',
    submitted_at = NOW()
WHERE exam_id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateParticipantScore :one
UPDATE exam_participants SET 
    total_score = $3,
    status = 'graded'
WHERE exam_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveParticipant :exec
DELETE FROM exam_participants WHERE exam_id = $1 AND user_id = $2;

-- name: ListUserExams :many
SELECT e.*, ep.status as participation_status, ep.total_score, ep.started_at, ep.submitted_at
FROM exam_participants ep
JOIN exams e ON e.id = ep.exam_id
WHERE ep.user_id = $1
ORDER BY e.start_time DESC;

-- =============================================
-- EXAM SUBMISSIONS
-- =============================================

-- name: CreateExamSubmission :one
INSERT INTO exam_submissions (
    exam_id, exam_problem_id, user_id, code, database_type, status,
    execution_time_ms, expected_output, actual_output, error_message, 
    is_correct, score, attempt_number
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetExamSubmission :one
SELECT * FROM exam_submissions
WHERE exam_id = $1 AND exam_problem_id = $2 AND user_id = $3
ORDER BY submitted_at DESC
LIMIT 1;

-- name: ListUserExamSubmissions :many
SELECT es.*, ep.points as max_points, p.title as problem_title
FROM exam_submissions es
JOIN exam_problems ep ON ep.id = es.exam_problem_id
JOIN problems p ON p.id = ep.problem_id
WHERE es.exam_id = $1 AND es.user_id = $2
ORDER BY es.submitted_at DESC;

-- name: CountUserExamSubmissions :one
SELECT COUNT(*) FROM exam_submissions
WHERE exam_id = $1 AND exam_problem_id = $2 AND user_id = $3;

-- name: GetExamResults :many
SELECT 
    u.id as user_id, u.full_name, u.student_id,
    ep.total_score, ep.started_at, ep.submitted_at, ep.status
FROM exam_participants ep
JOIN users u ON u.id = ep.user_id
WHERE ep.exam_id = $1
ORDER BY ep.total_score DESC;
