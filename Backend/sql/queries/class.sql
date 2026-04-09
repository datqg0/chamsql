-- =============================================
-- CLASSES QUERIES
-- =============================================

-- name: CreateClass :one
INSERT INTO classes (name, description, code, created_by, semester, year)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetClassByID :one
SELECT * FROM classes WHERE id = $1;

-- name: GetClassByCode :one
SELECT * FROM classes WHERE code = $1;

-- name: ListClassesByLecturer :many
SELECT * FROM classes 
WHERE created_by = $1 AND is_active = TRUE
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateClass :one
UPDATE classes 
SET name = COALESCE($2, name),
    description = COALESCE($3, description),
    semester = COALESCE($4, semester),
    year = COALESCE($5, year),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeactivateClass :exec
UPDATE classes SET is_active = FALSE WHERE id = $1;

-- name: DeleteClass :exec
DELETE FROM classes WHERE id = $1;

-- =============================================
-- CLASS_MEMBERS QUERIES
-- =============================================

-- name: AddClassMember :one
INSERT INTO class_members (class_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetClassMember :one
SELECT * FROM class_members 
WHERE class_id = $1 AND user_id = $2;

-- name: ListClassMembers :many
SELECT cm.id, cm.user_id, cm.role, cm.joined_at, u.email, u.full_name
FROM class_members cm
JOIN users u ON cm.user_id = u.id
WHERE cm.class_id = $1
ORDER BY cm.joined_at DESC
LIMIT $2 OFFSET $3;

-- name: RemoveClassMember :exec
DELETE FROM class_members 
WHERE class_id = $1 AND user_id = $2;

-- name: GetStudentClasses :many
SELECT c.* FROM classes c
JOIN class_members cm ON c.id = cm.class_id
WHERE cm.user_id = $1 AND c.is_active = TRUE
ORDER BY c.created_at DESC;

-- name: CountClassMembers :one
SELECT COUNT(*) as count FROM class_members WHERE class_id = $1;

-- =============================================
-- CLASS_EXAMS QUERIES
-- =============================================

-- name: AssignExamToClass :one
INSERT INTO class_exams (class_id, exam_id)
VALUES ($1, $2)
RETURNING *;

-- name: ListClassExams :many
SELECT ce.id, e.* FROM class_exams ce
JOIN exams e ON ce.exam_id = e.id
WHERE ce.class_id = $1
ORDER BY e.start_time DESC;

-- name: RemoveExamFromClass :exec
DELETE FROM class_exams 
WHERE class_id = $1 AND exam_id = $2;

-- name: GetClassExamByID :one
SELECT ce.id, ce.class_id, e.* FROM class_exams ce
JOIN exams e ON ce.exam_id = e.id
WHERE ce.id = $1;
