-- +goose Up
-- +goose StatementBegin

-- 1. PROBLEM_TEST_CASES: Đề bài với nhiều bộ test cases
CREATE TABLE problem_test_cases (
    id BIGSERIAL PRIMARY KEY,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    name VARCHAR(100),                               -- Tên test case (vd: Sample 1, Edge Case)
    description TEXT,
    init_script TEXT NOT NULL,                        -- Script khởi tạo riêng cho test này
    solution_query TEXT NOT NULL,                     -- Query đáp án chuẩn cho test này
    weight INT DEFAULT 1,                            -- Trọng số điểm (mặc định 1)
    is_hidden BOOLEAN DEFAULT FALSE,                  -- Có ẩn chi tiết test case với sinh viên?
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. SUBMISSION_TEST_RESULTS: Chi tiết kết quả từng test của bài nộp
CREATE TABLE submission_test_results (
    id BIGSERIAL PRIMARY KEY,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    test_case_id BIGINT NOT NULL REFERENCES problem_test_cases(id) ON DELETE CASCADE,
    
    status VARCHAR(20) NOT NULL,                      -- accepted, wrong_answer, error, timeout
    execution_time_ms INT,
    
    actual_output JSONB,                              -- Kết quả sinh viên làm
    error_message TEXT,
    
    is_correct BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Thêm cột tính điểm vào bảng submissions (tổng hợp từ các test cases)
ALTER TABLE submissions ADD COLUMN score DECIMAL(5,2) DEFAULT 0;
ALTER TABLE submissions ADD COLUMN total_test_cases INT DEFAULT 0;
ALTER TABLE submissions ADD COLUMN passed_test_cases INT DEFAULT 0;

-- Indexing
CREATE INDEX idx_test_cases_problem ON problem_test_cases(problem_id);
CREATE INDEX idx_test_results_submission ON submission_test_results(submission_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS submission_test_results;
DROP TABLE IF EXISTS problem_test_cases;

ALTER TABLE submissions DROP COLUMN IF EXISTS score;
ALTER TABLE submissions DROP COLUMN IF EXISTS total_test_cases;
ALTER TABLE submissions DROP COLUMN IF EXISTS passed_test_cases;

-- +goose StatementEnd
