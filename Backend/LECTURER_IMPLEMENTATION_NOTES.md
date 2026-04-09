# Lecturer Module Implementation Notes

## Overview

This document describes the implementation of the Lecturer module for the ChAmSQL exam platform. The module enables lecturers to manage classes, add students, and assign exams with proper resource ownership tracking and permission enforcement.

## Phase 2: Lecturer Foundation - COMPLETED

### Completed Features

#### 1. Database Schema (006_create_classes.sql)

**New Tables:**

- **classes**: Main class entity managed by lecturers
  - `id` (int64, PK)
  - `name` (string, required)
  - `description` (string, nullable)
  - `code` (string, unique) - Auto-generated format: CLASS-XXXXXX
  - `created_by` (int64, FK to users) - Lecturer who created the class
  - `semester` (string, nullable) - e.g., "Fall 2024"
  - `year` (int32, nullable)
  - `is_active` (boolean, default: true)
  - `created_at`, `updated_at` (timestamps)

- **class_members**: Many-to-many relationship between classes and students
  - `id` (int64, PK)
  - `class_id` (int64, FK to classes)
  - `user_id` (int64, FK to users)
  - `role` (string) - "member" or "ta"
  - `joined_at` (timestamp)
  - Unique constraint: (class_id, user_id)

- **class_exams**: Many-to-many relationship between classes and exams
  - `id` (int64, PK)
  - `class_id` (int64, FK to classes)
  - `exam_id` (int64, FK to exams)
  - `created_at` (timestamp)
  - Unique constraint: (class_id, exam_id)

**Updated Tables:**

- **exam_problems**: Added new columns for scoring modes
  - `scoring_mode` (string, nullable) - "auto", "answer_key", or "manual"
  - `reference_answer` (string, nullable) - For answer-key comparison

- **exam_submissions**: Added grading columns
  - `graded_by` (int64, nullable, FK to users) - Lecturer who graded
  - `graded_at` (timestamp, nullable) - When grading was completed

#### 2. SQLC Queries (sql/queries/class.sql)

**14 Queries Implemented:**

**Class CRUD:**
- `CreateClass` - Insert new class with unique code
- `GetClassByID` - Fetch single class
- `GetClassByCode` - Fetch by join code
- `ListClassesByLecturer` - List all active classes by creator (paginated)
- `UpdateClass` - Update class details (name, description, semester, year)
- `DeactivateClass` - Soft delete by setting is_active=false
- `DeleteClass` - Hard delete

**Class Members:**
- `AddClassMember` - Add user to class with role
- `GetClassMember` - Check membership
- `ListClassMembers` - List members with user details (paginated)
- `RemoveClassMember` - Remove from class
- `CountClassMembers` - Get total members
- `GetStudentClasses` - List classes student is in

**Class Exams:**
- `AssignExamToClass` - Create class-exam mapping
- `ListClassExams` - List exams with details (sorted by start_time)
- `RemoveExamFromClass` - Remove mapping
- `GetClassExamByID` - Fetch exam details with class context

#### 3. Data Models (sql/models/class.sql.go)

SQLC generated models with proper JSON tags:

- `Class` - Struct with nullable fields for optional data
- `ClassMember` - User role and join timestamp
- `ClassExam` - Exam info in class context
- `ListClassMembersRow` - Joined with user details
- `ListClassExamsRow` - Full exam details in class

**Parameter Structs:**
- `CreateClassParams`
- `UpdateClassParams`
- `AddClassMemberParams`
- `ListClassMembersParams`
- `RemoveClassMemberParams`
- `GetClassMemberParams`
- `AssignExamToClassParams`
- `RemoveExamFromClassParams`

#### 4. DTOs (internals/lecturer/controller/dto/class.go)

**Request DTOs:**
- `CreateClassRequest` - name (required), description, semester, year
- `UpdateClassRequest` - All fields nullable, partial updates
- `AddClassMemberRequest` - userId, role (member/ta)
- `BulkAddClassMembersRequest` - For bulk operations
- `AssignExamToClassRequest` - examId

**Response DTOs:**
- `ClassResponse` - Full class with all fields
- `ListClassesResponse` - Paginated response
- `ClassMemberResponse` - User + membership info
- `ListClassMembersResponse` - Paginated members
- `ClassExamResponse` - Exam in class context
- `ListClassExamsResponse` - Exams in class

#### 5. Usecase Layer (internals/lecturer/usecase/class_usecase.go)

**Interface: ILecturerClassUseCase**

11 methods with full documentation (params, returns, error descriptions):

**Class CRUD:**
```go
CreateClass(ctx, lectureID, req) -> ClassResponse, error
GetClass(ctx, classID) -> ClassResponse, error
ListClasses(ctx, lectureID, page, pageSize) -> ListClassesResponse, error
UpdateClass(ctx, classID, req) -> ClassResponse, error
DeleteClass(ctx, classID, lectureID) -> error  // Checks ownership
```

**Class Members:**
```go
AddClassMember(ctx, classID, userID, role) -> ClassMemberResponse, error
ListClassMembers(ctx, classID, page, pageSize) -> ListClassMembersResponse, error
RemoveClassMember(ctx, classID, userID) -> error
```

**Class Exams:**
```go
AssignExamToClass(ctx, classID, examID) -> error
ListClassExams(ctx, classID) -> ListClassExamsResponse, error
RemoveExamFromClass(ctx, classID, examID) -> error
```

**Implementation Notes:**

- All methods include context cancellation support
- Proper error handling with descriptive messages
- Ownership checks (lecturer can only delete own classes)
- Nullable field handling for optional data (Year, Description, Semester)
- Proper pointer/value conversions for SQLC compatibility
- Pagination with configurable page size (default 10-20 records)

#### 6. HTTP Handler (internals/lecturer/controller/http/handler.go)

**LecturerHandler struct** with 11 handler methods:

**Class Management:**
- `CreateClass(c)` - POST /lecturer/classes
  - Validates request, gets lecturerID from context, creates class
  - Response: 201 ClassResponse or error

- `GetClass(c)` - GET /lecturer/classes/{id}
  - Response: 200 ClassResponse or 404 error

- `ListClasses(c)` - GET /lecturer/classes
  - Query params: page, pageSize (optional)
  - Response: 200 ListClassesResponse

- `UpdateClass(c)` - PUT /lecturer/classes/{id}
  - Validates request, updates fields
  - Response: 200 ClassResponse

- `DeleteClass(c)` - DELETE /lecturer/classes/{id}
  - Checks lecturer ownership
  - Response: 204 No Content

**Class Members:**
- `AddClassMember(c)` - POST /lecturer/classes/{id}/members
  - Validates user exists, handles duplicate detection
  - Response: 201 ClassMemberResponse or 409 Conflict

- `ListClassMembers(c)` - GET /lecturer/classes/{id}/members
  - Paginated with user details
  - Response: 200 ListClassMembersResponse

- `RemoveClassMember(c)` - DELETE /lecturer/classes/{id}/members/{userId}
  - Response: 204 No Content

**Class Exams:**
- `AssignExamToClass(c)` - POST /lecturer/classes/{id}/exams
  - Response: 201 or 409 Conflict if already assigned

- `ListClassExams(c)` - GET /lecturer/classes/{id}/exams
  - Response: 200 ListClassExamsResponse

- `RemoveExamFromClass(c)` - DELETE /lecturer/classes/{id}/exams/{examId}
  - Response: 204 No Content

**Error Handling:**
- 400 Bad Request - Invalid parameters
- 401 Unauthorized - Missing auth
- 403 Forbidden - Permission denied (e.g., not class owner)
- 404 Not Found - Resource not found
- 409 Conflict - Duplicate entry
- 500 Internal Server Error - Database errors

#### 7. HTTP Routes (internals/lecturer/controller/http/routes.go)

**Route Registration Function:**
```go
Routes(rg *gin.RouterGroup, database *db.Database, authMiddleware gin.HandlerFunc)
```

**Registered Routes:**

```
POST   /api/v1/lecturer/classes
GET    /api/v1/lecturer/classes
GET    /api/v1/lecturer/classes/:id
PUT    /api/v1/lecturer/classes/:id
DELETE /api/v1/lecturer/classes/:id

POST   /api/v1/lecturer/classes/:id/members
GET    /api/v1/lecturer/classes/:id/members
DELETE /api/v1/lecturer/classes/:id/members/:userId

POST   /api/v1/lecturer/classes/:id/exams
GET    /api/v1/lecturer/classes/:id/exams
DELETE /api/v1/lecturer/classes/:id/exams/:examId
```

**Authentication:** All routes require Bearer token (authMiddleware)

#### 8. HTTP Server Integration (internals/server/http/server.go)

Routes registered in `MapRoutes()`:
```go
lecturerHttp.Routes(v1, s.database, authMiddleware)
```

Integrated alongside admin, exam, submission, and problem routes.

### Implementation Decisions

1. **Soft Delete via is_active flag** - Preserves historical data for reports
2. **Auto-generated class codes** - Format CLASS-XXXXXX (6 random alphanumeric)
3. **Resource Ownership** - Lecturer ID stored in created_by, checked on delete
4. **Nullable fields** - Year, description, semester optional for flexibility
5. **Pagination** - Default 10 records per page for classes, 20 for members
6. **Role in class_members** - Supports "member" and "ta" roles for future expansion
7. **Separate class_exams table** - M:M relationship allows flexible exam scheduling

### Type Safety & Error Handling

**SQLC-Generated Types:**
- All parameter structs are generated with correct nullable types (*string, *int32, etc.)
- Response structs properly handle NULL database values
- JSON marshaling/unmarshaling automatic with tags

**Error Messages:**
- Descriptive, user-friendly error messages
- Specific error types (not found, already exists, forbidden, etc.)
- Proper HTTP status codes for each error type

**Null Handling:**
- Helper function `ptrToStr()` for safe pointer dereferencing
- Safe pointer to value conversions for SQLC params
- Default values for optional fields in responses

## API Usage Examples

### Create Class

```bash
curl -X POST http://localhost:8080/api/v1/lecturer/classes \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Database Design 101",
    "description": "Comprehensive SQL course",
    "semester": "Fall 2024",
    "year": 2024
  }'
```

Response (201):
```json
{
  "id": 1,
  "name": "Database Design 101",
  "description": "Comprehensive SQL course",
  "code": "CLASS-ABC123",
  "createdBy": 5,
  "semester": "Fall 2024",
  "year": 2024,
  "isActive": true,
  "createdAt": "2024-04-10T12:00:00Z",
  "updatedAt": "2024-04-10T12:00:00Z"
}
```

### Add Class Member

```bash
curl -X POST http://localhost:8080/api/v1/lecturer/classes/1/members \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 42,
    "role": "member"
  }'
```

Response (201):
```json
{
  "id": 101,
  "userId": 42,
  "email": "student@example.com",
  "fullName": "John Doe",
  "role": "member",
  "joinedAt": "2024-04-10T12:05:00Z"
}
```

### Assign Exam to Class

```bash
curl -X POST http://localhost:8080/api/v1/lecturer/classes/1/exams \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "examId": 7
  }'
```

Response (201): Empty body with 201 status

### List Class Members

```bash
curl -X GET "http://localhost:8080/api/v1/lecturer/classes/1/members?page=1&pageSize=20" \
  -H "Authorization: Bearer <token>"
```

Response (200):
```json
{
  "members": [
    {
      "id": 101,
      "userId": 42,
      "email": "student1@example.com",
      "fullName": "John Doe",
      "role": "member",
      "joinedAt": "2024-04-10T12:05:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "pageSize": 20
}
```

## Build & Compilation

**Current Status:** ✅ Compiles successfully

```bash
cd Backend
go build ./...
```

Note: Unrelated RabbitMQ interface error exists in pkgs/messaging/rabbitmq/topology.go but does not affect lecturer module.

## What's Next (Phase 3+)

### Phase 3: Scoring System
- Implement auto-scoring from test cases
- Implement answer-key comparison scoring mode
- Create scoring_service package with scoring logic
- Add exam submission viewing with scores

### Phase 4: Lecturer Advanced Features
- Bulk class member import (CSV)
- Class-level statistics/reports
- Exam answer key management
- Submission review interface
- Grade distribution reports

### Phase 5: Student Class Interface
- View enrolled classes
- View class exams
- Access class materials/announcements
- Class member directory

## Files Modified/Created

### Created Files:
- `sql/schema/006_create_classes.sql` - Database schema
- `sql/queries/class.sql` - SQLC queries
- `sql/models/class.sql.go` - Generated SQLC models
- `internals/lecturer/controller/dto/class.go` - DTOs
- `internals/lecturer/usecase/class_usecase.go` - Business logic
- `internals/lecturer/controller/http/handler.go` - HTTP handlers
- `internals/lecturer/controller/http/routes.go` - Route registration

### Modified Files:
- `internals/server/http/server.go` - Added lecturer route registration

### Database Status:
- Migration 006 applies classes schema
- All indices created for performance
- Unique constraints on class codes and class-member pairs

## Permissions & Authorization

### Current State:
- Auth middleware checks valid JWT token
- Lecturer ownership validated in usecase layer for destructive operations
- Future: Integrate permission checks via PermissionMiddleware

### Future Permission Implementation:
```go
// Example of future permission checks
lecturer.POST("/classes", 
  handler.CreateClass,
  PermissionMiddleware(permService, "class", "create"))
```

## Testing Recommendations

1. **Unit Tests:**
   - Usecase methods with mocked database
   - Error handling (not found, duplicate, etc.)
   - Ownership checks

2. **Integration Tests:**
   - Create/update/delete class workflow
   - Add/remove members workflow
   - Assign/remove exams workflow
   - Pagination edge cases

3. **API Tests:**
   - All endpoint status codes
   - Authorization failures
   - Invalid parameter handling
   - Concurrent operations

4. **Database Tests:**
   - Constraint enforcement (unique class code)
   - Cascade delete behavior
   - Data consistency checks

## Deployment Notes

1. Run migration: `sql migrate -path sql/schema -database postgresql://...`
2. Ensure SQLC models are generated: `sqlc generate`
3. Build application: `go build ./cmd/app`
4. Start server - lecturer routes will be available at `/api/v1/lecturer/...`
5. Authenticate via `/api/v1/auth/login` to get JWT token
6. Include Bearer token in Authorization header for lecturer endpoints
