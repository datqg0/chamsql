-- name: CreateTopic :one
INSERT INTO topics (name, slug, description, icon, sort_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTopicByID :one
SELECT * FROM topics WHERE id = $1;

-- name: GetTopicBySlug :one
SELECT * FROM topics WHERE slug = $1;

-- name: ListTopics :many
SELECT * FROM topics 
WHERE is_active = TRUE
ORDER BY sort_order ASC, name ASC;

-- name: UpdateTopic :one
UPDATE topics SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    icon = COALESCE(sqlc.narg('icon'), icon),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order)
WHERE id = $1
RETURNING *;

-- name: DeleteTopic :exec
UPDATE topics SET is_active = FALSE WHERE id = $1;

-- name: CountProblemsPerTopic :many
SELECT t.id, t.name, t.slug, COUNT(p.id) as problem_count
FROM topics t
LEFT JOIN problems p ON p.topic_id = t.id AND p.is_active = TRUE
WHERE t.is_active = TRUE
GROUP BY t.id
ORDER BY t.sort_order ASC;
