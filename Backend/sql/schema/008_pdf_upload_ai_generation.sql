-- +goose Up
-- +goose StatementBegin

-- =============================================
-- PHASE 4: PDF Upload + AI Problem Generation
-- =============================================

-- 1. PDF_UPLOADS: Track uploaded files
CREATE TABLE pdf_uploads (
    id BIGSERIAL PRIMARY KEY,
    lecturer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_path VARCHAR(255) NOT NULL,           -- MinIO path
    file_name VARCHAR(255) NOT NULL,           -- Original filename
    original_filename VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'uploading', -- uploading, parsing, generating, completed, failed
    extraction_result JSONB,                    -- Parsed content {problems: [...]}
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. AI_GENERATED_CONTENT: Cache AI outputs
CREATE TABLE ai_generated_content (
    id BIGSERIAL PRIMARY KEY,
    pdf_upload_id BIGINT REFERENCES pdf_uploads(id) ON DELETE CASCADE,
    problem_number INT,                         -- Which problem in PDF
    content_type VARCHAR(50) NOT NULL,          -- 'solution', 'test_case', 'description'
    original_content TEXT,                      -- From PDF
    ai_generated_content TEXT,                  -- AI output
    confidence_score DECIMAL(3,2),              -- 0-1 (AI confidence)
    ai_provider VARCHAR(50),                    -- 'pattern', 'huggingface'
    is_approved BOOLEAN DEFAULT FALSE,
    lecturer_notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(pdf_upload_id, problem_number, content_type)
);

-- 3. TEST_CASE_TEMPLATES: AI-generated test cases
CREATE TABLE test_case_templates (
    id BIGSERIAL PRIMARY KEY,
    problem_id BIGINT REFERENCES problems(id) ON DELETE CASCADE,
    test_case_number INT,
    description TEXT,                           -- What this test case tests
    schema_sql TEXT NOT NULL,                   -- CREATE TABLE + setup
    test_data_sql TEXT NOT NULL,                -- INSERT statements
    expected_output JSONB NOT NULL,             -- Expected result
    is_public BOOLEAN DEFAULT FALSE,            -- Public (student sees) or hidden
    difficulty VARCHAR(20),                     -- For test case complexity
    created_by BIGINT REFERENCES users(id),
    is_validated BOOLEAN DEFAULT FALSE,
    validation_status VARCHAR(50),              -- 'passed', 'failed', 'error'
    validation_error TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(problem_id, test_case_number)
);

-- 4. PROBLEM_REVIEW_QUEUE: Track pending reviews
CREATE TABLE problem_review_queue (
    id BIGSERIAL PRIMARY KEY,
    pdf_upload_id BIGINT REFERENCES pdf_uploads(id) ON DELETE CASCADE,
    problem_number INT,                         -- Which problem from PDF
    problem_draft JSONB NOT NULL,               -- {title, description, difficulty, solution, test_cases}
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, editing
    reviewer_id BIGINT REFERENCES users(id),
    review_notes TEXT,
    edits_made JSONB,                           -- Track what lecturer edited
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(pdf_upload_id, problem_number)
);

-- 5. EXCEL_EXPORTS: Track exported files
CREATE TABLE excel_exports (
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT REFERENCES exams(id) ON DELETE CASCADE,
    export_type VARCHAR(50) NOT NULL,           -- 'results', 'analytics', 'submissions'
    file_path VARCHAR(255),                     -- MinIO path
    file_name VARCHAR(255),
    created_by BIGINT REFERENCES users(id),
    row_count INT,                              -- How many rows exported
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(exam_id, export_type, created_at)
);

-- 6. Modify problems table - Add AI metadata
ALTER TABLE problems 
    ADD COLUMN IF NOT EXISTS ai_generated BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS ai_provider VARCHAR(50),
    ADD COLUMN IF NOT EXISTS ai_confidence_score DECIMAL(3,2);

-- 7. Modify exam_submissions - Add test case results
ALTER TABLE exam_submissions 
    ADD COLUMN IF NOT EXISTS test_case_results JSONB; -- {passed: 3, total: 5, details: [...]}

-- =============================================
-- INDEXES
-- =============================================

CREATE INDEX idx_pdf_uploads_lecturer ON pdf_uploads(lecturer_id);
CREATE INDEX idx_pdf_uploads_status ON pdf_uploads(status);
CREATE INDEX idx_ai_generated_content_pdf ON ai_generated_content(pdf_upload_id);
CREATE INDEX idx_ai_generated_content_type ON ai_generated_content(content_type);
CREATE INDEX idx_test_case_templates_problem ON test_case_templates(problem_id);
CREATE INDEX idx_test_case_templates_public ON test_case_templates(problem_id, is_public);
CREATE INDEX idx_problem_review_queue_status ON problem_review_queue(status);
CREATE INDEX idx_problem_review_queue_pdf ON problem_review_queue(pdf_upload_id);
CREATE INDEX idx_excel_exports_exam ON excel_exports(exam_id);
CREATE INDEX idx_excel_exports_type ON excel_exports(export_type);
CREATE INDEX idx_problems_ai_generated ON problems(ai_generated);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_problems_ai_generated;
DROP INDEX IF EXISTS idx_excel_exports_type;
DROP INDEX IF EXISTS idx_excel_exports_exam;
DROP INDEX IF EXISTS idx_problem_review_queue_pdf;
DROP INDEX IF EXISTS idx_problem_review_queue_status;
DROP INDEX IF EXISTS idx_test_case_templates_public;
DROP INDEX IF EXISTS idx_test_case_templates_problem;
DROP INDEX IF EXISTS idx_ai_generated_content_type;
DROP INDEX IF EXISTS idx_ai_generated_content_pdf;
DROP INDEX IF EXISTS idx_pdf_uploads_status;
DROP INDEX IF EXISTS idx_pdf_uploads_lecturer;

ALTER TABLE exam_submissions DROP COLUMN IF EXISTS test_case_results;
ALTER TABLE problems DROP COLUMN IF EXISTS ai_confidence_score;
ALTER TABLE problems DROP COLUMN IF EXISTS ai_provider;
ALTER TABLE problems DROP COLUMN IF EXISTS ai_generated;

DROP TABLE IF EXISTS excel_exports;
DROP TABLE IF EXISTS problem_review_queue;
DROP TABLE IF EXISTS test_case_templates;
DROP TABLE IF EXISTS ai_generated_content;
DROP TABLE IF EXISTS pdf_uploads;

-- +goose StatementEnd
