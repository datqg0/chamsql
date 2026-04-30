-- +goose Up
-- +goose StatementBegin

-- =============================================
-- Topic Hierarchy: parent_id + level
-- =============================================
ALTER TABLE topics ADD COLUMN IF NOT EXISTS parent_id INT REFERENCES topics(id) ON DELETE SET NULL;
ALTER TABLE topics ADD COLUMN IF NOT EXISTS level INT DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_topics_parent ON topics(parent_id);

-- =============================================
-- Full-text search index for problems
-- =============================================
CREATE INDEX IF NOT EXISTS idx_problems_title_search ON problems USING gin(to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_problems_desc_search ON problems USING gin(to_tsvector('english', description));

-- =============================================
-- Analytics: grading duration tracking
-- =============================================
ALTER TABLE submissions ADD COLUMN IF NOT EXISTS grading_started_at TIMESTAMPTZ;
ALTER TABLE submissions ADD COLUMN IF NOT EXISTS grading_completed_at TIMESTAMPTZ;
ALTER TABLE submissions ADD COLUMN IF NOT EXISTS grading_duration_ms INT;

ALTER TABLE exam_submissions ADD COLUMN IF NOT EXISTS grading_started_at TIMESTAMPTZ;
ALTER TABLE exam_submissions ADD COLUMN IF NOT EXISTS grading_completed_at TIMESTAMPTZ;
ALTER TABLE exam_submissions ADD COLUMN IF NOT EXISTS grading_duration_ms INT;

-- Index for analytics queries
CREATE INDEX IF NOT EXISTS idx_submissions_submitted_date ON submissions(submitted_at);
CREATE INDEX IF NOT EXISTS idx_submissions_correct ON submissions(is_correct);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_submissions_correct;
DROP INDEX IF EXISTS idx_submissions_submitted_date;

ALTER TABLE exam_submissions DROP COLUMN IF EXISTS grading_duration_ms;
ALTER TABLE exam_submissions DROP COLUMN IF EXISTS grading_completed_at;
ALTER TABLE exam_submissions DROP COLUMN IF EXISTS grading_started_at;

ALTER TABLE submissions DROP COLUMN IF EXISTS grading_duration_ms;
ALTER TABLE submissions DROP COLUMN IF EXISTS grading_completed_at;
ALTER TABLE submissions DROP COLUMN IF EXISTS grading_started_at;

DROP INDEX IF EXISTS idx_problems_desc_search;
DROP INDEX IF EXISTS idx_problems_title_search;
DROP INDEX IF EXISTS idx_topics_parent;

ALTER TABLE topics DROP COLUMN IF EXISTS level;
ALTER TABLE topics DROP COLUMN IF EXISTS parent_id;

-- +goose StatementEnd
