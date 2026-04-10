# ChamsQL Backend - Testing Guide

## Setup Instructions

### 1. Start Docker (Database only)

```bash
docker-compose up -d
```

This will start only the PostgreSQL database (`sqlexam.postgres`) on port 5432.

### 2. Initialize Database & Load Test Data

Make sure `psql` (PostgreSQL client) is installed on your system.

**Windows (PowerShell):**
```powershell
.\scripts\migrate.ps1
```

**Linux/Mac (Bash):**
```bash
bash scripts/migrate.sh
```

### 3. Run the Application

```bash
go run ./cmd/app/main.go
```

The API will start on `http://localhost:8080`

## Test Users

After seeding, you'll have these users available:

### Admin
- Email: `admin@chamsql.com`
- Password: `password` (hashed with bcrypt)
- Username: `admin`

### Lecturer
- Email: `lecturer@chamsql.com`
- Password: `password`
- Username: `lecturer`

### Students
- Student 1: `student1@chamsql.com`
- Student 2: `student2@chamsql.com`
- Student 3: `student3@chamsql.com`

All have password: `password`

## Test Scenarios

### Phase 1: Admin RBAC System
- [ ] Verify users have correct roles assigned
- [ ] Verify role permissions are set correctly
- [ ] Check permission audit logs

### Phase 2: Lecturer System
- [ ] Lecturer creates a class (DB101)
- [ ] Students join the class
- [ ] Lecturer creates an exam
- [ ] Exam is linked to class

### Phase 3: Scoring System
- [ ] Verify scoring modes: auto, answer_key, manual
- [ ] Check exam_problems have correct scoring_mode assigned

### Phase 4: Exam Execution (MAIN TEST)
1. **Student joins exam**
   - POST /api/exams/:id/join

2. **Student starts exam**
   - POST /api/exams/:id/start

3. **Student views exam**
   - GET /api/exams/:id

4. **Student gets problem**
   - GET /api/exams/:id/problems/:problem_id
   - Should return: init_script, solution_query

5. **Student submits SQL code**
   - POST /api/exams/:id/problems/:problem_id/submit
   - Code is executed in sandbox
   - Output compared with solution
   - Score calculated

6. **Code Execution Details** (Executor Service)
   - Runs init script first
   - Runs student code
   - Compares output with solution query
   - Auto-grades based on matching
   - Returns score and status

7. **Submit exam**
   - POST /api/exams/:id/submit

### Phase 5: Results & Reporting
1. **Get exam results**
   - GET /api/exams/results with pagination & filtering
   - Filters: status, score range, date range

2. **Get class ranking**
   - GET /api/exams/:id/ranking
   - Sorted by total_score DESC

## API Test Examples

### 1. Login & Get Token

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "student1@chamsql.com",
    "password": "password"
  }'
```

Response includes `access_token` and `refresh_token`.

### 2. Start Exam

```bash
curl -X POST http://localhost:8080/api/exams/1/start \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

### 3. Get Problem

```bash
curl http://localhost:8080/api/exams/1/problems/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Returns problem with:
- title, description
- init_script (CREATE TABLE + INSERT)
- solution_query (expected output)

### 4. Submit Code

```bash
curl -X POST http://localhost:8080/api/exams/1/problems/1/submit \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "SELECT * FROM users WHERE id = 1;",
    "database_type": "postgresql"
  }'
```

Response includes:
- status (graded/error/pending_review)
- score (calculated based on output matching)
- actual_output
- execution_time_ms

### 5. Get Exam Results

```bash
curl "http://localhost:8080/api/exams/results?page=1&limit=10&status=graded" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Query Parameters:
- page: pagination
- limit: results per page
- status: graded, pending_review, error
- score_min, score_max: score range filter
- start_date, end_date: date range filter

## Database Schema

Key tables:
- **users**: admin, lecturer, students
- **roles**: admin, lecturer, student
- **permissions**: resource_type + action combinations
- **classes**: created by lecturer, students join
- **exams**: created by lecturer, linked to class
- **exam_problems**: problems in exam with scoring_mode
- **exam_participants**: students taking exam
- **exam_submissions**: individual code submissions

## Executor Service Details

Located in: `internals/student/usecase/executor.go`

### Execution Flow
1. Parse code by semicolon-separated statements
2. Execute in transaction (sandbox):
   - Run init script (CREATE TABLE + INSERT)
   - Run student code
   - Run solution query
3. Compare rows:
   - Type coercion for matching
   - Order-independent by default
4. Score calculation:
   - Matches / Total rows * 100
5. Auto-grading:
   - auto/answer_key modes: immediate grade
   - manual mode: status = pending_review
6. Error handling:
   - Timeout: 5 seconds
   - SQL parse errors captured
   - Execution errors captured

## Troubleshooting

### Database Connection Error
- Ensure PostgreSQL is running: `docker-compose up -d`
- Check DATABASE_URI in .env
- Verify database exists: `sql_exam_db`

### Migration Failed
- Check database is empty or migrations are idempotent
- Review schema files in `sql/schema/`
- Check PostgreSQL logs

### Test Data Not Loaded
- Run migrate script again
- Verify all schema files execute without errors
- Check seed file: `sql/schema/007_seed_test_data.sql`

### Code Submission Failed
- Check executor service logs
- Verify problem has valid init_script and solution_query
- Check SQL syntax in student code

## Next Steps After Testing

1. Fix any issues found during testing
2. Commit changes to git
3. Push to remote repository
4. Deploy to staging/production

---

**Status**: Ready for testing. All phases implemented and built successfully.
