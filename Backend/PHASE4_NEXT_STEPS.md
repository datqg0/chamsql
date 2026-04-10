# 🎯 PHASE 4 Execution Roadmap - Continue Building

**Current Status**: Week 1 Day 1 - Foundation Complete ✅  
**Next Priority**: AI Services Layer (Days 2-3)

---

## 📋 What's Done vs What's Next

### ✅ COMPLETED (Day 1)
1. Database migration (5 new tables, 12 indexes)
2. Domain models (17 models across 3 modules)
3. AI infrastructure (HuggingFace client, pattern matcher)
4. PDF parser infrastructure
5. Solution generator (hybrid approach)
6. All code compiles successfully

### 🔄 NEXT (Days 2-3)
1. **Test Case Generator** - Generate test data from schema + solution
2. **Test Case Validator** - Execute in sandbox to verify correctness
3. **AI Orchestrator** - Coordinate all AI services
4. **Caching Layer** - Cache AI outputs to reduce API calls
5. **Repository Layer** - Database access for AI content

### ⏳ THEN (Days 4-5)
1. PDF upload endpoint (multipart/form-data)
2. File storage (MinIO integration)
3. Problem review workflow
4. Excel export service

### 🎁 FINALLY (Days 6-7)
1. E2E testing
2. Integration testing
3. Error handling
4. Documentation

---

## 🚀 Immediate Next Steps (Continue from here)

### STEP 1: Complete AI Service Layer

**File to create**: `internals/ai/usecase/test_case_generator.go`

```go
// IAITestCaseGenerator interface
type IAITestCaseGenerator interface {
    GenerateTestCases(ctx context.Context, req domain.TestCaseGenerationInput) ([]domain.TestCaseGenerated, error)
}

// Implementation
type aiTestCaseGenerator struct {
    // Dependencies
}

// GenerateTestCases should:
// 1. Parse schema to understand tables/columns
// 2. Generate basic test data (empty, single row, multiple rows)
// 3. Generate boundary cases (NULL, max values)
// 4. Run solution query to get expected outputs
// 5. Return formatted test cases
```

**File to create**: `internals/ai/usecase/validator.go`

```go
// IAITestCaseValidator interface
type IAITestCaseValidator interface {
    ValidateTestCases(ctx context.Context, schemaSQL, solutionSQL string, testCases []domain.TestCaseData) (*domain.ValidationResult, error)
}

// Implementation should:
// 1. Create sandbox database
// 2. Execute schema
// 3. For each test case:
//    - Insert test data
//    - Run solution query
//    - Compare with expected output
//    - Track pass/fail
// 4. Return validation summary
```

**File to create**: `internals/ai/usecase/orchestrator.go`

```go
// AIOrchestrator coordinates all AI services
type AIOrchestrator struct {
    solutionGenerator IAISolutionGenerator
    testCaseGenerator IAITestCaseGenerator
    validator         IAITestCaseValidator
}

// Orchestrate:
// 1. Generate solution from description
// 2. Generate test cases from schema + solution
// 3. Validate test cases (execute & verify)
// 4. Return complete problem with all components
```

---

### STEP 2: Create PDF Upload Service

**File to create**: `internals/pdf/usecase/upload_manager.go`

```go
// IPDFUploadManager handles the upload workflow
type IPDFUploadManager interface {
    UploadPDF(ctx context.Context, lecturerID int64, file multipart.File, filename string) (*domain.PDFUpload, error)
    GetUploadStatus(ctx context.Context, uploadID int64) (*domain.PDFUpload, error)
    ExtractProblems(ctx context.Context, uploadID int64) ([]domain.ExtractedProblem, error)
}

// Workflow:
// 1. Validate file (is it PDF?)
// 2. Store in MinIO
// 3. Extract text from PDF
// 4. Parse problems
// 5. Store in database
// 6. Return extracted problems for review
```

---

### STEP 3: Create Repository Layer

**Files to create**:
- `internals/pdf/repository/pdf_repository.go` - PDF CRUD
- `internals/ai/repository/ai_repository.go` - AI content caching
- `internals/export/repository/export_repository.go` - Export tracking

Each repository should handle:
- Insert/Update/Get operations
- Proper error handling
- Database query builders

---

### STEP 4: Create API Endpoints

**Files to create**:
- `internals/pdf/controller/http/handler.go` - PDF upload handler
- `internals/ai/controller/http/handler.go` - AI generation endpoints
- `internals/export/controller/http/handler.go` - Export endpoints

Endpoints needed:
```
POST   /api/v1/lecturer/pdf/upload              - Upload PDF
GET    /api/v1/lecturer/pdf/{id}                - Get upload status
GET    /api/v1/lecturer/pdf/{id}/problems       - Get extracted problems
POST   /api/v1/lecturer/problem/review          - Submit for review
PUT    /api/v1/lecturer/problem/review/{id}    - Approve/reject
POST   /api/v1/lecturer/export/results          - Export results to Excel
GET    /api/v1/lecturer/export/{id}             - Download export
```

---

## 🎯 Recommended Execution Order

```
DAY 1 ✅ COMPLETE
└─ Foundation: Database, models, AI infrastructure

DAY 2 🔄 IN PROGRESS
├─ Test Case Generator (generate test data)
├─ Test Case Validator (execute & verify)
└─ AI Orchestrator (coordinate all)

DAY 3
├─ Caching layer (cache AI outputs)
├─ Repository layer (database access)
└─ Unit tests for AI services

DAY 4
├─ PDF upload manager
├─ MinIO integration
└─ PDF repository

DAY 5
├─ Problem review workflow
├─ Edit & approve logic
└─ Review repository

DAY 6
├─ Excel export service
├─ Results exporter
└─ Analytics exporter

DAY 7
├─ API endpoints (all)
├─ E2E testing
└─ Integration testing
```

---

## 💻 Example Implementation Pattern

For each service you build, follow this pattern:

```go
// 1. DOMAIN (already created)
// Location: internals/<module>/domain/models.go
// Contains: struct definitions, constants

// 2. USECASE (interface + implementation)
// Location: internals/<module>/usecase/*.go
type IMyService interface {
    DoSomething(ctx context.Context, input InputType) (*OutputType, error)
}

type myService struct {
    // Dependencies injected
    db    *db.Database
    cache redis.IRedis
}

func NewMyService(db *db.Database, cache redis.IRedis) IMyService {
    return &myService{db: db, cache: cache}
}

// 3. REPOSITORY (database access)
// Location: internals/<module>/repository/repository.go
type IMyRepository interface {
    Save(ctx context.Context, item *domain.Item) error
    Get(ctx context.Context, id int64) (*domain.Item, error)
}

type myRepository struct {
    db *db.Database
}

// 4. CONTROLLER (HTTP handler)
// Location: internals/<module>/controller/http/handler.go
func (h *handler) HandlePostRequest(c *gin.Context) {
    // Parse request
    // Call usecase
    // Return response
}

// 5. ROUTES (register endpoints)
// Location: internals/<module>/controller/http/routes.go
func RegisterRoutes(router *gin.Engine, usecase UseCase) {
    router.POST("/api/v1/resource", handler.Create)
    router.GET("/api/v1/resource/:id", handler.Get)
}
```

---

## 🧪 Testing Strategy

For each service, create tests:

```go
// Unit tests: internals/<module>/usecase/*_test.go
func TestGenerateSolution(t *testing.T) {
    // Arrange
    // Act
    // Assert
}

// Integration tests: internal/<module>/integration_test.go
func TestUploadAndParsePDF(t *testing.T) {
    // Full workflow test
}
```

Target: >80% coverage for all new code

---

## 🔧 Configuration Checklist

Before running Phase 4 endpoints:

```
☐ PostgreSQL running and migration 008 applied
☐ Redis running (for caching)
☐ MinIO running (for file storage)
☐ HuggingFace API key in .env (HUGGINGFACE_API_KEY)
☐ Database credentials in .env
☐ Port 8080 available (or configured)
```

---

## ✅ Phase 4 Complete Criteria

- ✅ Database migrations applied successfully
- ✅ All 17 domain models compile
- ✅ AI infrastructure working (pattern matcher, LLM client)
- ✅ PDF parser extracts problems correctly
- ✅ Solution generator produces valid SQL (>80% accuracy)
- ✅ Test case generator creates valid test cases
- ✅ Test case validator correctly identifies pass/fail
- ✅ PDF upload endpoint working
- ✅ Problem review workflow implemented
- ✅ Excel export generating correct files
- ✅ All endpoints have proper error handling
- ✅ >80% test coverage
- ✅ E2E test: Upload → Review → Create → Export ✅

---

## 📞 Quick Reference

**Key Files Created Today**:
```
internals/pdf/domain/models.go              (7 models)
internals/ai/domain/models.go               (6 models)
internals/export/domain/models.go           (4 models)
internals/ai/usecase/solution_generator.go  (Generator)
pkgs/ai/huggingface.go                      (LLM client)
pkgs/ai/pattern_matcher.go                  (Pattern-based)
pkgs/pdf/parser.go                          (PDF parsing)
sql/schema/008_pdf_upload_ai_generation.sql (Migration)
```

**Build Status**: ✅ PASSING

**Next Immediate Tasks**:
1. Test Case Generator
2. Test Case Validator
3. AI Orchestrator
4. Unit tests

---

**Ready to continue? Let me know when you're ready, and I'll implement the next layer! 🚀**
