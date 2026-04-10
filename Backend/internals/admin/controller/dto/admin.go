package dto

type ImportUsersRequest struct {
	Users []ImportUserData `json:"users" binding:"required,min=1"`
}

type ImportUserData struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=50"`
	FullName  string `json:"fullName" binding:"required,min=2,max=100"`
	StudentID string `json:"studentId" binding:"omitempty,max=20"`
	Role      string `json:"role" binding:"omitempty,oneof=student lecturer admin"`
	Password  string `json:"password" binding:"omitempty,min=6"` // If empty, use default
}

type ImportResult struct {
	TotalCount   int           `json:"totalCount"`
	SuccessCount int           `json:"successCount"`
	FailedCount  int           `json:"failedCount"`
	Errors       []ImportError `json:"errors,omitempty"`
}

type ImportError struct {
	Row     int    `json:"row"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type SystemStatsResponse struct {
	TotalUsers       int64          `json:"totalUsers"`
	TotalProblems    int64          `json:"totalProblems"`
	TotalExams       int64          `json:"totalExams"`
	TotalSubmissions int64          `json:"totalSubmissions"`
	UsersByRole      map[string]int `json:"usersByRole"`
	RecentActivity   []Activity     `json:"recentActivity,omitempty"`
}

type Activity struct {
	Type      string `json:"type"` // user_registered, submission, exam_created
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type UserListResponse struct {
	Users    []UserResponse `json:"users"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	Role      string `json:"role"`
	StudentID string `json:"studentId,omitempty"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=student lecturer admin"`
}

type UpdateUserRequest struct {
	Email     *string `json:"email" binding:"omitempty,email"`
	Username  *string `json:"username" binding:"omitempty,min=3,max=50"`
	FullName  *string `json:"fullName" binding:"omitempty,min=2,max=100"`
	StudentID *string `json:"studentId" binding:"omitempty,max=20"`
	Role      *string `json:"role" binding:"omitempty,oneof=student lecturer admin"`
}

type ToggleUserActiveRequest struct {
	IsActive bool `json:"isActive"`
}

type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// =============================================
// PERMISSION MANAGEMENT DTOs
// =============================================

// GrantRoleRequest - Assign role to user
type GrantRoleRequest struct {
	RoleID int32 `json:"roleId" binding:"required"`
}

// RevokeRoleRequest - Remove role from user
type RevokeRoleRequest struct {
	RoleID int32 `json:"roleId" binding:"required"`
}

// UserRoleResponse - User and their assigned roles
type UserRoleResponse struct {
	ID    int64        `json:"id"`
	Email string       `json:"email"`
	Roles []RoleDetail `json:"roles"`
}

// RoleDetail - Role information with metadata
type RoleDetail struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AssignedAt  string `json:"assignedAt"`
}

// PermissionDetail - Permission information
type PermissionDetail struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// RolePermissionsResponse - Show all permissions assigned to a role
type RolePermissionsResponse struct {
	RoleID      int32              `json:"roleId"`
	RoleName    string             `json:"roleName"`
	Permissions []PermissionDetail `json:"permissions"`
}

// GrantPermissionRequest - Assign permission to role
type GrantPermissionRequest struct {
	PermissionID int32 `json:"permissionId" binding:"required"`
}

// ListPermissionsResponse - Paginated list of permissions
type ListPermissionsResponse struct {
	Permissions []PermissionDetail `json:"permissions"`
	Total       int64              `json:"total"`
}

// AuditLogEntry - Single entry in permission audit log
type AuditLogEntry struct {
	ID           int64   `json:"id"`
	Action       string  `json:"action"` // role_assigned, role_revoked, permission_granted, etc
	UserID       *int64  `json:"userId,omitempty"`
	ResourceType *string `json:"resourceType,omitempty"`
	ResourceID   *int64  `json:"resourceId,omitempty"`
	OldValue     string  `json:"oldValue,omitempty"` // JSON string
	NewValue     string  `json:"newValue,omitempty"` // JSON string
	Reason       *string `json:"reason,omitempty"`
	IpAddress    *string `json:"ipAddress,omitempty"`
	UserAgent    *string `json:"userAgent,omitempty"`
	CreatedAt    string  `json:"createdAt"`
}

// AuditLogResponse - Paginated audit log
type AuditLogResponse struct {
	Logs  []AuditLogEntry `json:"logs"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"pageSize"`
}
