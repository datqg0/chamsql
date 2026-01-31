-- +goose Up
-- +goose StatementBegin

-- =============================================
-- SQL EXAM SYSTEM - Database Schema
-- =============================================

-- 1. USERS: Quản lý người dùng (Admin, Lecturer, Student)
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(100) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'student', -- 'admin', 'lecturer', 'student'
    student_id VARCHAR(20),                       -- Mã sinh viên (nullable)
    avatar_url VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. TOPICS: Chủ đề SQL (SELECT, JOIN, Subquery, Aggregate, etc.)
CREATE TABLE topics (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    icon VARCHAR(50),                              -- emoji hoặc icon name
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. PROBLEMS: Bài tập SQL (LeetCode-style)
CREATE TABLE problems (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT NOT NULL,                     -- Markdown content
    difficulty VARCHAR(20) NOT NULL,               -- easy, medium, hard
    topic_id INT REFERENCES topics(id) ON DELETE SET NULL,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    
    -- Sandbox execution settings
    init_script TEXT NOT NULL,                     -- CREATE TABLE + INSERT để khởi tạo data
    solution_query TEXT NOT NULL,                  -- Query đáp án chuẩn
    
    -- Multi-database support
    supported_databases VARCHAR(20)[] NOT NULL DEFAULT '{postgresql}',
    
    -- Comparison settings
    order_matters BOOLEAN DEFAULT FALSE,           -- Có cần so sánh thứ tự rows không?
    
    -- Metadata
    hints JSONB DEFAULT '[]',                      -- Array of hints
    sample_output JSONB,                           -- Example output để hiển thị
    
    is_public BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. SUBMISSIONS: Bài nộp practice của sinh viên
CREATE TABLE submissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    
    code TEXT NOT NULL,                            -- SQL query submitted
    database_type VARCHAR(20) NOT NULL,            -- postgresql, mysql, sqlserver
    
    status VARCHAR(20) NOT NULL,                   -- pending, running, accepted, wrong_answer, error, timeout
    execution_time_ms INT,
    
    expected_output JSONB,                         -- Expected result
    actual_output JSONB,                           -- Actual result  
    error_message TEXT,
    
    is_correct BOOLEAN DEFAULT FALSE,
    
    submitted_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. EXAMS: Kỳ thi do giảng viên tạo
CREATE TABLE exams (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    duration_minutes INT NOT NULL,                 -- Thời gian làm bài
    
    -- Database settings
    allowed_databases VARCHAR(20)[] DEFAULT '{postgresql}',
    
    -- AI & Features
    allow_ai_assistance BOOLEAN DEFAULT FALSE,     -- Cho phép AI hỗ trợ
    shuffle_problems BOOLEAN DEFAULT FALSE,
    show_result_immediately BOOLEAN DEFAULT TRUE,
    max_attempts INT DEFAULT 1,                    -- Số lần submit tối đa mỗi bài
    
    is_public BOOLEAN DEFAULT FALSE,               -- Thi thử công khai
    status VARCHAR(20) DEFAULT 'draft',            -- draft, published, ongoing, completed
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 6. EXAM_PROBLEMS: Bài tập trong kỳ thi
CREATE TABLE exam_problems (
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    points INT DEFAULT 10,                         -- Điểm cho bài này
    sort_order INT DEFAULT 0,
    UNIQUE(exam_id, problem_id)
);

-- 7. EXAM_PARTICIPANTS: Sinh viên tham gia thi
CREATE TABLE exam_participants (
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    started_at TIMESTAMPTZ,                        -- Thời điểm bắt đầu làm bài  
    submitted_at TIMESTAMPTZ,                      -- Thời điểm nộp bài
    
    total_score DECIMAL(5,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'registered',       -- registered, in_progress, submitted, graded
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(exam_id, user_id)
);

-- 8. EXAM_SUBMISSIONS: Bài nộp trong kỳ thi
CREATE TABLE exam_submissions (
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    exam_problem_id BIGINT NOT NULL REFERENCES exam_problems(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    code TEXT NOT NULL,
    database_type VARCHAR(20) NOT NULL,
    
    status VARCHAR(20) NOT NULL,                   -- pending, running, accepted, wrong_answer, error, timeout
    execution_time_ms INT,
    
    expected_output JSONB,
    actual_output JSONB,
    error_message TEXT,
    
    is_correct BOOLEAN DEFAULT FALSE,
    score DECIMAL(5,2) DEFAULT 0,
    attempt_number INT DEFAULT 1,
    
    submitted_at TIMESTAMPTZ DEFAULT NOW()
);

-- 9. USER_PROGRESS: Theo dõi tiến độ học tập
CREATE TABLE user_progress (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id BIGINT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    
    is_solved BOOLEAN DEFAULT FALSE,
    attempts INT DEFAULT 0,
    best_time_ms INT,
    
    first_attempted_at TIMESTAMPTZ,
    last_attempted_at TIMESTAMPTZ,
    solved_at TIMESTAMPTZ,
    
    UNIQUE(user_id, problem_id)
);

-- =============================================
-- INDEXES
-- =============================================

-- Users
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_email ON users(email);

-- Problems
CREATE INDEX idx_problems_topic ON problems(topic_id);
CREATE INDEX idx_problems_difficulty ON problems(difficulty);
CREATE INDEX idx_problems_public_active ON problems(is_public, is_active);

-- Submissions
CREATE INDEX idx_submissions_user ON submissions(user_id);
CREATE INDEX idx_submissions_problem ON submissions(problem_id);
CREATE INDEX idx_submissions_user_problem ON submissions(user_id, problem_id);

-- Exams
CREATE INDEX idx_exams_created_by ON exams(created_by);
CREATE INDEX idx_exams_status ON exams(status);
CREATE INDEX idx_exams_time ON exams(start_time, end_time);

-- Exam Participants
CREATE INDEX idx_exam_participants_user ON exam_participants(user_id);
CREATE INDEX idx_exam_participants_exam ON exam_participants(exam_id);

-- Exam Submissions
CREATE INDEX idx_exam_submissions_user ON exam_submissions(user_id);
CREATE INDEX idx_exam_submissions_exam ON exam_submissions(exam_id);

-- User Progress
CREATE INDEX idx_user_progress_user ON user_progress(user_id);
CREATE INDEX idx_user_progress_solved ON user_progress(user_id, is_solved);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_user_progress_solved;
DROP INDEX IF EXISTS idx_user_progress_user;
DROP INDEX IF EXISTS idx_exam_submissions_exam;
DROP INDEX IF EXISTS idx_exam_submissions_user;
DROP INDEX IF EXISTS idx_exam_participants_exam;
DROP INDEX IF EXISTS idx_exam_participants_user;
DROP INDEX IF EXISTS idx_exams_time;
DROP INDEX IF EXISTS idx_exams_status;
DROP INDEX IF EXISTS idx_exams_created_by;
DROP INDEX IF EXISTS idx_submissions_user_problem;
DROP INDEX IF EXISTS idx_submissions_problem;
DROP INDEX IF EXISTS idx_submissions_user;
DROP INDEX IF EXISTS idx_problems_public_active;
DROP INDEX IF EXISTS idx_problems_difficulty;
DROP INDEX IF EXISTS idx_problems_topic;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_role;

DROP TABLE IF EXISTS user_progress;
DROP TABLE IF EXISTS exam_submissions;
DROP TABLE IF EXISTS exam_participants;
DROP TABLE IF EXISTS exam_problems;
DROP TABLE IF EXISTS exams;
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS problems;
DROP TABLE IF EXISTS topics;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
