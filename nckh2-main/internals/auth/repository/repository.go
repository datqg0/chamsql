package repository

import (
	"context"
	"fmt"

	"backend/db"
	"backend/sql/models"

	"github.com/jackc/pgx/v5"
)

// IAuthRepository interface for auth operations
type IAuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	CreateUser(ctx context.Context, params models.CreateUserParams) (*models.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}

type authRepository struct {
	db      *db.Database
	queries *models.Queries
}

func NewAuthRepository(database *db.Database) IAuthRepository {
	return &authRepository{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByIdentifier(ctx context.Context, identifier string) (*models.User, error) {
	user, err := r.queries.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) CreateUser(ctx context.Context, params models.CreateUserParams) (*models.User, error) {
	// Default role is 'student'
	if params.Role == "" {
		params.Role = "student"
	}

	user, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (r *authRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	exists, err := r.queries.EmailExists(ctx, email)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *authRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	exists, err := r.queries.UsernameExists(ctx, username)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	return err == pgx.ErrNoRows
}
