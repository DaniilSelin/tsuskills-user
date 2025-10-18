package service

import (
	"time"
	"tsuskills-user/config"
	"tsuskills-user/internal/domain"
	"tsuskills-user/internal/logger"

	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService struct {
	repo IUserRepository
	sec  ISecurity
	log  logger.Logger
	cfg  *config.Config
}

func NewUserService(repo IUserRepository, security ISecurity, log logger.Logger, cfg *config.Config) *UserService {
	return &UserService{
		repo: repo,
		sec:  security,
		log:  log,
		cfg:  cfg,
	}
}

// Логирование + дока

// Register - получает уже валидированные данные с верхнего уровня.
// Возвращает
func (s *UserService) Register(ctx context.Context, registry domain.RegistrationRequest) (uuid.UUID, string, domain.ErrorCode) {
	s.log.Info(
		ctx,
		"Register: invoke",
	)

	hashedPswd, err := s.sec.GetHashPswd(registry.Pswd)
	if err != nil {
		s.log.Error(
			ctx,
			"Register: error hashing password",
			zap.Error(err),
		)
		return uuid.Nil, "", domain.CodeInternal
	}

	var user = domain.User{
		ID:         uuid.New(),
		Name:       registry.Name,
		HashPswd:   hashedPswd,
		Status:     domain.Active,
		IsVerified: false,
		CreatedAt:  time.Now(),
	}

	var email = domain.Email{
		UserID:     user.ID,
		Addr:       registry.Email,
		IsPrimary:  true,
		IsVerified: false,
	}

	tx, err := s.repo.BeginTX(ctx)

	userID, err := s.repo.CreateUserAndEmail(ctx, tx, user, email)
	if err != nil {
		s.log.Error(
			ctx,
			"Register: error creating user or email",
			zap.Error(err),
		)
		return uuid.Nil, "", domain.CodeInternal
	}

	token, err := s.sec.GenerateToken(userID)
	if err != nil {
		s.log.Error(
			ctx,
			"Register: error generating token",
			zap.Error(err),
		)
		return uuid.Nil, "", domain.CodeInternal
	}

	return userID, token, domain.CodeOK
}

// Доделать

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, *domain.Email, domain.ErrorCode) {
	s.log.Info(
		ctx,
		"GetUser: invoke",
	)

	u, e, err := s.repo.GetUser(ctx, id)
	if err != domain.CodeOK {
		s.log.Error(
			ctx,
			"GetUser: invoke",
		)
	}

	return u, e, err
}

func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	if user.HashPswd != "" {
		hashedPassword, err := s.sec.GetHashPswd(user.HashPswd)
		if err != nil {
			return errdefs.Wrapf(errdefs.ErrGetHashPswd, "error hashing password: %w", err)
		}
		user.Password = hashedPassword
	}
	return s.repo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) Login(ctx context.Context, email, pswd string) (string, error) {
	passwordRef, userId, err := s.repo.GetUserByEmail(ctx, email)

	if err != nil {
		if errdefs.Is(err, errdefs.ErrNotFound) {
			return "", errdefs.ErrNotFound
		}
		return "", errdefs.ErrDB
	}

	err = s.security.CompareHashAndPassword(passwordRef, pswd)
	if err != nil {
		return "", errdefs.ErrInvalidCredentials
	}

	token, err := s.security.GenerateToken(userId)
	if err != nil {
		// Сменить ошибку
		return "", errdefs.ErrGenerateToken
	}

	return token, nil
}

func (s *UserService) Auth(ctx context.Context, token string) (int, error) {
	claims, err := s.security.ValidateToken(token)
	if err != nil {
		if errdefs.Is(err, errdefs.ErrExpiredToken) {
			return -1, errdefs.ErrExpiredToken
		}
		if errdefs.Is(err, errdefs.ErrInvalidToken) {
			return -1, errdefs.ErrInvalidToken
		}
		return -1, errdefs.Wrapf(errdefs.ErrByScript, "token validation error: %w", err)
	}

	userID := claims.UserID

	_, err = s.repo.GetUser(ctx, userID)
	if err != nil {
		if errdefs.Is(err, errdefs.ErrNotFound) {
			return -1, errdefs.ErrNotFound
		}
		return -1, errdefs.ErrDB
	}

	return userID, nil
}

func (s *UserService) RefreshToken(ctx context.Context, token string) (string, error) {
	claims, err := s.security.ValidateToken(token)
	if err != nil && !errdefs.Is(err, errdefs.ErrExpiredToken) {
		return "", errdefs.ErrInvalidToken
	}

	userID := claims.UserID
	_, err = s.repo.GetUser(ctx, userID)
	if err != nil {
		if errdefs.Is(err, errdefs.ErrNotFound) {
			return "", errdefs.ErrNotFound
		}
		return "", errdefs.ErrDB
	}

	newToken, err := s.security.RefreshToken(token)
	if err != nil {
		return "", errdefs.Wrapf(errdefs.ErrByScript, "failed to refresh token: %w", err)
	}

	return newToken, nil
}
