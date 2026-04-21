-- +goose Up
-- +goose StatementBegin

-- 1. PDF Uploads: Lưu trữ file PDF giảng viên upload
CREATE TABLE pdf_uploads (
    id BIGSERIAL PRIMARY KEY,
    lecturer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_path VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending', -- pending, parsing, generating, completed, failed
    extraction_result JSONB,                         -- Kết quả parse PDF (danh sách bài)
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. AI Generated Content: Nội dung AI sinh ra từ PDF
CREATE TABLE ai_generated_content (
    id BIGSERIAL PRIMARY KEY,
    pdf_upload_id BIGINT NOT NULL REFERENCES pdf_uploads(id) ON DELETE CASCADE,
    problem_number INT NOT NULL,
    content_type VARCHAR(50) NOT NULL,           -- 'solution', 'description', 'test_case'
    original_content TEXT,
    ai_generated_content TEXT,
    confidence_score DECIMAL(5,4) DEFAULT 0,
    ai_provider VARCHAR(50),
    is_approved BOOLEAN DEFAULT FALSE,
    lecturer_notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Problem Review Queue: Hàng đợi giảng viên review bài từ PDF
CREATE TABLE problem_review_queue (
    id BIGSERIAL PRIMARY KEY,
    pdf_upload_id BIGINT NOT NULL REFERENCES pdf_uploads(id) ON DELETE CASCADE,
    problem_number INT NOT NULL,
    problem_draft JSONB NOT NULL,                -- Nháp bài (ProblemDraft)
    edits_made JSONB,                            -- Các chỉnh sửa của giảng viên
    status VARCHAR(30) NOT NULL DEFAULT 'pending', -- pending, editing, approved, rejected
    reviewer_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Test Case Templates: Template test case cho các bài
CREATE TABLE test_case_templates (
    id BIGSERIAL PRIMARY KEY,
    problem_id BIGINT REFERENCES problems(id) ON DELETE CASCADE,
    test_case_number INT NOT NULL,
    description TEXT,
    schema_sql TEXT,
    test_data_sql TEXT NOT NULL,
    expected_output JSONB,
    is_public BOOLEAN DEFAULT TRUE,
    difficulty VARCHAR(20) DEFAULT 'medium',
    is_validated BOOLEAN DEFAULT FALSE,
    validation_status VARCHAR(30),              -- valid, invalid, error
    validation_error TEXT,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Excel Exports: Lưu lịch sử xuất báo cáo
CREATE TABLE excel_exports (
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    export_type VARCHAR(50) NOT NULL,           -- 'full_results', 'summary', 'attendance'
    file_path VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    row_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_pdf_uploads_lecturer ON pdf_uploads(lecturer_id);
CREATE INDEX idx_pdf_uploads_status ON pdf_uploads(status);
CREATE INDEX idx_ai_content_upload ON ai_generated_content(pdf_upload_id);
CREATE INDEX idx_review_queue_upload ON problem_review_queue(pdf_upload_id);
CREATE INDEX idx_review_queue_status ON problem_review_queue(status);
CREATE INDEX idx_test_templates_problem ON test_case_templates(problem_id);
CREATE INDEX idx_excel_exports_exam ON excel_exports(exam_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS excel_exports;
DROP TABLE IF EXISTS test_case_templates;
DROP TABLE IF EXISTS problem_review_queue;
DROP TABLE IF EXISTS ai_generated_content;
DROP TABLE IF EXISTS pdf_uploads;

-- +goose StatementEnd
