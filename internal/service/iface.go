package service

import (
	"context"
	"tsuskills-user/internal/domain"

	"github.com/google/uuid"
)

type IUserRepository interface {
	BeginTX(ctx context.Context) (domain.TxExecutor, error)
	CreateUserAndEmail(
		context.Context,
		domain.TxExecutor,
		domain.User,
		domain.Email) (uuid.UUID, error)
	GetUser(
		context.Context,
		uuid.UUID) (*domain.User, *domain.Email, domain.ErrorCode)
	UpdateUser(context.Context, *domain.User) error
	DeleteUser(context.Context, int) error
	GetUserByEmail(context.Context, string) (string, int, error)
}

type ISecurity interface {
	GetHashPswd(string) (string, error)
	CompareHashAndPassword(string, string) error
	GenerateToken(uuid.UUID) (string, error)
	ValidateToken(string) (*domain.Claims, error)
	RefreshToken(string) (string, error)
}
