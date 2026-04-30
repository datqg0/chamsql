-- name: CreateTopic :one
INSERT INTO topics (name, slug, description, icon, sort_order, parent_id, level)
VALUES ($1, $2, $3, $4, $5, $6, $7)
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
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    parent_id = COALESCE(sqlc.narg('parent_id'), parent_id),
    level = COALESCE(sqlc.narg('level'), level)
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

-- =============================================
-- TOPIC TREE QUERIES
-- =============================================

-- name: GetTopicTree :many
SELECT t.id, t.name, t.slug, t.description, t.icon, t.sort_order,
       t.parent_id, t.level, t.is_active, t.created_at,
       p.name as parent_name, p.slug as parent_slug,
       COUNT(pr.id) as problem_count
FROM topics t
LEFT JOIN topics p ON p.id = t.parent_id
LEFT JOIN problems pr ON pr.topic_id = t.id AND pr.is_active = TRUE
WHERE t.is_active = TRUE
GROUP BY t.id, t.name, t.slug, t.description, t.icon, t.sort_order,
         t.parent_id, t.level, t.is_active, t.created_at,
         p.name, p.slug
ORDER BY t.parent_id NULLS FIRST, t.sort_order ASC, t.name ASC;

-- name: GetTopicChildren :many
SELECT t.id, t.name, t.slug, t.description, t.icon, t.sort_order,
       t.parent_id, t.level, t.is_active, t.created_at,
       COUNT(p.id) as problem_count
FROM topics t
LEFT JOIN problems p ON p.topic_id = t.id AND p.is_active = TRUE
WHERE t.parent_id = $1 AND t.is_active = TRUE
GROUP BY t.id, t.name, t.slug, t.description, t.icon, t.sort_order,
         t.parent_id, t.level, t.is_active, t.created_at
ORDER BY t.sort_order ASC;

-- name: GetTopicWithAncestors :many
WITH RECURSIVE topic_tree AS (
  SELECT t.id, t.name, t.slug, t.description, t.icon, t.sort_order, t.parent_id, t.level, t.is_active, t.created_at
  FROM topics t WHERE t.id = $1
  UNION ALL
  SELECT t2.id, t2.name, t2.slug, t2.description, t2.icon, t2.sort_order, t2.parent_id, t2.level, t2.is_active, t2.created_at
  FROM topics t2
  JOIN topic_tree tt ON t2.id = tt.parent_id
)
SELECT id, name, slug, description, icon, sort_order, parent_id, level, is_active, created_at FROM topic_tree;
