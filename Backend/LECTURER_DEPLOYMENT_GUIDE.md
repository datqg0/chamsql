# Lecturer Module - Deployment & Testing Guide

## Deployment Steps

### 1. Database Migration

```bash
# Connect to PostgreSQL and run the migration
cd Backend
migrate -path sql/schema -database "postgresql://user:password@localhost:5432/chamsql?sslmode=disable" up

# Or using Go migrate programmatically
# The migration runs on application startup if using auto-migration
```

**Migration Files:**
- `sql/schema/006_create_classes.sql` - Creates classes, class_members, class_exams tables

### 2. Generate SQLC Models

```bash
cd Backend
sqlc generate
```

This generates:
- `sql/models/class.sql.go` - All query methods and parameter structs

### 3. Build Application

```bash
cd Backend
go build ./cmd/app
```

**Expected Output:** `app` (or `app.exe` on Windows) executable

### 4. Run Application

```bash
./app
```

**Server starts on configured port (default: 8080)**

Access routes:
- Health check: `GET http://localhost:8080/health`
- Lecturer endpoints: `GET http://localhost:8080/api/v1/lecturer/classes` (requires auth)

## Testing Guide

### Prerequisites

1. Running PostgreSQL database
2. Application running on port 8080
3. Valid JWT token (from login endpoint)

### 1. Get Authentication Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "lecturer@example.com",
    "password": "your_password"
  }'
```

Response:
```json
{
  "token": "eyJhbGc...",
  "refreshToken": "eyJhbGc...",
  "user": {
    "id": 5,
    "email": "lecturer@example.com",
    "role": "lecturer"
  }
}
```

Save the `token` value for subsequent requests.

### 2. Class Management Tests

#### 2.1 Create Class

```bash
TOKEN="your_token_here"

curl -X POST http://localhost:8080/api/v1/lecturer/classes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Advanced SQL",
    "description": "Learn complex SQL queries",
    "semester": "Spring 2024",
    "year": 2024
  }'
```

**Expected:** 201 status with ClassResponse containing auto-generated code

#### 2.2 Get Class

```bash
curl -X GET http://localhost:8080/api/v1/lecturer/classes/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** 200 status with class details

#### 2.3 List Classes

```bash
curl -X GET "http://localhost:8080/api/v1/lecturer/classes?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** 200 status with paginated classes list

**Test pagination:**
```bash
# Test page 2
curl -X GET "http://localhost:8080/api/v1/lecturer/classes?page=2&pageSize=5" \
  -H "Authorization: Bearer $TOKEN"
```

#### 2.4 Update Class

```bash
curl -X PUT http://localhost:8080/api/v1/lecturer/classes/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Advanced SQL - Updated",
    "description": "Updated description"
  }'
```

**Expected:** 200 status with updated class

**Edge case - Partial update:**
```bash
# Only update name
curl -X PUT http://localhost:8080/api/v1/lecturer/classes/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Name"
  }'
```

#### 2.5 Delete Class

```bash
curl -X DELETE http://localhost:8080/api/v1/lecturer/classes/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** 204 status (No Content)

**Error case - Not owner:**
```bash
# Try to delete another lecturer's class
# Create class with lecturer1, try to delete with lecturer2
# Expected: 403 Forbidden
```

### 3. Class Member Tests

#### 3.1 Add Member to Class

First, get a student user ID (assume 42):

```bash
curl -X POST http://localhost:8080/api/v1/lecturer/classes/1/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 42,
    "role": "member"
  }'
```

**Expected:** 201 status with member details

#### 3.2 Add Teaching Assistant

```bash
curl -X POST http://localhost:8080/api/v1/lecturer/classes/1/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 43,
    "role": "ta"
  }'
```

**Expected:** 201 status with TA member

#### 3.3 List Class Members

```bash
curl -X GET "http://localhost:8080/api/v1/lecturer/classes/1/members?page=1&pageSize=20" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** 200 status with members list

#### 3.4 Remove Member

```bash
curl -X DELETE http://localhost:8080/api/v1/lecturer/classes/1/members/42 \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** 204 status (No Content)

### 4. Class Exam Tests

#### 4.1 Assign Exam to Class

First, create an exam (assume exam ID is 7):

```bash
curl -X POST http://localhost:8080/api/v1/lecturer/classes/1/exams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "examId": 7
  }'
```

**Expected:** 201 status (Created)

#### 4.2 List Class Exams

```bash
curl -X GET http://localhost:8080/api/v1/lecturer/classes/1/exams \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** 200 status with list of exams assigned to class

#### 4.3 Remove Exam from Class

```bash
curl -X DELETE http://localhost:8080/api/v1/lecturer/classes/1/exams/7 \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** 204 status (No Content)

### 5. Error Handling Tests

#### 5.1 Authorization Errors

**Missing token:**
```bash
curl -X GET http://localhost:8080/api/v1/lecturer/classes
```

**Expected:** 401 Unauthorized

**Invalid token:**
```bash
curl -X GET http://localhost:8080/api/v1/lecturer/classes \
  -H "Authorization: Bearer invalid_token"
```

**Expected:** 401 Unauthorized

#### 5.2 Validation Errors

**Invalid class ID:**
```bash
curl -X GET http://localhost:8080/api/v1/lecturer/classes/invalid \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** 400 Bad Request

**Missing required field:**
```bash
curl -X POST http://localhost:8080/api/v1/lecturer/classes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Missing name"
  }'
```

**Expected:** 400 Bad Request

#### 5.3 Not Found Errors

**Non-existent class:**
```bash
curl -X GET http://localhost:8080/api/v1/lecturer/classes/99999 \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** 404 Not Found

#### 5.4 Conflict Errors

**Duplicate member:**
```bash
# Add same member twice
curl -X POST http://localhost:8080/api/v1/lecturer/classes/1/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"userId": 42, "role": "member"}'

# Then try again
curl -X POST http://localhost:8080/api/v1/lecturer/classes/1/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"userId": 42, "role": "member"}'
```

**Expected:** 409 Conflict on second attempt

**Duplicate exam assignment:**
```bash
# Assign exam twice
curl -X POST http://localhost:8080/api/v1/lecturer/classes/1/exams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"examId": 7}'

# Then try again - should get 409
```

**Expected:** 409 Conflict on second attempt

### 6. Permission & Ownership Tests

#### 6.1 Delete Non-Owned Class

```bash
# Lecturer A creates class with ID 1
# Lecturer B tries to delete it

curl -X DELETE http://localhost:8080/api/v1/lecturer/classes/1 \
  -H "Authorization: Bearer lecturer_b_token"
```

**Expected:** 403 Forbidden with "only class creator can delete"

### 7. Pagination Tests

#### 7.1 First Page

```bash
curl -X GET "http://localhost:8080/api/v1/lecturer/classes?page=1&pageSize=5" \
  -H "Authorization: Bearer $TOKEN"
```

Response has 5 items with page=1

#### 7.2 Out of Range Page

```bash
curl -X GET "http://localhost:8080/api/v1/lecturer/classes?page=999&pageSize=10" \
  -H "Authorization: Bearer $TOKEN"
```

Response with page=999 but empty classes list

#### 7.3 Invalid PageSize

```bash
# pageSize > 100 should be capped
curl -X GET "http://localhost:8080/api/v1/lecturer/classes?page=1&pageSize=1000" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** Uses default or max pageSize

## Automated Integration Tests

Create `tests/lecturer_test.go`:

```go
package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestLecturerClassWorkflow(t *testing.T) {
	client := &http.Client{}
	baseURL := "http://localhost:8080/api/v1"
	token := getTestToken(t) // Helper to get JWT token

	// 1. Create class
	createReq := map[string]interface{}{
		"name":      "Test Class",
		"semester":  "Fall 2024",
	}
	classResp := createClass(t, client, baseURL, token, createReq)
	classID := classResp["id"]

	// 2. Add members
	addMemberReq := map[string]interface{}{
		"userId": 42,
		"role":   "member",
	}
	addClassMember(t, client, baseURL, token, classID, addMemberReq)

	// 3. Assign exam
	assignExamReq := map[string]interface{}{
		"examId": 7,
	}
	assignExamToClass(t, client, baseURL, token, classID, assignExamReq)

	// 4. List exams
	exams := listClassExams(t, client, baseURL, token, classID)
	if len(exams) != 1 {
		t.Fatalf("Expected 1 exam, got %d", len(exams))
	}

	// 5. Delete class
	deleteClass(t, client, baseURL, token, classID)
}

// Helper functions
func createClass(t *testing.T, client *http.Client, baseURL, token string, req map[string]interface{}) map[string]interface{} {
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", baseURL+"/lecturer/classes", bytes.NewBuffer(body))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201, got %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}
```

## Database Testing

### Check Schema

```sql
-- Verify classes table exists
SELECT * FROM information_schema.tables WHERE table_name = 'classes';

-- Verify unique constraint on class code
SELECT constraint_name FROM information_schema.table_constraints 
WHERE table_name = 'classes' AND constraint_type = 'UNIQUE';

-- Check class_members unique constraint
SELECT constraint_name FROM information_schema.table_constraints 
WHERE table_name = 'class_members' AND constraint_type = 'UNIQUE';
```

### Sample Data Insertion

```sql
-- Insert test lecturer
INSERT INTO users (email, username, password_hash, full_name, role, is_active, created_at, updated_at)
VALUES ('lecturer@test.com', 'lecturer1', 'hash', 'Test Lecturer', 'lecturer', true, NOW(), NOW());

-- Insert test students
INSERT INTO users (email, username, password_hash, full_name, role, is_active, created_at, updated_at)
VALUES 
  ('student1@test.com', 'student1', 'hash', 'Student One', 'student', true, NOW(), NOW()),
  ('student2@test.com', 'student2', 'hash', 'Student Two', 'student', true, NOW(), NOW());

-- Create test class
INSERT INTO classes (name, code, created_by, semester, year, is_active, created_at, updated_at)
VALUES ('Test Class', 'CLASS-TEST01', 1, 'Fall 2024', 2024, true, NOW(), NOW());

-- Add members
INSERT INTO class_members (class_id, user_id, role, joined_at)
VALUES 
  (1, 2, 'member', NOW()),
  (1, 3, 'member', NOW());
```

## Performance Testing

### Load Test Class Creation

```bash
# Using Apache Bench
ab -n 100 -c 10 -H "Authorization: Bearer $TOKEN" \
  -p payload.json -T "application/json" \
  http://localhost:8080/api/v1/lecturer/classes

# payload.json:
# {"name": "Load Test Class", "semester": "Fall 2024"}
```

### Check Database Indices

```sql
-- Verify indices for performance
SELECT * FROM pg_indexes WHERE tablename IN ('classes', 'class_members', 'class_exams');

-- Expected indices:
-- - classes.created_by (for lecturer queries)
-- - class_members.class_id, class_members.user_id
-- - class_exams.class_id, class_exams.exam_id
```

## Troubleshooting

### Issue: 401 Unauthorized on all endpoints

**Cause:** Invalid or expired JWT token

**Solution:**
1. Get new token: `POST /api/v1/auth/login`
2. Verify token has `user_id` claim
3. Check token expiration

### Issue: 404 Not Found for class

**Cause:** Class doesn't exist or wrong ID

**Solution:**
1. Verify class ID: `GET /api/v1/lecturer/classes`
2. Check database: `SELECT * FROM classes WHERE id = ?`

### Issue: 409 Conflict when adding member

**Cause:** User already in class

**Solution:**
1. List members: `GET /api/v1/lecturer/classes/{id}/members`
2. Remove duplicate: `DELETE /api/v1/lecturer/classes/{id}/members/{userId}`
3. Re-add with correct role

### Issue: Database migration not running

**Cause:** Migration tool not installed or wrong path

**Solution:**
```bash
# Install migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migration
migrate -path sql/schema -database "postgresql://user:pass@localhost/db" up
```

## Next Steps

1. **Integration Tests:** Add comprehensive test suite in `tests/` directory
2. **Performance Tests:** Load test with 1000+ concurrent users
3. **Security Tests:** SQL injection, authorization bypass attempts
4. **Documentation:** Generate OpenAPI/Swagger documentation
5. **Monitoring:** Add metrics/logging for production deployment

## Monitoring & Logging

### Log Levels

Application logs at different levels:
- `ERROR` - Database errors, validation failures
- `WARN` - Permission denials, invalid operations
- `INFO` - Successful operations, state changes
- `DEBUG` - SQL query details (if enabled)

### Recommended Monitoring

1. **API Response Times:** Track p50, p95, p99 for each endpoint
2. **Database Queries:** Monitor slow queries (> 100ms)
3. **Authentication:** Track failed login attempts
4. **Resource Usage:** CPU, memory, database connections
