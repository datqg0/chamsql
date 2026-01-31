package dto

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Password  string `json:"password" binding:"required,min=6"`
	FullName  string `json:"fullName" binding:"required,min=2,max=100"`
	StudentID string `json:"studentId" binding:"omitempty,max=20"`
	Role      string `json:"role" binding:"omitempty,oneof=student lecturer admin"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // email or username
	Password   string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken,omitempty"`
	ExpiresIn    int64        `json:"expiresIn"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	Role      string `json:"role"`
	StudentID string `json:"studentId,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type UpdateProfileRequest struct {
	FullName  string `json:"fullName" binding:"omitempty,min=2,max=100"`
	AvatarURL string `json:"avatarUrl" binding:"omitempty,url"`
}
