package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"backend/db"
	"backend/internals/admin/controller/dto"
	"backend/pkgs/redis"
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
	ListRoles(ctx context.Context) ([]dto.RoleResponse, error)
	ListUsers(ctx context.Context, page, pageSize int) (*dto.UserListResponse, error)
	UpdateUser(ctx context.Context, userID int64, req *dto.UpdateUserRequest) error
	UpdateUserRole(ctx context.Context, userID int64, role string) error
	ToggleUserActive(ctx context.Context, userID int64, isActive bool) error
	DeleteUser(ctx context.Context, userID int64) error

	// Dashboard & Analytics
	GetDashboard(ctx context.Context) (*dto.DashboardResponse, error)
	GetPerformanceTimeline(ctx context.Context, userID int64, problemID *int64) (*dto.PerformanceTimelineResponse, error)

	// Permission management methods
	GrantRoleToUser(ctx context.Context, userID int64, roleID int32, performedBy int64) error
	RevokeRoleFromUser(ctx context.Context, userID int64, roleID int32) error
	GetUserRoles(ctx context.Context, userID int64) (*dto.UserRoleResponse, error)
	ListPermissions(ctx context.Context) (*dto.ListPermissionsResponse, error)
	GetRolePermissions(ctx context.Context, roleID int32) (*dto.RolePermissionsResponse, error)
	GrantPermissionToRole(ctx context.Context, roleID int32, permissionID int32, performedBy int64) error
	RevokePermissionFromRole(ctx context.Context, roleID int32, permissionID int32) error
	GetAuditLog(ctx context.Context, page, pageSize int) (*dto.AuditLogResponse, error)
}

type adminUseCase struct {
	db      *db.Database
	queries *models.Queries
	cache   redis.IRedis
}

func NewAdminUseCase(database *db.Database, cache redis.IRedis) IAdminUseCase {
	return &adminUseCase{
		db:      database,
		queries: models.New(database.GetPool()),
		cache:   cache,
	}
}

// =============================================
// EXISTING USER MANAGEMENT METHODS
// =============================================

func (u *adminUseCase) ImportUsers(ctx context.Context, req *dto.ImportUsersRequest) (*dto.ImportResult, error) {
	result := &dto.ImportResult{
		TotalCount: len(req.Users),
		Errors:     make([]dto.ImportError, 0),
	}

	for i, userData := range req.Users {
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

		role := userData.Role
		if role == "" {
			role = "student"
		}

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
	// Try to get from cache first
	cacheKey := "stats:system"
	if u.cache != nil {
		var cached dto.SystemStatsResponse
		if err := u.cache.Get(cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	userCount, _ := u.queries.CountUsers(ctx)
	problemCount, _ := u.queries.CountProblems(ctx)

	users, _ := u.queries.ListUsers(ctx, models.ListUsersParams{Limit: 10000, Offset: 0})
	roleCount := make(map[string]int)
	for _, user := range users {
		roleCount[user.Role]++
	}

	response := &dto.SystemStatsResponse{
		TotalUsers:    userCount,
		TotalProblems: problemCount,
		UsersByRole:   roleCount,
	}

	// Cache for 1 hour
	if u.cache != nil {
		u.cache.SetWithExpiration(cacheKey, response, 1*time.Hour)
	}

	return response, nil
}

func (u *adminUseCase) ListRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := u.queries.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}

	result := make([]dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		result = append(result, dto.RoleResponse{
			ID:          strconv.Itoa(int(role.ID)),
			Name:        role.Name,
			Description: ptrToStr(role.Description),
		})
	}

	// Fallback to legacy roles if table is empty (unlikely with seed data)
	if len(result) == 0 {
		return []dto.RoleResponse{
			{ID: "student", Name: "Student", Description: "Sinh viên"},
			{ID: "lecturer", Name: "Lecturer", Description: "Giảng viên"},
			{ID: "admin", Name: "Admin", Description: "Quản trị viên"},
		}, nil
	}

	return result, nil
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
	for i, user := range users {
		result[i] = dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FullName:  user.FullName,
			Role:      user.Role,
			StudentID: ptrToStr(user.StudentID),
			IsActive:  ptrToBool(user.IsActive),
			CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
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

func (u *adminUseCase) UpdateUser(ctx context.Context, userID int64, req *dto.UpdateUserRequest) error {
	_, err := u.queries.UpdateUser(ctx, models.UpdateUserParams{
		ID:        userID,
		Email:     req.Email,
		Username:  req.Username,
		FullName:  req.FullName,
		StudentID: req.StudentID,
		Role:      req.Role,
	})
	return err
}

func (u *adminUseCase) UpdateUserRole(ctx context.Context, userID int64, role string) error {
	_, err := u.queries.UpdateUserRole(ctx, models.UpdateUserRoleParams{
		ID:   userID,
		Role: role,
	})
	return err
}

func (u *adminUseCase) ToggleUserActive(ctx context.Context, userID int64, isActive bool) error {
	if !isActive {
		return u.queries.DeactivateUser(ctx, userID)
	}
	return nil
}

func (u *adminUseCase) DeleteUser(ctx context.Context, userID int64) error {
	// For now, we use DeactivateUser as a soft delete
	// Since ListUsers and other queries filter by is_active = TRUE
	return u.queries.DeactivateUser(ctx, userID)
}

// =============================================
// PERMISSION MANAGEMENT METHODS
// =============================================

// GrantRoleToUser - Assign role to user
// Params:
//   - userID: ID of user to grant role to
//   - roleID: ID of role to assign
//   - performedBy: ID of admin performing action (for audit log)
//
// Error: Returns error if user or role not found, or if assignment fails
func (u *adminUseCase) GrantRoleToUser(ctx context.Context, userID int64, roleID int32, performedBy int64) error {
	_, err := u.queries.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	_, err = u.queries.GetRoleByID(ctx, roleID)
	if err != nil {
		return errors.New("role not found")
	}

	_, err = u.queries.GrantRoleToUser(ctx, models.GrantRoleToUserParams{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: ptrInt64(performedBy),
	})
	if err != nil {
		return fmt.Errorf("failed to grant role: %w", err)
	}

	// Log audit
	_, _ = u.queries.CreateAuditLog(ctx, models.CreateAuditLogParams{
		UserID: ptrInt64(performedBy),
		Action: "role_assigned",
		Reason: strPtr("Role assignment via admin panel"),
	})

	return nil
}

// RevokeRoleFromUser - Remove role from user
// Params:
//   - userID: ID of user to revoke role from
//   - roleID: ID of role to remove
//
// Error: Returns error if user or role not found, or if revocation fails
func (u *adminUseCase) RevokeRoleFromUser(ctx context.Context, userID int64, roleID int32) error {
	_, err := u.queries.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	_, err = u.queries.GetRoleByID(ctx, roleID)
	if err != nil {
		return errors.New("role not found")
	}

	err = u.queries.RevokeRoleFromUser(ctx, models.RevokeRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	return nil
}

// GetUserRoles - Get all roles assigned to user
// Returns: UserRoleResponse with user info and list of assigned roles
// Error: Returns error if user not found or query fails
func (u *adminUseCase) GetUserRoles(ctx context.Context, userID int64) (*dto.UserRoleResponse, error) {
	user, err := u.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	roles, err := u.queries.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user roles: %w", err)
	}

	roleDetails := make([]dto.RoleDetail, 0, len(roles))
	for _, role := range roles {
		roleDetails = append(roleDetails, dto.RoleDetail{
			ID:          role.ID,
			Name:        role.Name,
			Description: ptrToStr(role.Description),
			AssignedAt:  "",
		})
	}

	return &dto.UserRoleResponse{
		ID:    user.ID,
		Email: user.Email,
		Roles: roleDetails,
	}, nil
}

// ListPermissions - Get all available permissions
// Returns: ListPermissionsResponse with all permissions in system
// Error: Returns error if query fails
func (u *adminUseCase) ListPermissions(ctx context.Context) (*dto.ListPermissionsResponse, error) {
	perms, err := u.queries.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch permissions: %w", err)
	}

	permDetails := make([]dto.PermissionDetail, 0, len(perms))
	for _, perm := range perms {
		permDetails = append(permDetails, dto.PermissionDetail{
			ID:          perm.ID,
			Name:        perm.Name,
			Category:    ptrToStr(perm.Category),
			Description: ptrToStr(perm.Description),
		})
	}

	return &dto.ListPermissionsResponse{
		Permissions: permDetails,
		Total:       int64(len(permDetails)),
	}, nil
}

// GetRolePermissions - Get all permissions assigned to role
// Params:
//   - roleID: ID of role to query
//
// Returns: RolePermissionsResponse with role info and assigned permissions
// Error: Returns error if role not found or query fails
func (u *adminUseCase) GetRolePermissions(ctx context.Context, roleID int32) (*dto.RolePermissionsResponse, error) {
	role, err := u.queries.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, errors.New("role not found")
	}

	perms, err := u.queries.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch role permissions: %w", err)
	}

	permDetails := make([]dto.PermissionDetail, 0, len(perms))
	for _, perm := range perms {
		permDetails = append(permDetails, dto.PermissionDetail{
			ID:          perm.ID,
			Name:        perm.Name,
			Category:    ptrToStr(perm.Category),
			Description: ptrToStr(perm.Description),
		})
	}

	return &dto.RolePermissionsResponse{
		RoleID:      roleID,
		RoleName:    role.Name,
		Permissions: permDetails,
	}, nil
}

// GrantPermissionToRole - Assign permission to role
// Params:
//   - roleID: ID of role to grant permission to
//   - permissionID: ID of permission to assign
//   - performedBy: ID of admin performing action (for audit log)
//
// Error: Returns error if role or permission not found, or if assignment fails
func (u *adminUseCase) GrantPermissionToRole(ctx context.Context, roleID int32, permissionID int32, performedBy int64) error {
	_, err := u.queries.GetRoleByID(ctx, roleID)
	if err != nil {
		return errors.New("role not found")
	}

	_, err = u.queries.GetPermissionByID(ctx, permissionID)
	if err != nil {
		return errors.New("permission not found")
	}

	_, err = u.queries.GrantPermissionToRole(ctx, models.GrantPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
	if err != nil {
		return fmt.Errorf("failed to grant permission: %w", err)
	}

	// Log audit
	_, _ = u.queries.CreateAuditLog(ctx, models.CreateAuditLogParams{
		UserID: ptrInt64(performedBy),
		Action: "permission_granted",
		Reason: strPtr(fmt.Sprintf("Permission %d granted to role %d", permissionID, roleID)),
	})

	return nil
}

// RevokePermissionFromRole - Remove permission from role
// Params:
//   - roleID: ID of role to revoke permission from
//   - permissionID: ID of permission to remove
//
// Error: Returns error if role or permission not found, or if revocation fails
func (u *adminUseCase) RevokePermissionFromRole(ctx context.Context, roleID int32, permissionID int32) error {
	_, err := u.queries.GetRoleByID(ctx, roleID)
	if err != nil {
		return errors.New("role not found")
	}

	_, err = u.queries.GetPermissionByID(ctx, permissionID)
	if err != nil {
		return errors.New("permission not found")
	}

	err = u.queries.RevokePermissionFromRole(ctx, models.RevokePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	return nil
}

// GetAuditLog - Get paginated permission audit log
// Params:
//   - page: Page number (1-indexed)
//   - pageSize: Number of entries per page
//
// Returns: AuditLogResponse with paginated audit log entries
// Error: Returns error if query fails
func (u *adminUseCase) GetAuditLog(ctx context.Context, page, pageSize int) (*dto.AuditLogResponse, error) {
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	logs, err := u.queries.ListAuditLogs(ctx, models.ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch audit log: %w", err)
	}

	entries := make([]dto.AuditLogEntry, 0, len(logs))
	for _, log := range logs {
		entry := dto.AuditLogEntry{
			ID:           log.ID,
			Action:       log.Action,
			UserID:       log.UserID,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			OldValue:     string(log.OldValue),
			NewValue:     string(log.NewValue),
			Reason:       log.Reason,
			IpAddress:    log.IpAddress,
			UserAgent:    log.UserAgent,
			CreatedAt:    log.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}

		entries = append(entries, entry)
	}

	return &dto.AuditLogResponse{
		Logs:  entries,
		Total: int64(len(logs)),
		Page:  page,
		Size:  pageSize,
	}, nil
}

// =============================================
// DASHBOARD & ANALYTICS
// =============================================

func (u *adminUseCase) GetDashboard(ctx context.Context) (*dto.DashboardResponse, error) {
	// Try cache first
	cacheKey := "dashboard:full"
	if u.cache != nil {
		var cached dto.DashboardResponse
		if err := u.cache.Get(cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	response := &dto.DashboardResponse{}

	// 1. Overview stats
	userCount, _ := u.queries.CountUsers(ctx)
	problemCount, _ := u.queries.CountProblems(ctx)
	activeUsers, _ := u.queries.GetActiveUsersWeek(ctx)
	avgSolveTimeRow, _ := u.queries.GetAvgSolveTime(ctx)

	users, _ := u.queries.ListUsers(ctx, models.ListUsersParams{Limit: 10000, Offset: 0})
	roleCount := make(map[string]int)
	for _, user := range users {
		roleCount[user.Role]++
	}

	response.Overview = dto.OverviewStats{
		TotalUsers:      userCount,
		TotalProblems:   problemCount,
		ActiveUsersWeek: activeUsers,
		AvgSolveTimeMs:  avgSolveTimeRow,
		UsersByRole:     roleCount,
	}

	// 2. Grading stats
	gradingStats, err := u.queries.GetSystemGradingStats(ctx)
	if err == nil {
		var passRate float64
		if gradingStats.TotalSubmissions > 0 {
			passRate = float64(gradingStats.TotalCorrect) / float64(gradingStats.TotalSubmissions) * 100
		}

		// Safe type conversion for min/max grading times
		var minMs, maxMs int64
		if v, ok := gradingStats.MinGradingTimeMs.(int64); ok {
			minMs = v
		}
		if v, ok := gradingStats.MaxGradingTimeMs.(int64); ok {
			maxMs = v
		}

		response.GradingStats = dto.GradingStats{
			TotalSubmissions:       gradingStats.TotalSubmissions,
			AvgGradingTimeMs:       gradingStats.AvgGradingTimeMs,
			MinGradingTimeMs:       minMs,
			MaxGradingTimeMs:       maxMs,
			TotalCorrect:           gradingStats.TotalCorrect,
			TotalUsers:             gradingStats.TotalUsers,
			TotalProblemsAttempted: gradingStats.TotalProblemsAttempted,
			PassRate:               passRate,
		}
		response.Overview.TotalSubmissions = gradingStats.TotalSubmissions
	}

	// 3. Daily submissions (last 30 days)
	dailyStats, err := u.queries.GetDailySubmissionStats(ctx)
	if err == nil {
		response.DailySubmissions = make([]dto.DailySubmission, len(dailyStats))
		for i, d := range dailyStats {
			response.DailySubmissions[i] = dto.DailySubmission{
				Date:             d.SubmitDate.Time.Format("2006-01-02"),
				TotalSubmissions: d.TotalSubmissions,
				CorrectCount:     d.CorrectCount,
				AvgExecutionMs:   d.AvgExecutionMs,
			}
		}
	}

	// 4. Pass rate per problem (top 20)
	passRates, err := u.queries.GetPassRatePerProblem(ctx, 20)
	if err == nil {
		response.PassRates = make([]dto.ProblemPassRate, len(passRates))
		for i, p := range passRates {
			response.PassRates[i] = dto.ProblemPassRate{
				ID:               p.ID,
				Title:            p.Title,
				Difficulty:       p.Difficulty,
				TotalSubmissions: p.TotalSubmissions,
				CorrectCount:     p.CorrectCount,
				PassRate:         float64(p.PassRate),
			}
		}
	}

	// 5. Top problems by submissions
	topProblems, err := u.queries.GetTopProblemsBySubmissions(ctx, 10)
	if err == nil {
		response.TopProblems = make([]dto.TopProblem, len(topProblems))
		for i, p := range topProblems {
			response.TopProblems[i] = dto.TopProblem{
				ID:              p.ID,
				Title:           p.Title,
				Slug:            p.Slug,
				Difficulty:      p.Difficulty,
				SubmissionCount: p.SubmissionCount,
				UniqueUsers:     p.UniqueUsers,
			}
		}
	}

	// Cache for 10 minutes
	if u.cache != nil {
		u.cache.SetWithExpiration(cacheKey, response, 10*time.Minute)
	}

	return response, nil
}

func (u *adminUseCase) GetPerformanceTimeline(ctx context.Context, userID int64, problemID *int64) (*dto.PerformanceTimelineResponse, error) {
	var pid int64
	if problemID != nil {
		pid = *problemID
	}

	timeline, err := u.queries.GetUserPerformanceTimeline(ctx, models.GetUserPerformanceTimelineParams{
		UserID:  userID,
		Column2: pid,
	})
	if err != nil {
		return nil, err
	}

	entries := make([]dto.PerformanceTimelineEntry, len(timeline))
	for i, t := range timeline {
		var bestMs int64
		if v, ok := t.BestTimeMs.(int64); ok {
			bestMs = v
		}
		entries[i] = dto.PerformanceTimelineEntry{
			Date:            t.SubmitDate.Time.Format("2006-01-02"),
			AvgTimeMs:       t.AvgTimeMs,
			BestTimeMs:      bestMs,
			SubmissionCount: t.SubmissionCount,
			CorrectCount:    t.CorrectCount,
		}
	}

	return &dto.PerformanceTimelineResponse{
		UserID:    userID,
		ProblemID: problemID,
		Timeline:  entries,
	}, nil
}

// =============================================
// HELPER FUNCTIONS
// =============================================

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
		return true
	}
	return *b
}

func ptrInt32(val int32) *int32 {
	return &val
}

func ptrInt64(val int64) *int64 {
	return &val
}

// jsonStringify - Helper to convert value to []byte for JSON storage
func jsonStringify(value interface{}) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		return []byte{}
	}
	return data
}
