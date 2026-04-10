package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
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

// Helper function to generate unique class code (6-8 uppercase alphanumeric characters)
func (u *lecturerClassUseCase) generateClassCode() (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 7
	code := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", fmt.Errorf("failed to generate class code: %w", err)
		}
		code[i] = chars[num.Int64()]
	}

	return string(code), nil
}

// Helper function to convert Class model to ClassResponse DTO
func (u *lecturerClassUseCase) classModelToDTO(class models.Class) *dto.ClassResponse {
	description := ""
	if class.Description != nil {
		description = *class.Description
	}

	semester := fmt.Sprintf("%d", class.Semester)
	year := int(class.Year)

	isActive := false
	if class.IsActive != nil {
		isActive = *class.IsActive
	}

	return &dto.ClassResponse{
		ID:          class.ID,
		Name:        class.Name,
		Description: description,
		Code:        class.Code,
		CreatedBy:   class.CreatedBy,
		Semester:    semester,
		Year:        year,
		IsActive:    isActive,
		CreatedAt:   class.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   class.UpdatedAt.Time.Format(time.RFC3339),
	}
}

// CreateClass - Create new class (lecturer only)
func (u *lecturerClassUseCase) CreateClass(ctx context.Context, lectureID int64, req *dto.CreateClassRequest) (*dto.ClassResponse, error) {
	// Generate unique class code
	code, err := u.generateClassCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate class code: %w", err)
	}

	// Parse semester and year
	semester := int32(0)
	year := int32(0)

	if req.Semester != "" {
		fmt.Sscanf(req.Semester, "%d", &semester)
	}
	if req.Year != nil {
		year = int32(*req.Year)
	}

	// Create class in database
	class, err := u.queries.CreateClass(ctx, models.CreateClassParams{
		Name:        req.Name,
		Description: &req.Description,
		Code:        code,
		CreatedBy:   lectureID,
		Semester:    semester,
		Year:        year,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	return u.classModelToDTO(class), nil
}

// GetClass - Get class by ID
func (u *lecturerClassUseCase) GetClass(ctx context.Context, classID int64) (*dto.ClassResponse, error) {
	class, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("class not found: %w", err)
	}

	return u.classModelToDTO(class), nil
}

// ListClasses - List classes for lecturer with pagination
func (u *lecturerClassUseCase) ListClasses(ctx context.Context, lectureID int64, page, pageSize int) (*dto.ListClassesResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	classes, err := u.queries.ListClassesByLecturer(ctx, models.ListClassesByLecturerParams{
		CreatedBy: lectureID,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list classes: %w", err)
	}

	// Convert to DTOs
	classResponses := make([]dto.ClassResponse, 0)
	for _, class := range classes {
		classResponses = append(classResponses, *u.classModelToDTO(class))
	}

	// Count total classes for pagination
	total := int64(len(classResponses))

	return &dto.ListClassesResponse{
		Classes:  classResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateClass - Update class details
func (u *lecturerClassUseCase) UpdateClass(ctx context.Context, classID int64, req *dto.UpdateClassRequest) (*dto.ClassResponse, error) {
	// Get current class to verify it exists
	currentClass, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("class not found: %w", err)
	}

	// Use current values if update fields are not provided
	name := currentClass.Name
	if req.Name != nil {
		name = *req.Name
	}

	description := currentClass.Description
	if req.Description != nil {
		description = req.Description
	}

	semester := currentClass.Semester
	if req.Semester != nil && *req.Semester != "" {
		var s int32
		fmt.Sscanf(*req.Semester, "%d", &s)
		semester = s
	}

	year := currentClass.Year
	if req.Year != nil {
		year = int32(*req.Year)
	}

	// Update class
	updatedClass, err := u.queries.UpdateClass(ctx, models.UpdateClassParams{
		ID:          classID,
		Name:        name,
		Description: description,
		Semester:    semester,
		Year:        year,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update class: %w", err)
	}

	return u.classModelToDTO(updatedClass), nil
}

// DeleteClass - Delete class (lecturer only)
func (u *lecturerClassUseCase) DeleteClass(ctx context.Context, classID, lectureID int64) error {
	// Verify lecturer owns the class
	class, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return fmt.Errorf("class not found: %w", err)
	}

	if class.CreatedBy != lectureID {
		return fmt.Errorf("unauthorized: you don't own this class")
	}

	// Delete class
	if err := u.queries.DeleteClass(ctx, classID); err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}

	return nil
}

// AddClassMember - Add member to class
func (u *lecturerClassUseCase) AddClassMember(ctx context.Context, classID, userID int64, role string) (*dto.ClassMemberResponse, error) {
	// Verify class exists
	_, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("class not found: %w", err)
	}

	// Check if user already a member
	_, err = u.queries.GetClassMember(ctx, models.GetClassMemberParams{
		ClassID: classID,
		UserID:  userID,
	})
	if err == nil {
		return nil, fmt.Errorf("user is already a member of this class")
	}

	// Add member
	member, err := u.queries.AddClassMember(ctx, models.AddClassMemberParams{
		ClassID: classID,
		UserID:  userID,
		Role:    &role,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add class member: %w", err)
	}

	memberRole := ""
	if member.Role != nil {
		memberRole = *member.Role
	}

	return &dto.ClassMemberResponse{
		ID:       member.ID,
		UserID:   member.UserID,
		Email:    "", // Will be filled by fetching user separately
		FullName: "",
		Role:     memberRole,
		JoinedAt: member.JoinedAt.Time.Format(time.RFC3339),
	}, nil
}

// ListClassMembers - List members in class with pagination
func (u *lecturerClassUseCase) ListClassMembers(ctx context.Context, classID int64, page, pageSize int) (*dto.ListClassMembersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	// Verify class exists
	_, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("class not found: %w", err)
	}

	members, err := u.queries.ListClassMembers(ctx, models.ListClassMembersParams{
		ClassID: classID,
		Offset:  offset,
		Limit:   limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list class members: %w", err)
	}

	// Convert to DTOs
	memberResponses := make([]dto.ClassMemberResponse, 0)
	for _, member := range members {
		memberRole := ""
		if member.Role != nil {
			memberRole = *member.Role
		}

		memberResponses = append(memberResponses, dto.ClassMemberResponse{
			ID:       member.ID,
			UserID:   member.UserID,
			Email:    member.Email,
			FullName: member.FullName,
			Role:     memberRole,
			JoinedAt: member.JoinedAt.Time.Format(time.RFC3339),
		})
	}

	// Get total count
	total, err := u.queries.CountClassMembers(ctx, classID)
	if err != nil {
		total = int64(len(memberResponses))
	}

	return &dto.ListClassMembersResponse{
		Members:  memberResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// RemoveClassMember - Remove member from class
func (u *lecturerClassUseCase) RemoveClassMember(ctx context.Context, classID, userID int64) error {
	// Verify class exists
	_, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return fmt.Errorf("class not found: %w", err)
	}

	// Verify member exists
	_, err = u.queries.GetClassMember(ctx, models.GetClassMemberParams{
		ClassID: classID,
		UserID:  userID,
	})
	if err != nil {
		return fmt.Errorf("member not found in this class: %w", err)
	}

	// Remove member
	if err := u.queries.RemoveClassMember(ctx, models.RemoveClassMemberParams{
		ClassID: classID,
		UserID:  userID,
	}); err != nil {
		return fmt.Errorf("failed to remove class member: %w", err)
	}

	return nil
}

// AssignExamToClass - Assign exam to class
func (u *lecturerClassUseCase) AssignExamToClass(ctx context.Context, classID, examID int64) error {
	// Verify class exists
	_, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return fmt.Errorf("class not found: %w", err)
	}

	// Verify exam exists
	// NOTE: We'll just try to assign; DB will handle constraint if exam doesn't exist

	_, err = u.queries.AssignExamToClass(ctx, models.AssignExamToClassParams{
		ClassID: classID,
		ExamID:  examID,
	})
	if err != nil {
		return fmt.Errorf("failed to assign exam to class: %w", err)
	}

	return nil
}

// ListClassExams - List exams assigned to class
func (u *lecturerClassUseCase) ListClassExams(ctx context.Context, classID int64) (*dto.ListClassExamsResponse, error) {
	// Verify class exists
	_, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("class not found: %w", err)
	}

	exams, err := u.queries.ListClassExams(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to list class exams: %w", err)
	}

	// Convert to DTOs
	examResponses := make([]dto.ClassExamResponse, 0)
	for _, exam := range exams {
		status := ""
		if exam.Status != nil {
			status = *exam.Status
		}

		examResponses = append(examResponses, dto.ClassExamResponse{
			ID:     exam.ID,
			ExamID: exam.ID_2,
			Title:  exam.Title,
			Status: status,
		})
	}

	return &dto.ListClassExamsResponse{
		Exams: examResponses,
		Total: int64(len(examResponses)),
	}, nil
}

// RemoveExamFromClass - Remove exam from class
func (u *lecturerClassUseCase) RemoveExamFromClass(ctx context.Context, classID, examID int64) error {
	// Verify class exists
	_, err := u.queries.GetClassByID(ctx, classID)
	if err != nil {
		return fmt.Errorf("class not found: %w", err)
	}

	// Remove exam
	if err := u.queries.RemoveExamFromClass(ctx, models.RemoveExamFromClassParams{
		ClassID: classID,
		ExamID:  examID,
	}); err != nil {
		return fmt.Errorf("failed to remove exam from class: %w", err)
	}

	return nil
}
