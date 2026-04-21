BEGIN;

-- Topic
INSERT INTO topics (name, slug, description, sort_order, is_active)
VALUES ('SQL Fundamentals', 'sql-fundamentals', 'Core SQL exercises for grading demo', 1, TRUE)
ON CONFLICT (slug) DO NOTHING;

-- Users (password hash is a placeholder bcrypt)
INSERT INTO users (email, username, password_hash, full_name, role, student_id, is_active)
VALUES
    ('lecturer.demo@chamsql.local', 'lecturer_demo', '$2a$10$7EqJtq98hPqEX7fNZaFWoO.HdQ4/6QdM7gc3KJ2YMBX1IzBgE4f6K', 'Lecturer Demo', 'lecturer', NULL, TRUE),
    ('student.demo1@chamsql.local', 'student_demo_1', '$2a$10$7EqJtq98hPqEX7fNZaFWoO.HdQ4/6QdM7gc3KJ2YMBX1IzBgE4f6K', 'Student Demo 1', 'student', 'SV0001', TRUE),
    ('student.demo2@chamsql.local', 'student_demo_2', '$2a$10$7EqJtq98hPqEX7fNZaFWoO.HdQ4/6QdM7gc3KJ2YMBX1IzBgE4f6K', 'Student Demo 2', 'student', 'SV0002', TRUE)
ON CONFLICT (email) DO NOTHING;

-- Problems
INSERT INTO problems (
    title,
    slug,
    description,
    difficulty,
    topic_id,
    created_by,
    init_script,
    solution_query,
    supported_databases,
    order_matters,
    is_public,
    is_active
)
SELECT
    'Find Active Users',
    'find-active-users-demo',
    'Return all active users from users_demo ordered by id.',
    'easy',
    t.id,
    u.id,
    'CREATE TABLE users_demo (id INT, name VARCHAR(50), is_active BOOLEAN); INSERT INTO users_demo VALUES (1, ''An'', TRUE), (2, ''Binh'', FALSE), (3, ''Chi'', TRUE);',
    'SELECT id, name FROM users_demo WHERE is_active = TRUE ORDER BY id;',
    ARRAY['postgresql', 'mysql', 'sqlserver'],
    TRUE,
    TRUE,
    TRUE
FROM topics t
JOIN users u ON u.email = 'lecturer.demo@chamsql.local'
WHERE t.slug = 'sql-fundamentals'
  AND NOT EXISTS (SELECT 1 FROM problems p WHERE p.slug = 'find-active-users-demo');

INSERT INTO problems (
    title,
    slug,
    description,
    difficulty,
    topic_id,
    created_by,
    init_script,
    solution_query,
    supported_databases,
    order_matters,
    is_public,
    is_active
)
SELECT
    'Top Salary By Department',
    'top-salary-dept-demo',
    'Return each department and max salary.',
    'medium',
    t.id,
    u.id,
    'CREATE TABLE salaries (emp_id INT, department VARCHAR(50), salary INT); INSERT INTO salaries VALUES (1, ''IT'', 1000), (2, ''IT'', 1500), (3, ''HR'', 1200), (4, ''HR'', 900);',
    'SELECT department, MAX(salary) AS max_salary FROM salaries GROUP BY department ORDER BY department;',
    ARRAY['postgresql', 'mysql', 'sqlserver'],
    TRUE,
    TRUE,
    TRUE
FROM topics t
JOIN users u ON u.email = 'lecturer.demo@chamsql.local'
WHERE t.slug = 'sql-fundamentals'
  AND NOT EXISTS (SELECT 1 FROM problems p WHERE p.slug = 'top-salary-dept-demo');

-- Exam
INSERT INTO exams (
    title,
    description,
    created_by,
    start_time,
    end_time,
    duration_minutes,
    allowed_databases,
    allow_ai_assistance,
    shuffle_problems,
    show_result_immediately,
    max_attempts,
    is_public,
    status
)
SELECT
    'Demo Grading Exam',
    'Seeded exam for grading workflow',
    u.id,
    NOW() - INTERVAL '1 hour',
    NOW() + INTERVAL '2 days',
    60,
    ARRAY['postgresql', 'mysql', 'sqlserver'],
    FALSE,
    FALSE,
    TRUE,
    3,
    FALSE,
    'published'
FROM users u
WHERE u.email = 'lecturer.demo@chamsql.local'
  AND NOT EXISTS (
      SELECT 1 FROM exams e
      WHERE e.title = 'Demo Grading Exam'
  );

-- Link exam problems
INSERT INTO exam_problems (exam_id, problem_id, points, sort_order)
SELECT e.id, p.id, 10, 1
FROM exams e
JOIN problems p ON p.slug = 'find-active-users-demo'
WHERE e.title = 'Demo Grading Exam'
ON CONFLICT (exam_id, problem_id) DO NOTHING;

INSERT INTO exam_problems (exam_id, problem_id, points, sort_order)
SELECT e.id, p.id, 15, 2
FROM exams e
JOIN problems p ON p.slug = 'top-salary-dept-demo'
WHERE e.title = 'Demo Grading Exam'
ON CONFLICT (exam_id, problem_id) DO NOTHING;

-- Participants
INSERT INTO exam_participants (exam_id, user_id, status)
SELECT e.id, u.id, 'in_progress'
FROM exams e
JOIN users u ON u.email IN ('student.demo1@chamsql.local', 'student.demo2@chamsql.local')
WHERE e.title = 'Demo Grading Exam'
ON CONFLICT (exam_id, user_id) DO NOTHING;

-- Submissions for grading dashboard demo
INSERT INTO exam_submissions (
    exam_id,
    exam_problem_id,
    user_id,
    code,
    database_type,
    status,
    execution_time_ms,
    expected_output,
    actual_output,
    error_message,
    is_correct,
    score,
    attempt_number,
    submitted_at
)
SELECT
    e.id,
    ep.id,
    u.id,
    'SELECT id, name FROM users_demo WHERE is_active = TRUE ORDER BY id;',
    'postgresql',
    'accepted',
    24,
    '[{"id":1,"name":"An"},{"id":3,"name":"Chi"}]'::jsonb,
    '[{"id":1,"name":"An"},{"id":3,"name":"Chi"}]'::jsonb,
    NULL,
    TRUE,
    10,
    1,
    NOW() - INTERVAL '15 minutes'
FROM exams e
JOIN exam_problems ep ON ep.exam_id = e.id
JOIN problems p ON p.id = ep.problem_id
JOIN users u ON u.email = 'student.demo1@chamsql.local'
WHERE e.title = 'Demo Grading Exam'
  AND p.slug = 'find-active-users-demo'
  AND NOT EXISTS (
      SELECT 1 FROM exam_submissions es
      WHERE es.exam_id = e.id AND es.exam_problem_id = ep.id AND es.user_id = u.id
  );

INSERT INTO exam_submissions (
    exam_id,
    exam_problem_id,
    user_id,
    code,
    database_type,
    status,
    execution_time_ms,
    expected_output,
    actual_output,
    error_message,
    is_correct,
    score,
    attempt_number,
    submitted_at
)
SELECT
    e.id,
    ep.id,
    u.id,
    'SELECT department, MAX(salary) AS max_salary FROM salaries GROUP BY department ORDER BY department;',
    'postgresql',
    'running',
    NULL,
    NULL,
    NULL,
    NULL,
    FALSE,
    0,
    1,
    NOW() - INTERVAL '10 minutes'
FROM exams e
JOIN exam_problems ep ON ep.exam_id = e.id
JOIN problems p ON p.id = ep.problem_id
JOIN users u ON u.email = 'student.demo2@chamsql.local'
WHERE e.title = 'Demo Grading Exam'
  AND p.slug = 'top-salary-dept-demo'
  AND NOT EXISTS (
      SELECT 1 FROM exam_submissions es
      WHERE es.exam_id = e.id AND es.exam_problem_id = ep.id AND es.user_id = u.id
  );

COMMIT;
