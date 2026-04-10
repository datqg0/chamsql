# Code Execution Service Implementation

## Overview

Implemented a complete **Code Executor Service** that:
- ✅ Executes student SQL submissions in a **sandbox transaction**
- ✅ Compares output against expected results (if provided)
- ✅ Supports automatic scoring (auto/answer_key modes)
- ✅ Handles errors gracefully and returns detailed feedback
- ✅ Measures execution time and performance
- ✅ Supports init scripts and solution queries

## Architecture

### File: `internals/student/usecase/executor.go`

**Two key interfaces:**

1. **CodeExecutor** - Main interface for code execution
   ```go
   ExecuteCode(ctx context.Context, code, initScript, solutionQuery string, 
               databaseType string, timeout time.Duration) (*ExecutionResult, error)
   ```

2. **ExecutionResult** - Returns detailed execution data
   ```go
   Success       bool                          // Did execution succeed?
   Output        []map[string]interface{}      // Student query results
   ExpectedOutput []map[string]interface{}     // Expected results
   IsCorrect     bool                          // Output matches expected?
   Score         float64                       // Calculated score (0-100)
   ErrorMessage  string                        // Error if failed
   ExecutionTime int32                         // Execution time in ms
   ```

## Implementation Details

### 1. Code Execution Flow

```
1. Parse SQL into individual statements (;-delimited)
2. Start sandbox transaction (auto-rollback after execution)
3. Execute init script in transaction
4. Execute student code in transaction
5. Execute solution query for comparison (in same transaction)
6. Compare outputs row-by-row
7. Calculate score based on comparison
8. Return detailed ExecutionResult
```

### 2. Transaction Isolation

**Why sandbox transaction?**
- ✅ All changes rollback after execution
- ✅ No pollution of database
- ✅ Safe execution of untrusted code
- ✅ Can safely execute INSERT/UPDATE/DELETE

### 3. Query Classification

**Modifying queries:** INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, TRUNCATE
- Executed with `tx.Exec()` (no result scanning)

**Read-only queries:** SELECT
- Executed with `tx.Query()` (results collected)
- Results converted to `map[string]interface{}`

### 4. Output Comparison

Compares row-by-row with type coercion:
- `int32`, `int64`, `float64` treated as same type family
- String comparison is exact match
- Boolean comparison is exact
- Null values handled properly

### 5. SQL Parsing

**Splits by semicolon:**
```sql
CREATE TABLE temp (id INT);
INSERT INTO temp VALUES (1);
SELECT * FROM temp;
```

Becomes 3 separate statements, executed in order.

## Integration with SubmitCode

### Updated `SubmitCode` method:

**Before:** Submissions marked "pending", no execution
**After:** Full execution with auto-grading

```go
executionResult, err := su.executor.ExecuteCode(ctx, 
    req.Code, 
    problem.InitScript, 
    problem.SolutionQuery, 
    req.DatabaseType, 
    5*time.Second)  // 5 second timeout
```

### Status Mapping

| Scenario | Status | Score |
|----------|--------|-------|
| Execution fails | error | 0.0 |
| Auto-mode, output matches | graded | 100.0 |
| Auto-mode, output differs | graded | 0.0 |
| Manual mode (pending review) | pending_review | 0.0 |

### Scoring Modes

1. **auto**: Auto-grade using solution query
   - Compare outputs immediately
   - Return pass/fail result

2. **answer_key**: Same as auto, uses reference answer
   - Alias for automatic comparison

3. **manual**: Manual grading by lecturer
   - Execute code, store result
   - Mark as pending_review
   - Lecturer grades later

## Features

### ✅ Error Handling
- Init script errors caught and reported
- SQL syntax errors caught and reported
- Connection/transaction errors handled
- Timeout after 5 seconds (configurable)

### ✅ Performance Metrics
- Execution time measured in milliseconds
- Stored in submission record
- Useful for analysis

### ✅ Type Flexibility
- pgtype.Numeric converted to float64
- PostgreSQL types handled seamlessly
- Type coercion for comparison

### ✅ Security
- Transactions auto-rollback
- No persistent changes
- Timeout prevents infinite loops
- Can execute on sandbox database

## API Changes

### POST `/student/exams/:examID/problems/:problemID/submit`

**Response now includes:**
```json
{
  "submission_id": 123,
  "exam_id": 1,
  "exam_problem_id": 45,
  "status": "graded",         // Instead of "pending"
  "score": 100.0,             // Actual score
  "is_correct": true,         // Comparison result
  "attempt_number": 1,
  "execution_time_ms": 245,   // New: execution time
  "error_message": null,      // If error occurred
  "submitted_at": "2024-04-10T...",
  "scoring_mode": "auto"
}
```

## Usage Examples

### Example 1: Simple SELECT query

**Problem setup:**
- InitScript: `CREATE TABLE numbers (id INT, value INT);`
  `INSERT INTO numbers VALUES (1, 10), (2, 20), (3, 30);`
- SolutionQuery: `SELECT SUM(value) FROM numbers;`

**Student submits:**
```sql
SELECT SUM(value) FROM numbers;
```

**Result:** IsCorrect=true, Score=100.0, Status=graded ✅

### Example 2: Incorrect query

**Same problem setup**

**Student submits:**
```sql
SELECT COUNT(*) FROM numbers;
```

**Result:** IsCorrect=false, Score=0.0, Status=graded ❌

### Example 3: SQL Error

**Student submits:**
```sql
SELECT * FORM numbers;  -- Typo: FORM instead of FROM
```

**Result:** 
```
Success: false
Status: error
ErrorMessage: "query error: syntax error at or near 'FORM'"
```

## Database Integration

### Updated `exam_submissions` table fields:
- `score` - Calculated score
- `is_correct` - Boolean result
- `status` - execution status (error, graded, pending_review, completed)
- `error_message` - Error details
- `execution_time_ms` - Execution time

## Performance Considerations

### Timeout: 5 seconds
- Prevents infinite loops
- Prevents resource exhaustion
- Can be tuned per problem

### Transaction Scope
- Single transaction for init + code + solution
- All in same session context
- Ensures data consistency

### Query Limits
- No built-in result set limit
- Consider adding LIMIT to solution queries
- Prevent OOM from massive result sets

## Testing Considerations

### Unit Test Ideas
1. Simple SELECT with matching output
2. SELECT with different output (should fail)
3. SQL syntax error handling
4. Init script that sets up temp data
5. INSERT/UPDATE/DELETE operations
6. Timeout (infinite loop query)
7. Null value comparison
8. Type coercion (int vs float)

### Integration Test Ideas
1. Full exam submission flow
2. Multiple submissions for same problem
3. Different scoring modes
4. Comparison with reference answers
5. Performance metrics collection

## Future Enhancements

### Possible improvements:
1. **Partial scoring:** Award points for partial correct output
2. **Query plan analysis:** Suggest optimizations
3. **Code quality metrics:** Check for efficiency
4. **Resource limits:** Memory, CPU per query
5. **Output formatting:** More flexible comparisons
6. **Statistics:** Track common student errors
7. **Plagiarism detection:** Compare code submissions

## Build Status

✅ Application compiles successfully
✅ No type errors
✅ Ready for integration testing
✅ Code execution integrated with submission flow

## Key Statistics

**New files:**
- `internals/student/usecase/executor.go` - 308 lines

**Modified files:**
- `internals/student/usecase/exam.go` - Updated SubmitCode method
- `internals/student/usecase/results.go` - Enhanced with pagination
- `internals/student/controller/dto/results.go` - Added filtering DTOs
- `internals/student/controller/http/handler.go` - Added pagination handling

**Total changes:**
- 15+ new methods/functions
- 8+ helper functions
- Comprehensive error handling
- Full transaction support
