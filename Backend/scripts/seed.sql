-- ChamsQL Seed Data - Demo NCKH
-- Password for all users: 'password123'
-- Bcrypt hash: '$2a$10$8K1p/a0dxv.T5A6W8K1p/O5v.T5A6W8K1p/a0dxv.T5A6W8K1p/'

-- 1. Create Users
INSERT INTO users (username, email, password_hash, role, is_active) VALUES
  ('admin', 'admin@chamsql.edu.vn', '$2a$10$8K1p/a0dxv.T5A6W8K1p/O5v.T5A6W8K1p/a0dxv.T5A6W8K1p/', 'admin', true),
  ('gv_nguyen', 'gv.nguyen@chamsql.edu.vn', '$2a$10$8K1p/a0dxv.T5A6W8K1p/O5v.T5A6W8K1p/a0dxv.T5A6W8K1p/', 'lecturer', true),
  ('sv_001', 'sv001@student.edu.vn', '$2a$10$8K1p/a0dxv.T5A6W8K1p/O5v.T5A6W8K1p/a0dxv.T5A6W8K1p/', 'student', true),
  ('sv_002', 'sv002@student.edu.vn', '$2a$10$8K1p/a0dxv.T5A6W8K1p/O5v.T5A6W8K1p/a0dxv.T5A6W8K1p/', 'student', true),
  ('sv_003', 'sv003@student.edu.vn', '$2a$10$8K1p/a0dxv.T5A6W8K1p/O5v.T5A6W8K1p/a0dxv.T5A6W8K1p/', 'student', true);

-- 2. Create Topics
INSERT INTO topics (name, slug, description) VALUES
  ('SQL Cơ Bản', 'sql-co-ban', 'Các câu lệnh SELECT, WHERE, ORDER BY cơ bản'),
  ('JOIN & Relationships', 'join-relationships', 'Truy vấn trên nhiều bảng dữ liệu'),
  ('Aggregate Functions', 'aggregate', 'Sử dụng COUNT, SUM, AVG, GROUP BY'),
  ('Subqueries', 'subqueries', 'Truy vấn con và logic phức tạp');

-- 3. Create Problems
INSERT INTO problems (title, slug, description, difficulty, init_script, solution_query, topic_id, created_by, is_public, supported_databases) VALUES
  ('Liệt kê tất cả nhân viên',
   'liet-ke-nhan-vien',
   '## Yêu cầu\nViết câu SQL để liệt kê tất cả nhân viên trong bảng `employees`, sắp xếp theo tên tăng dần.',
   'easy',
   'CREATE TABLE employees (id SERIAL PRIMARY KEY, name VARCHAR(100), dept VARCHAR(50), salary NUMERIC); INSERT INTO employees VALUES (1,''Alice'',''IT'',75000),(2,''Bob'',''HR'',60000),(3,''Charlie'',''IT'',80000);',
   'SELECT * FROM employees ORDER BY name ASC;',
   1, 2, true, '{"postgresql", "mysql", "sqlserver"}'),

  ('Tính lương trung bình theo phòng ban',
   'luong-trung-binh-phong-ban',
   '## Yêu cầu\nViết câu SQL tính lương trung bình của từng phòng ban. Kết quả gồm: `dept`, `avg_salary` (làm tròn 2 chữ số).',
   'medium',
   'CREATE TABLE employees (id SERIAL PRIMARY KEY, name VARCHAR(100), dept VARCHAR(50), salary NUMERIC); INSERT INTO employees VALUES (1,''Alice'',''IT'',75000),(2,''Bob'',''HR'',60000),(3,''Charlie'',''IT'',80000),(4,''Diana'',''HR'',65000);',
   'SELECT dept, ROUND(AVG(salary), 2) as avg_salary FROM employees GROUP BY dept ORDER BY avg_salary DESC;',
   3, 2, true, '{"postgresql", "mysql", "sqlserver"}'),

  ('Sinh viên chưa có điểm',
   'sinh-vien-chua-co-diem',
   '## Yêu cầu\nTìm tất cả sinh viên chưa có điểm trong bảng `grades`.',
   'medium',
   'CREATE TABLE students (id SERIAL PRIMARY KEY, name VARCHAR(100)); CREATE TABLE grades (student_id INT, subject VARCHAR(50), score NUMERIC); INSERT INTO students VALUES (1,''An''),(2,''Binh''),(3,''Cuong''); INSERT INTO grades VALUES (1,''Toan'',8.5),(1,''Van'',7.0);',
   'SELECT s.* FROM students s LEFT JOIN grades g ON s.id = g.student_id WHERE g.student_id IS NULL;',
   2, 2, true, '{"postgresql", "mysql", "sqlserver"}');

-- 4. Create Exam
INSERT INTO exams (title, description, start_time, end_time, duration_minutes, created_by, status, is_public, max_attempts) VALUES
  ('Bài Thi SQL Cơ Bản - Demo NCKH',
   'Kỳ thi thử nghiệm cho nghiên cứu khoa học. Gồm 3 câu hỏi SQL từ cơ bản đến trung bình.',
   CURRENT_TIMESTAMP - INTERVAL '1 hour',
   CURRENT_TIMESTAMP + INTERVAL '24 hours',
   60,
   2,
   'ongoing',
   true,
   3);

-- 5. Assign Problems to Exam
INSERT INTO exam_problems (exam_id, problem_id, points, sort_order) VALUES
  (1, 1, 30, 1),
  (1, 2, 35, 2),
  (1, 3, 35, 3);
