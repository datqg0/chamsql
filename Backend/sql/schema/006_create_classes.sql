-- +goose Up
-- +goose StatementBegin

-- =============================================
-- CLASSES: Giảng viên quản lý lớp học
-- =============================================

CREATE TABLE IF NOT EXISTS classes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    code VARCHAR(20) UNIQUE NOT NULL,                -- Mã lớp để sinh viên tham gia
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    semester VARCHAR(20),                           -- Ví dụ: Fall 2024, Spring 2025
    year INT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- =============================================
-- CLASS_MEMBERS: Sinh viên trong lớp
-- =============================================

CREATE TABLE IF NOT EXISTS class_members (
    id BIGSERIAL PRIMARY KEY,
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',              -- member, ta (teaching assistant)
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(class_id, user_id)
);

-- =============================================
-- CLASS_EXAMS: Ánh xạ kỳ thi đến lớp học
-- =============================================

CREATE TABLE IF NOT EXISTS class_exams (
    id BIGSERIAL PRIMARY KEY,
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(class_id, exam_id)
);

-- =============================================
-- UPDATE EXAM_PROBLEMS: Thêm cột scoring mode
-- =============================================

ALTER TABLE exam_problems ADD COLUMN IF NOT EXISTS scoring_mode VARCHAR(20) DEFAULT 'auto';  -- auto, answer_key, manual
ALTER TABLE exam_problems ADD COLUMN IF NOT EXISTS reference_answer TEXT;                     -- Câu trả lời tham khảo

-- =============================================
-- UPDATE EXAM_SUBMISSIONS: Thêm cột graded info
-- =============================================

ALTER TABLE exam_submissions ADD COLUMN IF NOT EXISTS graded_by BIGINT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE exam_submissions ADD COLUMN IF NOT EXISTS graded_at TIMESTAMPTZ;

-- =============================================
-- INDEXES
-- =============================================

CREATE INDEX idx_classes_created_by ON classes(created_by);
CREATE INDEX idx_classes_code ON classes(code);
CREATE INDEX idx_classes_active ON classes(is_active);

CREATE INDEX idx_class_members_class ON class_members(class_id);
CREATE INDEX idx_class_members_user ON class_members(user_id);

CREATE INDEX idx_class_exams_class ON class_exams(class_id);
CREATE INDEX idx_class_exams_exam ON class_exams(exam_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_class_exams_exam;
DROP INDEX IF EXISTS idx_class_exams_class;
DROP INDEX IF EXISTS idx_class_members_user;
DROP INDEX IF EXISTS idx_class_members_class;
DROP INDEX IF EXISTS idx_classes_active;
DROP INDEX IF EXISTS idx_classes_code;
DROP INDEX IF EXISTS idx_classes_created_by;

ALTER TABLE exam_submissions DROP COLUMN IF EXISTS graded_at;
ALTER TABLE exam_submissions DROP COLUMN IF EXISTS graded_by;

ALTER TABLE exam_problems DROP COLUMN IF EXISTS reference_answer;
ALTER TABLE exam_problems DROP COLUMN IF EXISTS scoring_mode;

DROP TABLE IF EXISTS class_exams;
DROP TABLE IF EXISTS class_members;
DROP TABLE IF EXISTS classes;

-- +goose StatementEnd
