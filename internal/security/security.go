package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"tsuskills-user/config"
	"tsuskills-user/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Security struct {
	cfg *config.JWTConfig
}

func NewSecurity(cfg *config.JWTConfig) *Security {
	return &Security{cfg: cfg}
}

func (s *Security) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrHashPassword, err)
	}
	return string(hash), nil
}

func (s *Security) CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func generateRandomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Security) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return s.generateToken(userID, s.cfg.AccessExpiration)
}

func (s *Security) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return s.generateToken(userID, s.cfg.RefreshExpiration)
}

func (s *Security) generateToken(userID uuid.UUID, expiration time.Duration) (string, error) {
	claims := &domain.Claims{
		UserID:  userID.String(),
		TokenID: generateRandomID(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.SecretKey))
	if err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrGenerateToken, err)
	}
	return tokenString, nil
}

func (s *Security) ValidateToken(tokenString string) (*domain.Claims, error) {
	claims := &domain.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return claims, domain.ErrExpiredToken
		}
		return nil, domain.ErrInvalidToken
	}

	if !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func (s *Security) RefreshAccessToken(refreshTokenString string) (string, *domain.Claims, error) {
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return "", nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", nil, domain.ErrInvalidToken
	}

	newAccess, err := s.GenerateAccessToken(userID)
	if err != nil {
		return "", nil, err
	}

	return newAccess, claims, nil
}
