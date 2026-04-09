# Phase 4: Exam Execution - Implementation Plan

## Overview
Phase 4 enables students to participate in exams by joining, viewing problems, and submitting code/answers in real-time.

## Key Features to Implement

### 1. Student Exam Participation
- **Join Exam**: Student registers for exam (creates exam_participants record)
- **Start Exam**: Student marks exam as started (sets started_at timestamp)
- **Exam Status**: Track exam progress (registered → in_progress → submitted → graded)
- **Time Validation**: Check if exam is within start/end time and duration

### 2. Student Views Exam
- **Get Exam Details**: List all problems for the exam
- **View Problem**: Get single problem details with hints/description
- **View Submissions**: See student's previous submissions for a problem
- **Time Remaining**: Calculate time remaining based on exam duration

### 3. Student Submits Code
- **Submit Code**: Store code submission, validate syntax, execute code
- **Real-time Execution**: Run SQL query against database
- **Capture Output**: Store actual output and execution result
- **Scoring**: Auto-score based on scoring mode (if set to auto)
- **Attempt Tracking**: Track attempt number per problem

### 4. Submission Status Tracking
- Submission statuses: `pending`, `running`, `accepted`, `wrong_answer`, `error`, `timeout`
- Store actual vs expected output
- Store error messages if any
- Record execution time

## Database Schema (Already Exists)

### exam_participants
```
id, exam_id, user_id, started_at, submitted_at, 
total_score, status (registered/in_progress/submitted/graded), created_at
```

### exam_submissions
```
id, exam_id, exam_problem_id, user_id, code, database_type,
status, execution_time_ms, expected_output, actual_output, 
error_message, is_correct, score, attempt_number, submitted_at
```

### exam_problems
```
id, exam_id, problem_id, points, sort_order, 
scoring_mode (auto/answer_key/manual), reference_answer
```

### exams
```
id, title, description, created_by, start_time, end_time, 
duration_minutes, status
```

## New Components Needed

### DTOs (internals/student/controller/dto/)
1. **JoinExamRequest**: exam_id
2. **JoinExamResponse**: participant_id, exam details, total_problems
3. **StartExamRequest**: exam_id
4. **StartExamResponse**: started_at, time_remaining
5. **GetExamRequest**: exam_id, user_id
6. **GetExamResponse**: exam details, list of problems, time_remaining, status
7. **GetProblemRequest**: exam_id, exam_problem_id
8. **GetProblemResponse**: problem details, student's submissions, attempt_number
9. **SubmitCodeRequest**: exam_id, exam_problem_id, code
10. **SubmitCodeResponse**: submission_id, status, output, score (if auto-graded)
11. **SubmitExamRequest**: exam_id
12. **SubmitExamResponse**: total_score, submitted_at

### Usecase (internals/student/usecase/)
1. **IStudentExamUseCase interface** with methods:
   - JoinExam(ctx, examID, userID) → JoinExamResponse
   - StartExam(ctx, examID, userID) → StartExamResponse
   - GetExam(ctx, examID, userID) → GetExamResponse
   - GetProblem(ctx, examID, examProblemID, userID) → GetProblemResponse
   - SubmitCode(ctx, examID, examProblemID, userID, code) → SubmitCodeResponse
   - SubmitExam(ctx, examID, userID) → SubmitExamResponse
   - GetTimeRemaining(ctx, examID, userID) → time_remaining

2. **Implementation**: StudentExamUseCase struct

### HTTP Handlers (internals/student/controller/http/)
1. **JoinExamHandler**: POST /student/exams/:examId/join
2. **StartExamHandler**: POST /student/exams/:examId/start
3. **GetExamHandler**: GET /student/exams/:examId
4. **GetProblemHandler**: GET /student/exams/:examId/problems/:problemId
5. **SubmitCodeHandler**: POST /student/exams/:examId/problems/:problemId/submit
6. **SubmitExamHandler**: POST /student/exams/:examId/submit
7. **GetTimeRemainingHandler**: GET /student/exams/:examId/time-remaining

### Routes (internals/student/controller/http/)
Register all endpoints under `/student/exams`

### SQLC Queries (sql/queries/exam.sql)
Add queries for:
- Get exam with problems (for student view)
- Get student's exam submission history
- Create exam submission
- Update exam submission with actual_output, status, execution_time
- Get time remaining for exam

## Implementation Steps

1. Create student module directory structure
2. Write SQLC queries
3. Create DTOs
4. Implement StudentExamUseCase
5. Create HTTP handlers
6. Register routes in server
7. Test each endpoint
8. Test full workflow (join → start → submit → view → grading)

## Testing Strategy

- Test join exam (valid exam, already joined, etc.)
- Test start exam (time validation)
- Test get exam (permissions, time remaining)
- Test submit code (execution, scoring)
- Test edge cases (exam timeout, multiple attempts, etc.)

## Notes
- Use existing auth middleware to get userID from context
- Validate user can only see their own submissions
- Enforce exam time windows (start_time, end_time, duration_minutes)
- Real SQL execution happens via database runner (separate service)
- Auto-scoring uses Phase 3 scoring module
