package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"backend/db"
	"backend/internals/lecturer/controller/dto"
	"backend/pkgs/redis"
	"backend/sql/models"
)

// ILecturerClassUseCase - Class management operations for lecturers
type ILecturerClassUseCase interface {
	// Class CRUD
	CreateClass(ctx context.Context, lectureID int64, req *dto.CreateClassRequest) (*dto.ClassResponse, error)
	GetClass(ctx context.Context, classID int64) (*dto.ClassResponse, error)
	ListClasses(ctx context.Context, lectureID int64, page, pageSize int) (*dto.ListClassesResponse, error)
	UpdateClass(ctx context.Context, classID int64, req *dto.UpdateClassRequest) (*dto.ClassResponse, error)
	DeleteClass(ctx context.Context, classID, lectureID int64) error

	// Class members
	AddClassMember(ctx context.Context, classID, userID int64, role string) (*dto.ClassMemberResponse, error)
	ListClassMembers(ctx context.Context, classID int64, page, pageSize int) (*dto.ListClassMembersResponse, error)
	RemoveClassMember(ctx context.Context, classID, userID int64) error

	// Class exams
	AssignExamToClass(ctx context.Context, classID, examID int64) error
	ListClassExams(ctx context.Context, classID int64) (*dto.ListClassExamsResponse, error)
	RemoveExamFromClass(ctx context.Context, classID, examID int64) error
}

type lecturerClassUseCase struct {
	db      *db.Database
	queries *models.Queries
	cache   redis.IRedis
}

// NewLecturerClassUseCase - Create new lecturer class usecase
func NewLecturerClassUseCase(database *db.Database, cache redis.IRedis) ILecturerClassUseCase {
	return &lecturerClassUseCase{
		db:      database,
		queries: models.New(database.GetPool()),
		cache:   cache,
	}
}

// =============================================
// CLASS CRUD OPERATIONS
// =============================================

// CreateClass - Create new class (lecturer only)
// Params:
//   - lectureID: ID of lecturer creating class
//   - req: CreateClassRequest with name, description, semester, year
//
// Returns: ClassResponse with generated code and details
// Error: Returns error if creation fails
func (u *lecturerClassUseCase) CreateClass(ctx context.Context, lectureID int64, req *dto.CreateClassRequest) (*dto.ClassResponse, error) {
	// Generate unique class code (format: CLASS-XXXXXX)
	code := generateClassCode()

	params := models.CreateClassParams{
		Name:        req.Name,
		Description: &req.Description,
		Code:        code,
		CreatedBy:   lectureID,
		Semester:    &req.Semester,
	}

	if req.Year != nil {
		year := int32(*req.Year)
		params.Year = &year
	}

	class, err := u.queries.CreateClass(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	return classToResponse(class), nil
}

// GetClass - Get class by ID
// Params:
//   - classID: ID of class
//
// Returns: ClassResponse with all details
// Error: Returns error if class not found
func (u *lecturerClassUseCase) GetClass(ctx context.Context, classID int64) (*dto.ClassResponse, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("class:%d", classID)
	if u.cache != nil {
		var cached dto.ClassResponse
		if err := u.cache.Get(cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	class, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	response := classToResponse(class)

	// Cache for 1 hour
	if u.cache != nil {
		u.cache.SetWithExpiration(cacheKey, response, 1*time.Hour)
	}

	return response, nil
}

// ListClasses - List all classes created by lecturer
// Params:
//   - lectureID: ID of lecturer
//   - page: Page number (1-indexed)
//   - pageSize: Records per page
//
// Returns: ListClassesResponse with paginated classes
// Error: Returns error if query fails
func (u *lecturerClassUseCase) ListClasses(ctx context.Context, lectureID int64, page, pageSize int) (*dto.ListClassesResponse, error) {
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	classes, err := u.queries.ListClassesByLecturer(ctx, models.ListClassesByLecturerParams{
		CreatedBy: lectureID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch classes: %w", err)
	}

	responses := make([]dto.ClassResponse, 0, len(classes))
	for _, class := range classes {
		responses = append(responses, *classToResponse(class))
	}

	return &dto.ListClassesResponse{
		Classes:  responses,
		Total:    int64(len(responses)),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateClass - Update class details
// Params:
//   - classID: ID of class
//   - req: UpdateClassRequest with fields to update
//
// Returns: ClassResponse with updated details
// Error: Returns error if update fails
func (u *lecturerClassUseCase) UpdateClass(ctx context.Context, classID int64, req *dto.UpdateClassRequest) (*dto.ClassResponse, error) {
	_, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	// Build update params with provided values
	var name string
	if req.Name != nil {
		name = *req.Name
	}

	params := models.UpdateClassParams{
		ID:          classID,
		Name:        name,
		Description: req.Description,
		Semester:    req.Semester,
	}

	if req.Year != nil {
		year := int32(*req.Year)
		params.Year = &year
	}

	updated, err := u.queries.UpdateClass(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update class: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("class:%d", classID)
	if u.cache != nil {
		u.cache.Remove(cacheKey)
	}

	return classToResponse(updated), nil
}

// DeleteClass - Delete class (lecturer must own it)
// Params:
//   - classID: ID of class
//   - lectureID: ID of lecturer (for ownership check)
//
// Returns: error if deletion fails
// Ownership check must be done at handler level
func (u *lecturerClassUseCase) DeleteClass(ctx context.Context, classID, lectureID int64) error {
	class, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return errors.New("class not found")
	}

	if class.CreatedBy != lectureID {
		return errors.New("only class creator can delete")
	}

	// Invalidate cache before deletion
	cacheKey := fmt.Sprintf("class:%d", classID)
	if u.cache != nil {
		u.cache.Remove(cacheKey)
	}

	return u.queries.DeleteClass(ctx, classID)
}

// =============================================
// CLASS MEMBER OPERATIONS
// =============================================

// AddClassMember - Add student to class
// Params:
//   - classID: ID of class
//   - userID: ID of student to add
//   - role: "member" or "ta"
//
// Returns: ClassMemberResponse with member details
// Error: Returns error if user already in class or user not found
func (u *lecturerClassUseCase) AddClassMember(ctx context.Context, classID, userID int64, role string) (*dto.ClassMemberResponse, error) {
	// Check class exists
	_, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	// Check user exists
	user, err := u.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Add to class
	member, err := u.queries.AddClassMember(ctx, models.AddClassMemberParams{
		ClassID: classID,
		UserID:  userID,
		Role:    &role,
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, errors.New("user already in class")
		}
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	roleStr := ""
	if member.Role != nil {
		roleStr = *member.Role
	}

	return &dto.ClassMemberResponse{
		ID:       member.ID,
		UserID:   userID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     roleStr,
		JoinedAt: member.JoinedAt.Time.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// ListClassMembers - List all members in class
// Params:
//   - classID: ID of class
//   - page: Page number
//   - pageSize: Records per page
//
// Returns: ListClassMembersResponse with members
// Error: Returns error if query fails
func (u *lecturerClassUseCase) ListClassMembers(ctx context.Context, classID int64, page, pageSize int) (*dto.ListClassMembersResponse, error) {
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	members, err := u.queries.ListClassMembers(ctx, models.ListClassMembersParams{
		ClassID: classID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members: %w", err)
	}

	responses := make([]dto.ClassMemberResponse, 0, len(members))
	for _, member := range members {
		roleStr := ""
		if member.Role != nil {
			roleStr = *member.Role
		}

		responses = append(responses, dto.ClassMemberResponse{
			ID:       member.ID,
			UserID:   member.UserID,
			Email:    member.Email,
			FullName: member.FullName,
			Role:     roleStr,
			JoinedAt: member.JoinedAt.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	total, _ := u.queries.CountClassMembers(ctx, classID)

	return &dto.ListClassMembersResponse{
		Members:  responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// RemoveClassMember - Remove student from class
// Params:
//   - classID: ID of class
//   - userID: ID of student to remove
//
// Returns: error if removal fails
// Error: Returns error if not found or removal fails
func (u *lecturerClassUseCase) RemoveClassMember(ctx context.Context, classID, userID int64) error {
	// Check member exists
	_, err := u.queries.GetClassMember(ctx, models.GetClassMemberParams{
		ClassID: classID,
		UserID:  userID,
	})
	if err != nil {
		return errors.New("member not found in class")
	}

	return u.queries.RemoveClassMember(ctx, models.RemoveClassMemberParams{
		ClassID: classID,
		UserID:  userID,
	})
}

// =============================================
// CLASS EXAM OPERATIONS
// =============================================

// AssignExamToClass - Assign exam to class
// Params:
//   - classID: ID of class
//   - examID: ID of exam
//
// Returns: error if assignment fails
// Note: Exam must be owned by same lecturer (check at handler level)
func (u *lecturerClassUseCase) AssignExamToClass(ctx context.Context, classID, examID int64) error {
	// Check class exists
	_, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return errors.New("class not found")
	}

	// Check exam exists
	_, err = u.queries.GetExamByID(ctx, examID)
	if err != nil {
		return errors.New("exam not found")
	}

	_, err = u.queries.AssignExamToClass(ctx, models.AssignExamToClassParams{
		ClassID: classID,
		ExamID:  examID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return errors.New("exam already assigned to class")
		}
		return fmt.Errorf("failed to assign exam: %w", err)
	}

	return nil
}

// ListClassExams - List all exams assigned to class
// Params:
//   - classID: ID of class
//
// Returns: ListClassExamsResponse with exams
// Error: Returns error if query fails
func (u *lecturerClassUseCase) ListClassExams(ctx context.Context, classID int64) (*dto.ListClassExamsResponse, error) {
	exams, err := u.queries.ListClassExams(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exams: %w", err)
	}

	responses := make([]dto.ClassExamResponse, 0, len(exams))
	for _, exam := range exams {
		responses = append(responses, dto.ClassExamResponse{
			ID:     exam.ID,
			ExamID: exam.ID,
			Title:  exam.Title,
			Status: ptrToStr(exam.Status),
		})
	}

	return &dto.ListClassExamsResponse{
		Exams: responses,
		Total: int64(len(responses)),
	}, nil
}

// RemoveExamFromClass - Remove exam from class
// Params:
//   - classID: ID of class
//   - examID: ID of exam
//
// Returns: error if removal fails
func (u *lecturerClassUseCase) RemoveExamFromClass(ctx context.Context, classID, examID int64) error {
	return u.queries.RemoveExamFromClass(ctx, models.RemoveExamFromClassParams{
		ClassID: classID,
		ExamID:  examID,
	})
}

// =============================================
// HELPER FUNCTIONS
// =============================================

// generateClassCode - Generate unique class code
// Format: CLASS-XXXXXX (6 random alphanumeric)
func generateClassCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 6)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return "CLASS-" + string(result)
}

// classToResponse - Convert Class model to DTO
func classToResponse(class any) *dto.ClassResponse {
	if cls, ok := class.(models.Class); ok {
		year := 0
		if cls.Year != nil {
			year = int(*cls.Year)
		}
		isActive := true
		if cls.IsActive != nil {
			isActive = *cls.IsActive
		}
		return &dto.ClassResponse{
			ID:          cls.ID,
			Name:        cls.Name,
			Description: ptrToStr(cls.Description),
			Code:        cls.Code,
			CreatedBy:   cls.CreatedBy,
			Semester:    ptrToStr(cls.Semester),
			Year:        year,
			IsActive:    isActive,
			CreatedAt:   cls.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   cls.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}
	return nil
}

// ptrToStr - Convert *string to string
func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
