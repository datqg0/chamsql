-- =============================================
-- TEST DATA SEED SCRIPT
-- =============================================

-- 1. Create test users (Admin, Lecturer, Students)
INSERT INTO users (email, username, password_hash, full_name, role) VALUES
    -- Admin
    ('admin@chamsql.com', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/LFm', 'Admin User', 'admin'),
    
    -- Lecturer
    ('lecturer@chamsql.com', 'lecturer', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/LFm', 'Prof. Nguyễn Văn A', 'lecturer'),
    
    -- Students
    ('student1@chamsql.com', 'student1', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/LFm', 'Trần Văn B', 'student'),
    ('student2@chamsql.com', 'student2', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/LFm', 'Lê Thị C', 'student'),
    ('student3@chamsql.com', 'student3', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/LFm', 'Phạm Thanh D', 'student')
ON CONFLICT (email) DO NOTHING;

-- 2. Create topics
INSERT INTO topics (name, slug, description, icon) VALUES
    ('SELECT', 'select', 'Basic SELECT queries', '📄'),
    ('WHERE & Filters', 'where-filters', 'WHERE clause and filtering', '🔍'),
    ('JOIN', 'join', 'INNER, LEFT, RIGHT, FULL OUTER joins', '🔗'),
    ('Aggregate Functions', 'aggregate', 'GROUP BY, COUNT, SUM, AVG, MAX, MIN', '📊'),
    ('Subqueries', 'subquery', 'Nested queries and IN/EXISTS clauses', '📦'),
    ('Window Functions', 'window', 'ROW_NUMBER, RANK, PARTITION BY', '🪟')
ON CONFLICT (slug) DO NOTHING;

-- 3. Create problems
INSERT INTO problems (title, slug, description, difficulty, topic_id, created_by, init_script, solution_query, supported_databases) VALUES
    (
        'Simple SELECT Query',
        'simple-select',
        'Write a SELECT query to retrieve all columns from the users table.',
        'easy',
        1,
        2, -- lecturer
        'CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100), email VARCHAR(100)); INSERT INTO users VALUES (1, ''John Doe'', ''john@example.com''), (2, ''Jane Smith'', ''jane@example.com'');',
        'SELECT * FROM users;',
        '{postgresql}'
    ),
    (
        'Simple WHERE Filter',
        'simple-where',
        'Select all users with id = 1.',
        'easy',
        2,
        2,
        'CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100), email VARCHAR(100)); INSERT INTO users VALUES (1, ''John Doe'', ''john@example.com''), (2, ''Jane Smith'', ''jane@example.com'');',
        'SELECT * FROM users WHERE id = 1;',
        '{postgresql}'
    ),
    (
        'INNER JOIN Example',
        'inner-join',
        'Join orders and customers tables.',
        'medium',
        3,
        2,
        'CREATE TABLE customers (id INT PRIMARY KEY, name VARCHAR(100)); CREATE TABLE orders (id INT PRIMARY KEY, customer_id INT, amount DECIMAL(10,2)); INSERT INTO customers VALUES (1, ''Customer A''), (2, ''Customer B''); INSERT INTO orders VALUES (1, 1, 100.00), (2, 1, 200.00), (3, 2, 150.00);',
        'SELECT c.name, o.amount FROM customers c INNER JOIN orders o ON c.id = o.customer_id;',
        '{postgresql}'
    ),
    (
        'COUNT and GROUP BY',
        'count-group-by',
        'Count orders per customer.',
        'medium',
        4,
        2,
        'CREATE TABLE customers (id INT PRIMARY KEY, name VARCHAR(100)); CREATE TABLE orders (id INT PRIMARY KEY, customer_id INT, amount DECIMAL(10,2)); INSERT INTO customers VALUES (1, ''Customer A''), (2, ''Customer B''); INSERT INTO orders VALUES (1, 1, 100.00), (2, 1, 200.00), (3, 2, 150.00);',
        'SELECT c.id, c.name, COUNT(o.id) as order_count FROM customers c LEFT JOIN orders o ON c.id = o.customer_id GROUP BY c.id, c.name;',
        '{postgresql}'
    )
ON CONFLICT (slug) DO NOTHING;

-- 4. Create a class
INSERT INTO classes (name, description, code, created_by, semester, year) VALUES
    ('Database Basics - Class 1', 'Introduction to SQL and Database Design', 'DB101', 2, 'Fall', 2024)
ON CONFLICT (code) DO NOTHING;

-- 5. Add students to class
INSERT INTO class_members (class_id, user_id, role) 
SELECT c.id, u.id, 'member'
FROM classes c, users u
WHERE c.code = 'DB101' AND u.role = 'student'
ON CONFLICT (class_id, user_id) DO NOTHING;

-- 6. Create an exam
INSERT INTO exams (title, description, created_by, start_time, end_time, duration_minutes, status)
VALUES (
    'Midterm Exam - SQL Fundamentals',
    'Test your knowledge of basic SQL queries',
    2, -- lecturer
    NOW() + INTERVAL '1 day',
    NOW() + INTERVAL '1 day 2 hours',
    120,
    'published'
)
ON CONFLICT DO NOTHING;

-- 7. Link exam problems
INSERT INTO exam_problems (exam_id, problem_id, points, sort_order, scoring_mode)
SELECT e.id, p.id, 10, ROW_NUMBER() OVER (ORDER BY p.id), 'auto'
FROM exams e, problems p
WHERE e.title = 'Midterm Exam - SQL Fundamentals'
LIMIT 4
ON CONFLICT DO NOTHING;

-- 8. Link exam to class
INSERT INTO class_exams (class_id, exam_id)
SELECT c.id, e.id
FROM classes c, exams e
WHERE c.code = 'DB101' AND e.title = 'Midterm Exam - SQL Fundamentals'
ON CONFLICT (class_id, exam_id) DO NOTHING;

-- 9. Assign roles to users
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE (u.email = 'admin@chamsql.com' AND r.name = 'admin')
   OR (u.email = 'lecturer@chamsql.com' AND r.name = 'lecturer')
   OR (u.email IN ('student1@chamsql.com', 'student2@chamsql.com', 'student3@chamsql.com') AND r.name = 'student')
ON CONFLICT (user_id, role_id) DO NOTHING;

-- Done!
SELECT 'Seed data loaded successfully!' as message;
