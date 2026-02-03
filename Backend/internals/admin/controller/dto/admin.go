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

type ToggleUserActiveRequest struct {
	IsActive bool `json:"isActive"`
}
