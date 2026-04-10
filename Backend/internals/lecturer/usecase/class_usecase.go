package usecase

import (
	"context"
	"errors"

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

// NOTE: Class management functionality not yet implemented
// All methods below return "not implemented" errors until class schema is ready

// CreateClass - Create new class (lecturer only)
func (u *lecturerClassUseCase) CreateClass(ctx context.Context, lectureID int64, req *dto.CreateClassRequest) (*dto.ClassResponse, error) {
	return nil, errors.New("class management functionality not yet implemented")
}

// GetClass - Get class by ID
func (u *lecturerClassUseCase) GetClass(ctx context.Context, classID int64) (*dto.ClassResponse, error) {
	return nil, errors.New("class management functionality not yet implemented")
}

// ListClasses - List classes for lecturer
func (u *lecturerClassUseCase) ListClasses(ctx context.Context, lectureID int64, page, pageSize int) (*dto.ListClassesResponse, error) {
	return nil, errors.New("class management functionality not yet implemented")
}

// UpdateClass - Update class details
func (u *lecturerClassUseCase) UpdateClass(ctx context.Context, classID int64, req *dto.UpdateClassRequest) (*dto.ClassResponse, error) {
	return nil, errors.New("class management functionality not yet implemented")
}

// DeleteClass - Delete class
func (u *lecturerClassUseCase) DeleteClass(ctx context.Context, classID, lectureID int64) error {
	return errors.New("class management functionality not yet implemented")
}

// AddClassMember - Add member to class
func (u *lecturerClassUseCase) AddClassMember(ctx context.Context, classID, userID int64, role string) (*dto.ClassMemberResponse, error) {
	return nil, errors.New("class management functionality not yet implemented")
}

// ListClassMembers - List members in class
func (u *lecturerClassUseCase) ListClassMembers(ctx context.Context, classID int64, page, pageSize int) (*dto.ListClassMembersResponse, error) {
	return nil, errors.New("class management functionality not yet implemented")
}

// RemoveClassMember - Remove member from class
func (u *lecturerClassUseCase) RemoveClassMember(ctx context.Context, classID, userID int64) error {
	return errors.New("class management functionality not yet implemented")
}

// AssignExamToClass - Assign exam to class
func (u *lecturerClassUseCase) AssignExamToClass(ctx context.Context, classID, examID int64) error {
	return errors.New("class management functionality not yet implemented")
}

// ListClassExams - List exams assigned to class
func (u *lecturerClassUseCase) ListClassExams(ctx context.Context, classID int64) (*dto.ListClassExamsResponse, error) {
	return nil, errors.New("class management functionality not yet implemented")
}

// RemoveExamFromClass - Remove exam from class
func (u *lecturerClassUseCase) RemoveExamFromClass(ctx context.Context, classID, examID int64) error {
	return errors.New("class management functionality not yet implemented")
}
