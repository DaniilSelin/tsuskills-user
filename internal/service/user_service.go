package service

import (
	"context"
	"errors"
	"time"

	"tsuskills-user/internal/domain"
	"tsuskills-user/internal/infra/kafka"
	"tsuskills-user/internal/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService struct {
	repo      IUserRepository
	sec       ISecurity
	publisher kafka.Publisher
	log       logger.Logger
}

func NewUserService(repo IUserRepository, sec ISecurity, publisher kafka.Publisher, log logger.Logger) *UserService {
	return &UserService{repo: repo, sec: sec, publisher: publisher, log: log}
}

// Register создаёт нового пользователя, хеширует пароль, сохраняет в БД,
// возвращает access и refresh токены.
func (s *UserService) Register(
	ctx context.Context,
	req domain.RegistrationRequest,
) (uuid.UUID, string, string, domain.ErrorCode) {
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		s.log.Error(ctx, "Register: check email exists", zap.Error(err))
		return uuid.Nil, "", "", domain.CodeInternal
	}
	if exists {
		s.log.Warn(ctx, "Register: email already taken", zap.String("email", req.Email))
		return uuid.Nil, "", "", domain.CodeConflict
	}

	hashedPswd, err := s.sec.HashPassword(req.Password)
	if err != nil {
		s.log.Error(ctx, "Register: hash password", zap.Error(err))
		return uuid.Nil, "", "", domain.CodeInternal
	}

	now := time.Now()
	user := domain.User{
		ID:           uuid.New(),
		Name:         req.Name,
		PasswordHash: hashedPswd,
		Status:       domain.StatusActive,
		IsVerified:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	email := domain.Email{
		UserID:     user.ID,
		Addr:       req.Email,
		IsPrimary:  true,
		IsVerified: false,
	}

	userID, err := s.repo.CreateUserAndEmail(ctx, user, email)
	if err != nil {
		s.log.Error(ctx, "Register: create user", zap.Error(err))
		return uuid.Nil, "", "", domain.CodeInternal
	}

	if err := s.publishUserEvent(ctx, kafka.EventUserCreated, userID.String(), map[string]interface{}{
		"name":        user.Name,
		"status":      user.Status,
		"is_verified": user.IsVerified,
	}); err != nil {
		s.log.Warn(ctx, "Register: publish user created event failed", zap.Error(err))
	}

	accessToken, err := s.sec.GenerateAccessToken(userID)
	if err != nil {
		s.log.Error(ctx, "Register: generate access token", zap.Error(err))
		return uuid.Nil, "", "", domain.CodeInternal
	}

	refreshToken, err := s.sec.GenerateRefreshToken(userID)
	if err != nil {
		s.log.Error(ctx, "Register: generate refresh token", zap.Error(err))
		return uuid.Nil, "", "", domain.CodeInternal
	}

	s.log.Info(ctx, "Register: success", zap.String("user_id", userID.String()))
	return userID, accessToken, refreshToken, domain.CodeOK
}

// Login проверяет email+пароль, возвращает access и refresh токены.
func (s *UserService) Login(
	ctx context.Context,
	req domain.LoginRequest,
) (string, string, domain.ErrorCode) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.log.Warn(ctx, "Login: user not found", zap.String("email", req.Email))
			return "", "", domain.CodeInvalidCredentials
		}
		s.log.Error(ctx, "Login: get user by email", zap.Error(err))
		return "", "", domain.CodeInternal
	}

	if err := s.sec.CompareHashAndPassword(user.PasswordHash, req.Password); err != nil {
		s.log.Warn(ctx, "Login: wrong password", zap.String("email", req.Email))
		return "", "", domain.CodeInvalidCredentials
	}

	accessToken, err := s.sec.GenerateAccessToken(user.ID)
	if err != nil {
		s.log.Error(ctx, "Login: generate access token", zap.Error(err))
		return "", "", domain.CodeInternal
	}

	refreshToken, err := s.sec.GenerateRefreshToken(user.ID)
	if err != nil {
		s.log.Error(ctx, "Login: generate refresh token", zap.Error(err))
		return "", "", domain.CodeInternal
	}

	s.log.Info(ctx, "Login: success", zap.String("user_id", user.ID.String()))
	return accessToken, refreshToken, domain.CodeOK
}

// Auth валидирует access token и возвращает user ID.
func (s *UserService) Auth(ctx context.Context, token string) (uuid.UUID, domain.ErrorCode) {
	claims, err := s.sec.ValidateToken(token)
	if err != nil {
		if errors.Is(err, domain.ErrExpiredToken) {
			return uuid.Nil, domain.CodeUnauthorized
		}
		return uuid.Nil, domain.CodeUnauthorized
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, domain.CodeUnauthorized
	}

	// проверяем что пользователь ещё существует
	_, _, repoErr := s.repo.GetUserByID(ctx, userID)
	if repoErr != nil {
		if errors.Is(repoErr, domain.ErrNotFound) {
			return uuid.Nil, domain.CodeNotFound
		}
		return uuid.Nil, domain.CodeInternal
	}

	return userID, domain.CodeOK
}

// RefreshToken обновляет access token по refresh token.
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (string, domain.ErrorCode) {
	newAccess, claims, err := s.sec.RefreshAccessToken(refreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrExpiredToken) {
			return "", domain.CodeUnauthorized
		}
		return "", domain.CodeUnauthorized
	}

	// проверяем что пользователь ещё существует
	userID, parseErr := uuid.Parse(claims.UserID)
	if parseErr != nil {
		return "", domain.CodeUnauthorized
	}
	_, _, repoErr := s.repo.GetUserByID(ctx, userID)
	if repoErr != nil {
		return "", domain.CodeUnauthorized
	}

	s.log.Info(ctx, "RefreshToken: success", zap.String("user_id", claims.UserID))
	return newAccess, domain.CodeOK
}

// GetUser возвращает пользователя по ID.
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, *domain.Email, domain.ErrorCode) {
	user, email, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil, domain.CodeNotFound
		}
		s.log.Error(ctx, "GetUser: repo error", zap.Error(err))
		return nil, nil, domain.CodeInternal
	}
	return user, email, domain.CodeOK
}

// UpdateUser обновляет имя пользователя. Если передан новый пароль — хеширует.
func (s *UserService) UpdateUser(ctx context.Context, user *domain.User, newPassword string) domain.ErrorCode {
	if newPassword != "" {
		hashed, err := s.sec.HashPassword(newPassword)
		if err != nil {
			s.log.Error(ctx, "UpdateUser: hash password", zap.Error(err))
			return domain.CodeInternal
		}
		user.PasswordHash = hashed
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.CodeNotFound
		}
		s.log.Error(ctx, "UpdateUser: repo error", zap.Error(err))
		return domain.CodeInternal
	}

	if err := s.publishUserEvent(ctx, kafka.EventUserUpdated, user.ID.String(), map[string]interface{}{
		"name":        user.Name,
		"status":      user.Status,
		"is_verified": user.IsVerified,
	}); err != nil {
		s.log.Warn(ctx, "UpdateUser: publish user updated event failed", zap.Error(err))
	}

	return domain.CodeOK
}

// DeleteUser выполняет мягкое удаление пользователя.
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) domain.ErrorCode {
	if err := s.repo.SoftDeleteUser(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.CodeNotFound
		}
		s.log.Error(ctx, "DeleteUser: repo error", zap.Error(err))
		return domain.CodeInternal
	}

	if err := s.publishUserEvent(ctx, kafka.EventUserDeleted, id.String(), nil); err != nil {
		s.log.Warn(ctx, "DeleteUser: publish user deleted event failed", zap.Error(err))
	}

	return domain.CodeOK
}

func (s *UserService) publishUserEvent(ctx context.Context, eventType, entityID string, payload map[string]interface{}) error {
	event, err := kafka.NewEvent(eventType, kafka.EntityUser, entityID, payload)
	if err != nil {
		return err
	}
	return s.publisher.Publish(ctx, event)
}
