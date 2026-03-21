package models

import "context"

const listAdminProblems = `-- name: ListAdminProblems :many
SELECT p.id, p.title, p.slug, p.description, p.difficulty, p.topic_id, p.created_by, p.init_script, p.solution_query, p.supported_databases, p.order_matters, p.hints, p.sample_output, p.is_public, p.is_active, p.created_at, p.updated_at, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2
`

func (q *Queries) ListAdminProblems(ctx context.Context, arg ListProblemsParams) ([]ListProblemsRow, error) {
	rows, err := q.db.Query(ctx, listAdminProblems, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListProblemsRow{}
	for rows.Next() {
		var i ListProblemsRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Slug,
			&i.Description,
			&i.Difficulty,
			&i.TopicID,
			&i.CreatedBy,
			&i.InitScript,
			&i.SolutionQuery,
			&i.SupportedDatabases,
			&i.OrderMatters,
			&i.Hints,
			&i.SampleOutput,
			&i.IsPublic,
			&i.IsActive,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.TopicName,
			&i.TopicSlug,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAdminProblemsByTopic = `-- name: ListAdminProblemsByTopic :many
SELECT p.id, p.title, p.slug, p.description, p.difficulty, p.topic_id, p.created_by, p.init_script, p.solution_query, p.supported_databases, p.order_matters, p.hints, p.sample_output, p.is_public, p.is_active, p.created_at, p.updated_at, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.topic_id = $1 AND p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3
`

func (q *Queries) ListAdminProblemsByTopic(ctx context.Context, arg ListProblemsByTopicParams) ([]ListProblemsByTopicRow, error) {
	rows, err := q.db.Query(ctx, listAdminProblemsByTopic, arg.TopicID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListProblemsByTopicRow{}
	for rows.Next() {
		var i ListProblemsByTopicRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Slug,
			&i.Description,
			&i.Difficulty,
			&i.TopicID,
			&i.CreatedBy,
			&i.InitScript,
			&i.SolutionQuery,
			&i.SupportedDatabases,
			&i.OrderMatters,
			&i.Hints,
			&i.SampleOutput,
			&i.IsPublic,
			&i.IsActive,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.TopicName,
			&i.TopicSlug,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAdminProblemsByDifficulty = `-- name: ListAdminProblemsByDifficulty :many
SELECT p.id, p.title, p.slug, p.description, p.difficulty, p.topic_id, p.created_by, p.init_script, p.solution_query, p.supported_databases, p.order_matters, p.hints, p.sample_output, p.is_public, p.is_active, p.created_at, p.updated_at, t.name as topic_name, t.slug as topic_slug
FROM problems p
LEFT JOIN topics t ON t.id = p.topic_id
WHERE p.difficulty = $1 AND p.is_active = TRUE
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3
`

func (q *Queries) ListAdminProblemsByDifficulty(ctx context.Context, arg ListProblemsByDifficultyParams) ([]ListProblemsByDifficultyRow, error) {
	rows, err := q.db.Query(ctx, listAdminProblemsByDifficulty, arg.Difficulty, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListProblemsByDifficultyRow{}
	for rows.Next() {
		var i ListProblemsByDifficultyRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Slug,
			&i.Description,
			&i.Difficulty,
			&i.TopicID,
			&i.CreatedBy,
			&i.InitScript,
			&i.SolutionQuery,
			&i.SupportedDatabases,
			&i.OrderMatters,
			&i.Hints,
			&i.SampleOutput,
			&i.IsPublic,
			&i.IsActive,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.TopicName,
			&i.TopicSlug,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
