package usecase

import (
	"context"
	"errors"
	"time"

	"backend/configs"
	"backend/db"
	"backend/internals/auth/controller/dto"
	"backend/internals/auth/repository"
	"backend/pkgs/jwt"
	"backend/pkgs/redis"
	"backend/sql/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	queries *models.Queries
}

func NewAuthUseCase(
	repo repository.IAuthRepository,
	jwtProv jwt.JWTProvider,
	cache redis.IRedis,
	database *db.Database, // Added database dependency
	cfg *configs.Config,
) IAuthUseCase {
	return &authUseCase{
		repo:    repo,
		jwtProv: jwtProv,
		cache:   cache,
		cfg:     cfg,
		queries: models.New(database.GetPool()),
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
	// 1. Get token from DB
	storedToken, err := u.queries.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 2. Check if revoked
	if storedToken.Revoked != nil && *storedToken.Revoked {
		// Security: If a revoked token is used, it might be a theft.
		// We could revoke all user sessions here for safety.
		// u.queries.RevokeAllUserTokens(ctx, storedToken.UserID)
		return nil, ErrInvalidToken
	}

	// 3. Check if expired
	if storedToken.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrInvalidToken
	}

	// 4. Get user
	user, err := u.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}

	// 5. Check if user is active
	if !isUserActive(user) {
		return nil, ErrUserNotActive
	}

	// 6. Rotate token:
	// Revoke the old token
	err = u.queries.RevokeRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new pair
	return u.generateAuthResponse(ctx, user)
}

func (u *authUseCase) Logout(ctx context.Context, tokenString string) error {
	// With stateless access tokens, we can't really "revoke" them without a blacklist.
	// But we CAN revoke the refresh token if provided. 
	// However, the Logout endpoint usually receives the Access Token in Authorization header.
	// If the frontend also sends the refresh token (e.g. via cookie), we can revoke it.
	
	// Since this method signature only takes a "tokenString" (which is typically the access token from the handler),
	// and we are moving to HttpOnly cookies for refresh tokens, the Handler should handle reading the cookie
	// and passing the refresh token to a Revoke method.
	
	// BUT, `IAuthUseCase.Logout` signature in this codebase seems to have been designed for Access Token blacklisting (Redis).
	// We will repurpose it or add a new method `RevokeRefreshToken`.
	
	// For now, let's assume the Handler will call a new method or we change this one.
	// Let's implement a specific RevokeRefreshToken method in the interface if possible, 
	// or just overload this one if `tokenString` is the refresh token.
	
	// Given the interface `Logout(ctx, token string) error`, let's assume `token` here IS the refresh token 
	// passed from the handler (extracted from cookie).
	
	return u.queries.RevokeRefreshToken(ctx, tokenString)
}

func (u *authUseCase) generateAuthResponse(ctx context.Context, user *models.User) (*dto.AuthResponse, error) {
	// Generate access token
	td, err := u.jwtProv.GenerateToken(user.ID, user.Role, u.cfg.AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (UUID)
	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(u.cfg.RefreshTokenDuration)

	// Store refresh token in DB
	_, err = u.queries.CreateRefreshToken(ctx, models.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: pgTime(expiresAt),
	})
	if err != nil {
		return nil, err
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

// Helper for pgtype
func pgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
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
