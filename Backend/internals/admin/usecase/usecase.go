package usecase

import (
	"context"
	"errors"

	"backend/db"
	"backend/internals/admin/controller/dto"
	"backend/sql/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

const defaultPassword = "123456" // Default password for imported users

type IAdminUseCase interface {
	ImportUsers(ctx context.Context, req *dto.ImportUsersRequest) (*dto.ImportResult, error)
	GetSystemStats(ctx context.Context) (*dto.SystemStatsResponse, error)
	ListUsers(ctx context.Context, page, pageSize int) (*dto.UserListResponse, error)
	UpdateUserRole(ctx context.Context, userID int64, role string) error
	ToggleUserActive(ctx context.Context, userID int64, isActive bool) error
}

type adminUseCase struct {
	db      *db.Database
	queries *models.Queries
}

func NewAdminUseCase(database *db.Database) IAdminUseCase {
	return &adminUseCase{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (u *adminUseCase) ImportUsers(ctx context.Context, req *dto.ImportUsersRequest) (*dto.ImportResult, error) {
	result := &dto.ImportResult{
		TotalCount: len(req.Users),
		Errors:     make([]dto.ImportError, 0),
	}

	for i, userData := range req.Users {
		// Check email exists
		exists, _ := u.queries.EmailExists(ctx, userData.Email)
		if exists {
			result.FailedCount++
			result.Errors = append(result.Errors, dto.ImportError{
				Row:     i + 1,
				Email:   userData.Email,
				Message: "Email already exists",
			})
			continue
		}

		// Check username exists
		usernameExists, _ := u.queries.UsernameExists(ctx, userData.Username)
		if usernameExists {
			result.FailedCount++
			result.Errors = append(result.Errors, dto.ImportError{
				Row:     i + 1,
				Email:   userData.Email,
				Message: "Username already exists",
			})
			continue
		}

		// Hash password
		password := userData.Password
		if password == "" {
			password = defaultPassword
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, dto.ImportError{
				Row:     i + 1,
				Email:   userData.Email,
				Message: "Failed to hash password",
			})
			continue
		}

		// Default role is student
		role := userData.Role
		if role == "" {
			role = "student"
		}

		// Create user
		_, err = u.queries.CreateUser(ctx, models.CreateUserParams{
			Email:        userData.Email,
			Username:     userData.Username,
			PasswordHash: string(hashedPassword),
			FullName:     userData.FullName,
			Role:         role,
			StudentID:    strPtr(userData.StudentID),
		})

		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, dto.ImportError{
				Row:     i + 1,
				Email:   userData.Email,
				Message: err.Error(),
			})
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

func (u *adminUseCase) GetSystemStats(ctx context.Context) (*dto.SystemStatsResponse, error) {
	userCount, _ := u.queries.CountUsers(ctx)
	problemCount, _ := u.queries.CountProblems(ctx)

	// Count users by role (simplified)
	users, _ := u.queries.ListUsers(ctx, models.ListUsersParams{Limit: 10000, Offset: 0})
	roleCount := make(map[string]int)
	for _, user := range users {
		roleCount[user.Role]++
	}

	return &dto.SystemStatsResponse{
		TotalUsers:    userCount,
		TotalProblems: problemCount,
		UsersByRole:   roleCount,
	}, nil
}

func (u *adminUseCase) ListUsers(ctx context.Context, page, pageSize int) (*dto.UserListResponse, error) {
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	users, err := u.queries.ListUsers(ctx, models.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.UserResponse, len(users))
	for i, u := range users {
		result[i] = dto.UserResponse{
			ID:        u.ID,
			Email:     u.Email,
			Username:  u.Username,
			FullName:  u.FullName,
			Role:      u.Role,
			StudentID: ptrToStr(u.StudentID),
			IsActive:  ptrToBool(u.IsActive),
			CreatedAt: u.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}

	total, _ := u.queries.CountUsers(ctx)

	return &dto.UserListResponse{
		Users:    result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (u *adminUseCase) UpdateUserRole(ctx context.Context, userID int64, role string) error {
	_, err := u.queries.UpdateUserRole(ctx, models.UpdateUserRoleParams{
		ID:   userID,
		Role: role,
	})
	return err
}

func (u *adminUseCase) ToggleUserActive(ctx context.Context, userID int64, isActive bool) error {
	// DeactivateUser just takes an id and sets is_active = FALSE
	// For toggling active, we would need a different query
	// For now, only support deactivation
	if !isActive {
		return u.queries.DeactivateUser(ctx, userID)
	}
	// TODO: Add an ActivateUser query for re-activation
	return nil
}

// Helper functions
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrToBool(b *bool) bool {
	if b == nil {
		return true // default active
	}
	return *b
}
