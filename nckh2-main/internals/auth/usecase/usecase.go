package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/configs"
	"backend/internals/auth/controller/dto"
	"backend/internals/auth/repository"
	"backend/pkgs/jwt"
	"backend/pkgs/redis"
	"backend/sql/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already registered")
	ErrUsernameExists     = errors.New("username already taken")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotActive      = errors.New("user account is not active")
)

type IAuthUseCase interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error)
}

type authUseCase struct {
	repo    repository.IAuthRepository
	jwtProv jwt.JWTProvider
	cache   redis.IRedis
	cfg     *configs.Config
}

func NewAuthUseCase(
	repo repository.IAuthRepository,
	jwtProv jwt.JWTProvider,
	cache redis.IRedis,
	cfg *configs.Config,
) IAuthUseCase {
	return &authUseCase{
		repo:    repo,
		jwtProv: jwtProv,
		cache:   cache,
		cfg:     cfg,
	}
}

func (u *authUseCase) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check email exists
	emailExists, err := u.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, ErrEmailExists
	}

	// Check username exists
	usernameExists, err := u.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Default role is student
	role := req.Role
	if role == "" {
		role = "student"
	}

	// Create user
	user, err := u.repo.CreateUser(ctx, models.CreateUserParams{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Role:         role,
		StudentID:    stringPtr(req.StudentID),
	})
	if err != nil {
		return nil, err
	}

	return u.generateAuthResponse(ctx, user)
}

func (u *authUseCase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by identifier (email or username)
	user, err := u.repo.GetUserByIdentifier(ctx, req.Identifier)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !isUserActive(user) {
		return nil, ErrUserNotActive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return u.generateAuthResponse(ctx, user)
}

func (u *authUseCase) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// Check refresh token in Redis
	key := fmt.Sprintf("refresh_token:%s", req.RefreshToken)
	var userID int64
	err := u.cache.Get(key, &userID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user is still active
	if !isUserActive(user) {
		return nil, ErrUserNotActive
	}

	// Rotate token: delete old, create new
	u.cache.Remove(key)

	return u.generateAuthResponse(ctx, user)
}

func (u *authUseCase) Logout(ctx context.Context, tokenString string) error {
	// Validate token format & signature
	claims, err := u.jwtProv.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	// Calculate remaining time for expiration
	expFloat, ok := (*claims)["exp"].(float64)
	if !ok {
		return ErrInvalidToken
	}
	expTime := time.Unix(int64(expFloat), 0)
	remainingTime := time.Until(expTime)

	if remainingTime <= 0 {
		return nil // Already expired
	}

	// Add to Redis blacklist
	if u.cache != nil && u.cache.IsConnected() {
		blacklistKey := fmt.Sprintf("blacklist:%s", tokenString)
		err := u.cache.SetWithExpiration(blacklistKey, "revoked", remainingTime)
		if err != nil {
			return fmt.Errorf("failed to blacklist token: %w", err)
		}
	}

	return nil
}

func (u *authUseCase) generateAuthResponse(ctx context.Context, user *models.User) (*dto.AuthResponse, error) {
	// Generate access token
	td, err := u.jwtProv.GenerateToken(user.ID, user.Role, u.cfg.AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (UUID)
	refreshToken := uuid.New().String()

	// Store refresh token in Redis
	if u.cache != nil && u.cache.IsConnected() {
		key := fmt.Sprintf("refresh_token:%s", refreshToken)
		_ = u.cache.SetWithExpiration(key, user.ID, u.cfg.RefreshTokenDuration)
	}

	return &dto.AuthResponse{
		AccessToken:  td.AccessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(u.cfg.AccessTokenDuration.Seconds()),
		User: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FullName:  user.FullName,
			Role:      user.Role,
			StudentID: ptrToString(user.StudentID),
			AvatarURL: ptrToString(user.AvatarUrl),
		},
	}, nil
}

// Helper functions
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isUserActive(user *models.User) bool {
	if user.IsActive == nil {
		return true // default to active if nil
	}
	return *user.IsActive
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
