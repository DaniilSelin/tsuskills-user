package service

import (
	"context"

	"tsuskills-user/internal/domain"

	"github.com/google/uuid"
)

type IUserRepository interface {
	CreateUserAndEmail(ctx context.Context, user domain.User, email domain.Email) (uuid.UUID, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, *domain.Email, error)
	GetUserByEmail(ctx context.Context, emailAddr string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	SoftDeleteUser(ctx context.Context, id uuid.UUID) error
	EmailExists(ctx context.Context, addr string) (bool, error)
}

type ISecurity interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hashedPassword, password string) error
	GenerateAccessToken(userID uuid.UUID) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateToken(tokenString string) (*domain.Claims, error)
	RefreshAccessToken(refreshTokenString string) (string, *domain.Claims, error)
}
