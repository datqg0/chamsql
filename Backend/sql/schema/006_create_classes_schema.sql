-- +goose Up
-- +goose StatementBegin

-- =============================================
-- CLASSES MANAGEMENT - Database Schema
-- =============================================

-- Classes table - for grouping students and assigning exams
CREATE TABLE classes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    semester INT NOT NULL,
    year INT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Class members - students enrolled in a class
CREATE TABLE class_members (
    id BIGSERIAL PRIMARY KEY,
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'student', -- 'student', 'ta' (teaching assistant)
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(class_id, user_id)
);

-- Class exams - exams assigned to a class
CREATE TABLE class_exams (
    id BIGSERIAL PRIMARY KEY,
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(class_id, exam_id)
);

-- Indexes for better query performance
CREATE INDEX idx_classes_created_by ON classes(created_by);
CREATE INDEX idx_classes_is_active ON classes(is_active);
CREATE INDEX idx_class_members_class_id ON class_members(class_id);
CREATE INDEX idx_class_members_user_id ON class_members(user_id);
CREATE INDEX idx_class_exams_class_id ON class_exams(class_id);
CREATE INDEX idx_class_exams_exam_id ON class_exams(exam_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS class_exams CASCADE;
DROP TABLE IF EXISTS class_members CASCADE;
DROP TABLE IF EXISTS classes CASCADE;

-- +goose StatementEnd
