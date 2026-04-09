package dto

// =============================================
// CLASS MANAGEMENT DTOs
// =============================================

// CreateClassRequest - Create new class
type CreateClassRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	Semester    string `json:"semester" binding:"omitempty,max=20"`
	Year        *int   `json:"year" binding:"omitempty"`
}

// UpdateClassRequest - Update class details
type UpdateClassRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	Semester    *string `json:"semester" binding:"omitempty,max=20"`
	Year        *int    `json:"year" binding:"omitempty"`
}

// ClassResponse - Class information
type ClassResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
	CreatedBy   int64  `json:"createdBy"`
	Semester    string `json:"semester"`
	Year        int    `json:"year"`
	IsActive    bool   `json:"isActive"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ListClassesResponse - Paginated classes
type ListClassesResponse struct {
	Classes  []ClassResponse `json:"classes"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// =============================================
// CLASS MEMBER DTOs
// =============================================

// AddClassMemberRequest - Add student to class
type AddClassMemberRequest struct {
	UserID int64  `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=member ta"`
}

// BulkAddClassMembersRequest - Add multiple students
type BulkAddClassMembersRequest struct {
	UserIDs []int64 `json:"userIds" binding:"required,min=1"`
	Role    string  `json:"role" binding:"required,oneof=member ta"`
}

// ClassMemberResponse - Class member information
type ClassMemberResponse struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"userId"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
}

// ListClassMembersResponse - Paginated members
type ListClassMembersResponse struct {
	Members  []ClassMemberResponse `json:"members"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}

// =============================================
// CLASS EXAM DTOs
// =============================================

// AssignExamToClassRequest - Assign exam to class
type AssignExamToClassRequest struct {
	ExamID int64 `json:"examId" binding:"required"`
}

// ClassExamResponse - Exam assigned to class
type ClassExamResponse struct {
	ID     int64  `json:"id"`
	ExamID int64  `json:"examId"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// ListClassExamsResponse - Exams in class
type ListClassExamsResponse struct {
	Exams []ClassExamResponse `json:"exams"`
	Total int64               `json:"total"`
}
