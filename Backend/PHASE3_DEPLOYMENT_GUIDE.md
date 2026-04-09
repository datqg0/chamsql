# Phase 3: Scoring System Deployment & Testing Guide

## Deployment Checklist

### Pre-Deployment Verification
- [x] Code compiles without errors
- [x] All new files in correct directories
- [x] SQLC queries properly formatted
- [x] Routes registered correctly
- [x] Handlers properly implemented
- [x] DTOs include validation tags

### Build & Compilation
```bash
# Navigate to backend directory
cd Backend

# Build the application
go build ./cmd/app

# Expected: No compilation errors
```

### Database Preparation
No new migrations required - database schema already includes:
- `exam_submissions.score` (DECIMAL)
- `exam_submissions.graded_by` (BIGINT)
- `exam_submissions.graded_at` (TIMESTAMPTZ)
- `exam_problems.scoring_mode` (VARCHAR)
- `exam_problems.reference_answer` (TEXT)

If running fresh database:
```bash
# Run all migrations
migrate -path sql/schema -database "postgresql://user:password@localhost:5432/chamsql" up
```

### Environment Setup
No new environment variables needed. Existing setup should work.

## Testing Guide

### 1. Unit Testing (Scoring Package)

#### Test Auto-Scorer
```go
// Create auto scorer
scorer := scoring.NewAutoScorer()

// Test matching outputs
request := &scoring.GradingRequest{
    ScoringMode:      scoring.ScoringModeAuto,
    ActualOutput:     []byte(`[{"id": 1, "name": "John"}]`),
    ExpectedOutput:   []byte(`[{"id": 1, "name": "John"}]`),
    MaxPoints:        10,
    SubmissionStatus: "accepted",
}

result, err := scorer.CalculateScore(request)
// Expected: Score = 10, IsCorrect = true
```

#### Test Answer-Key Scorer
```go
scorer := scoring.NewAnswerKeyScorer(nil) // Uses FlexibleAnswerComparer

request := &scoring.GradingRequest{
    ScoringMode:      scoring.ScoringModeAnswerKey,
    StudentAnswer:    stringPtr("SELECT * FROM users WHERE id = 1"),
    ReferenceAnswer:  stringPtr("select * from users where id = 1"),
    MaxPoints:        5,
    SubmissionStatus: "pending",
}

result, err := scorer.CalculateScore(request)
// Expected: Score = 5, IsCorrect = true (due to normalization)
```

### 2. API Testing

#### Setup Test Data
```sql
-- Verify exam exists
SELECT id, title FROM exams WHERE id = 1;

-- Verify submissions exist
SELECT id, exam_id, user_id, status FROM exam_submissions WHERE exam_id = 1;

-- Verify scoring mode is set
SELECT id, scoring_mode, reference_answer FROM exam_problems WHERE exam_id = 1;
```

#### Test 1: View Submission for Grading
```bash
curl -X GET http://localhost:8080/lecturer/submissions/1 \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json"

# Expected Response (200 OK):
{
  "submissionId": 1,
  "examId": 1,
  "problemId": 1,
  "problemTitle": "Simple SELECT Query",
  "studentId": 100,
  "studentName": "John Doe",
  "studentEmail": "john@example.com",
  "code": "SELECT * FROM users WHERE id = 1",
  "status": "accepted",
  "scoringMode": "answer_key",
  "score": 0,
  "maxPoints": 10,
  "isCorrect": false,
  "actualOutput": [...],
  "expectedOutput": [...],
  "referenceAnswer": "SELECT * FROM users WHERE id = 1",
  "submittedAt": "2024-04-10T10:30:00Z"
}
```

#### Test 2: Grade a Submission
```bash
curl -X POST http://localhost:8080/lecturer/submissions/1/grade \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "score": 8.5,
    "feedback": "Good approach but consider edge cases",
    "isCorrect": true
  }'

# Expected Response (200 OK):
{
  "submissionId": 1,
  "studentId": 100,
  "studentName": "John Doe",
  "problemTitle": "Simple SELECT Query",
  "score": 8.5,
  "maxPoints": 10,
  "isCorrect": true,
  "scoringMode": "answer_key",
  "gradedBy": 2,
  "gradedByName": "Dr. Smith",
  "gradedAt": "2024-04-10T15:45:00Z",
  "feedback": "Good approach but consider edge cases",
  "submittedAt": "2024-04-10T10:30:00Z"
}
```

#### Test 3: List Ungraded Submissions
```bash
curl -X GET http://localhost:8080/lecturer/exams/1/ungraded \
  -H "Authorization: Bearer <TOKEN>"

# Expected Response (200 OK):
{
  "submissions": [
    {
      "submissionId": 2,
      "studentId": 101,
      "studentName": "Jane Smith",
      "problemTitle": "JOIN Query",
      "score": 0,
      "maxPoints": 10,
      "isCorrect": false,
      "scoringMode": "auto",
      "submittedAt": "2024-04-10T11:15:00Z"
    }
  ],
  "total": 1,
  "examId": 1,
  "ungradedCount": 1,
  "gradedCount": 2
}
```

#### Test 4: Get Grading Statistics
```bash
curl -X GET http://localhost:8080/lecturer/exams/1/grading-stats \
  -H "Authorization: Bearer <TOKEN>"

# Expected Response (200 OK):
{
  "examId": 1,
  "totalSubmissions": 3,
  "gradedCount": 2,
  "ungradedCount": 1,
  "gradingPercentage": 66.67,
  "averageScore": 8.25,
  "maxScore": 9.0,
  "minScore": 7.5
}
```

#### Test 5: Bulk Grade Submissions
```bash
curl -X POST http://localhost:8080/lecturer/submissions/bulk-grade \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "submissions": [
      {
        "submissionId": 2,
        "score": 7.5,
        "feedback": "Missing index usage optimization"
      },
      {
        "submissionId": 3,
        "score": 9.0,
        "feedback": "Excellent query optimization"
      }
    ]
  }'

# Expected Response (200 OK):
{
  "processedCount": 2,
  "failedCount": 0,
  "results": [
    {
      "submissionId": 2,
      "studentId": 101,
      "studentName": "Jane Smith",
      "score": 7.5,
      "maxPoints": 10,
      "isCorrect": false,
      "gradedAt": "2024-04-10T15:50:00Z"
    },
    {
      "submissionId": 3,
      "studentId": 102,
      "studentName": "Bob Johnson",
      "score": 9.0,
      "maxPoints": 10,
      "isCorrect": true,
      "gradedAt": "2024-04-10T15:50:00Z"
    }
  ],
  "errors": []
}
```

### 3. Error Case Testing

#### Test Missing Authorization
```bash
curl -X GET http://localhost:8080/lecturer/submissions/1
# Expected: 401 Unauthorized
```

#### Test Invalid Submission ID
```bash
curl -X GET http://localhost:8080/lecturer/submissions/invalid \
  -H "Authorization: Bearer <TOKEN>"
# Expected: 400 Bad Request - "invalid submission id"
```

#### Test Submission Not Found
```bash
curl -X GET http://localhost:8080/lecturer/submissions/99999 \
  -H "Authorization: Bearer <TOKEN>"
# Expected: 404 Not Found
```

#### Test Invalid Score (Negative)
```bash
curl -X POST http://localhost:8080/lecturer/submissions/1/grade \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"score": -1}'
# Expected: 400 Bad Request - "score cannot be negative"
```

#### Test Missing Required Fields in Bulk Grade
```bash
curl -X POST http://localhost:8080/lecturer/submissions/bulk-grade \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"submissions": []}'
# Expected: 400 Bad Request - "bulk grading request cannot be empty"
```

### 4. Database Verification

After grading, verify database updates:

```sql
-- Check if score was updated
SELECT id, score, graded_by, graded_at FROM exam_submissions WHERE id = 1;

-- Expected: score = 8.5, graded_by = 2 (lecturer ID), graded_at = timestamp

-- Verify grading stats
SELECT 
  COUNT(*) as total,
  COUNT(CASE WHEN graded_by IS NOT NULL THEN 1 END) as graded,
  AVG(score) as avg_score
FROM exam_submissions 
WHERE exam_id = 1;
```

### 5. Performance Testing

#### Load Test - Multiple Submissions
```bash
# Grade 100 submissions
for i in {1..100}; do
  curl -X POST http://localhost:8080/lecturer/submissions/$i/grade \
    -H "Authorization: Bearer <TOKEN>" \
    -H "Content-Type: application/json" \
    -d "{\"score\": 8.5}"
done
```

#### Response Time Benchmark
- Single submission grade: < 500ms
- List ungraded (10 items): < 300ms
- Get stats: < 200ms
- Bulk grade (50 items): < 5s

### 6. Integration Testing

#### Test Scoring Modes End-to-End

**Auto-Scoring Mode**:
1. Create exam with problem in auto-scoring mode
2. Submit test case with outputs
3. Call AutoScoreSubmission to calculate score
4. Verify score matches output comparison

**Answer-Key Mode**:
1. Create exam with problem in answer-key mode
2. Set reference_answer in exam_problems
3. Grade submission with answer comparison
4. Verify score based on answer match

**Manual Mode**:
1. Create exam with problem in manual mode
2. Submit answer
3. Lecturer manually grades
4. Verify score and graded_by/graded_at set correctly

## Troubleshooting

### Issue: SQLC queries not found
**Solution**: Run `sqlc generate` after modifying `sql/queries/exam.sql`

### Issue: Compilation error "undefined ScoreCalculator"
**Solution**: Ensure `pkgs/scoring/types.go` is in correct location and package

### Issue: Routes not accessible
**Solution**: Verify routes.go initializes both classUseCase and gradingUseCase

### Issue: Database connection errors
**Solution**: Verify database connection string in environment and database is running

### Issue: Authorization failures on grading endpoints
**Solution**: Ensure auth middleware is applied to /lecturer group and token is valid

## Rollback Plan

If issues occur:

1. **Code Rollback**:
   ```bash
   git revert HEAD  # Revert Phase 3 changes
   go build ./cmd/app
   ```

2. **Database Rollback**:
   - No migrations were added, so database is unaffected
   - Existing data in exam_submissions table is safe

3. **Known Safe State**:
   - Phase 1 (Admin RBAC) - ✅ Fully functional
   - Phase 2 (Lecturer Foundation) - ✅ Fully functional
   - All previous endpoints continue to work

## Post-Deployment Verification

1. **Health Check**:
   ```bash
   curl http://localhost:8080/health
   # Expected: 200 OK
   ```

2. **Grading Endpoint Check**:
   ```bash
   # Should return 200 or 401 (not 404)
   curl http://localhost:8080/lecturer/submissions/1/grade
   ```

3. **Database Check**:
   ```bash
   # Verify new queries work
   psql -d chamsql -c "SELECT * FROM exam_submissions LIMIT 1;"
   ```

4. **Log Verification**:
   ```bash
   # Check application logs for errors
   tail -f app.log | grep -i "error\|grading"
   ```

## Performance Optimization Tips

1. **For large submissions**:
   - Use pagination in ListUngradedSubmissions
   - Consider async grading for bulk operations

2. **For SQL output comparison**:
   - Add indexes on exam_submissions.status, graded_by
   - Cache scoring mode lookups

3. **For statistics**:
   - Add database index on graded_by and graded_at
   - Consider materialized views for stats

## Security Considerations

1. **Authorization**:
   - All grading endpoints require auth
   - Lecturer can only grade their own exams (TODO: implement with PermissionService)

2. **Input Validation**:
   - Score validation (0 to max_points)
   - Request size limits
   - SQL injection prevention (using parameterized queries)

3. **Data Protection**:
   - Sensitive student data only shown to lecturer
   - Audit log tracking (TODO: enhance with detailed logging)

## Next Steps

1. **Immediate** (within 1 week):
   - Run comprehensive test suite
   - Integration testing with actual data
   - Performance profiling

2. **Short-term** (1-2 weeks):
   - Integrate PermissionService for ownership verification
   - Add detailed error logging and monitoring
   - Create student notification system (when graded)

3. **Medium-term** (2-4 weeks):
   - Implement rubric/criteria-based scoring
   - Add scoring templates for reuse
   - Create analytics dashboard

## Support & Documentation

- **Implementation Details**: See PHASE3_SCORING_IMPLEMENTATION.md
- **API Documentation**: Swagger comments in handler.go
- **Code Examples**: See this file's API Testing section
- **Questions**: Refer to inline code comments and docstrings
