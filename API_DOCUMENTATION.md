# ChamSQL API Documentation

## PDF Upload & Problem Extraction (Lecturer)

### Upload PDF
```http
POST /api/v1/lecturer/pdf/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <PDF file>
```

**Response:**
```json
{
  "id": 123,
  "status": "uploading",
  "file_name": "exam_problems.pdf",
  "created_at": "2024-01-15T10:00:00Z",
  "message": "Upload successful. Processing extraction..."
}
```

### Get Upload Status
```http
GET /api/v1/lecturer/pdf/{uploadId}/status
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": 123,
  "status": "completed",
  "file_name": "exam_problems.pdf",
  "extraction_result": {
    "total_problems": 5,
    "problems": [...]
  },
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:30Z"
}
```

### Get Extracted Problems
```http
GET /api/v1/lecturer/pdf/{uploadId}/problems
Authorization: Bearer <token>
```

**Response:**
```json
{
  "upload_id": 123,
  "problems": [
    {
      "id": 1,
      "problem_number": 1,
      "title": "Select all employees",
      "description": "Write a query to select all employees from the employees table...",
      "difficulty": "easy",
      "status": "pending",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "note": "PDF contains problem descriptions only. Solution queries must be added manually."
}
```

### Update Problem Solution
```http
PUT /api/v1/lecturer/pdf/problems/{problemId}/solution
Authorization: Bearer <token>
Content-Type: application/json

{
  "solution_query": "SELECT * FROM employees",
  "db_type": "postgresql"
}
```

**Response:**
```json
{
  "id": 1,
  "solution_query": "SELECT * FROM employees",
  "db_type": "postgresql",
  "status": "confirmed",
  "message": "Solution updated and problem confirmed successfully"
}
```

## Sandbox Management (Admin/Lecturer)

### Get Sandbox Status
```http
GET /api/v1/admin/sandbox/status
Authorization: Bearer <token>
```

**Response:**
```json
{
  "postgres": {
    "connected": true,
    "uri": "postgresql://sandbox_...",
    "error": ""
  },
  "mysql": {
    "connected": true,
    "uri": "mysql://sandbox_...",
    "error": ""
  },
  "sqlserver": {
    "connected": true,
    "uri": "sqlserver://sandbox_...",
    "error": ""
  }
}
```

### Test Sandbox Query
```http
POST /api/v1/admin/sandbox/test
Authorization: Bearer <token>
Content-Type: application/json

{
  "db_type": "postgresql",
  "query": "SELECT * FROM employees LIMIT 10"
}
```

**Response:**
```json
{
  "db_type": "postgresql",
  "query": "SELECT * FROM employees LIMIT 10",
  "success": true,
  "result": {
    "columns": ["id", "name", "department"],
    "rows": [[1, "John", "IT"], [2, "Jane", "HR"]],
    "row_count": 2,
    "execution_ms": 15
  },
  "execution_ms": 15
}
```

## Grading (Lecturer)

### List Ungraded Submissions
```http
GET /api/v1/lecturer/exams/{examId}/ungraded
Authorization: Bearer <token>
```

### Get Grading Stats
```http
GET /api/v1/lecturer/exams/{examId}/grading-stats
Authorization: Bearer <token>
```

**Response:**
```json
{
  "totalSubmissions": 50,
  "pendingCount": 10,
  "gradedCount": 35,
  "errorCount": 5,
  "averageScore": 7.5
}
```

### View Submission
```http
GET /api/v1/lecturer/submissions/{submissionId}
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": 123,
  "studentID": 456,
  "studentCode": "SV001",
  "studentName": "Nguyen Van A",
  "examID": 789,
  "examTitle": "Midterm Exam",
  "problemID": 101,
  "problemTitle": "Select all employees",
  "submittedCode": "SELECT * FROM employees",
  "status": "pending",
  "score": null,
  "maxScore": 10,
  "feedback": null,
  "executionTimeMs": 25,
  "submittedAt": "2024-01-15T10:30:00Z"
}
```

### Grade Submission
```http
POST /api/v1/lecturer/submissions/{submissionId}/grade
Authorization: Bearer <token>
Content-Type: application/json

{
  "score": 8.5,
  "feedback": "Good query, but could optimize with WHERE clause"
}
```

### Auto-Grade Submission
```http
POST /api/v1/lecturer/submissions/{submissionId}/auto-grade
Authorization: Bearer <token>
```

### Bulk Grade
```http
POST /api/v1/lecturer/submissions/bulk-grade
Authorization: Bearer <token>
Content-Type: application/json

{
  "submissionIDs": [1, 2, 3],
  "score": 5.0,
  "feedback": "Partial credit"
}
```

## Student Exam Flow

### List My Exams
```http
GET /api/v1/my-exams
Authorization: Bearer <token>
```

### Start Exam
```http
POST /api/v1/exams/{examId}/start
Authorization: Bearer <token>
```

### Submit Answer
```http
POST /api/v1/exams/{examId}/submit
Authorization: Bearer <token>
Content-Type: application/json

{
  "answers": [
    {
      "exam_problem_id": 1,
      "code": "SELECT * FROM employees",
      "database_type": "postgresql"
    }
  ]
}
```

### Finish Exam
```http
POST /api/v1/exams/{examId}/finish
Authorization: Bearer <token>
```

## Exam Management (Lecturer)

### Create Exam
```http
POST /api/v1/exams
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "SQL Midterm Exam",
  "description": "Test SQL knowledge",
  "start_time": "2024-01-20T09:00:00Z",
  "end_time": "2024-01-20T11:00:00Z",
  "duration_minutes": 120,
  "max_attempts": 1,
  "is_public": false,
  "allowed_databases": ["postgresql", "mysql"]
}
```

### Add Problem to Exam
```http
POST /api/v1/exams/{examId}/problems
Authorization: Bearer <token>
Content-Type: application/json

{
  "problem_id": 1,
  "points": 10,
  "sort_order": 1
}
```

### Add Participants
```http
POST /api/v1/exams/{examId}/participants
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_ids": [1, 2, 3]
}
```
