package security

import (
	"errors"
	"tsuskills-user/config"
	"tsuskills-user/internal/domain"

	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Использованные ошибки -
// ErrInvalidToken
// ErrExpiredToken

type Security struct {
	cfg *config.Config
}

func NewSecurity(cfg *config.Config) *Security {
	return &Security{
		cfg: cfg,
	}
}

func (s *Security) GetHashPswd(pswd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pswd), 12)
	if err != nil {
		// Обрабатывать будем в service
		return "", fmt.Errorf("error hashing password: %w", err)
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

func (s *Security) GenerateToken(userID int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.cfg.JWT.Expiration) * time.Minute)

	claims := &domain.Claims{
		UserID:  userID,
		TokenID: generateRandomID(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.SecKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, nil
}

func (s *Security) ValidateToken(tokenString string) (*domain.Claims, error) {
	claims := &domain.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.SecKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrExpiredToken
		}
		return nil, domain.ErrInvalidToken
	}

	if !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func (s *Security) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil && !errors.Is(err, domain.ErrExpiredToken) {
		return "", err
	}

	expirationTime := time.Now().Add(time.Duration(s.cfg.JWT.Expiration) * time.Minute)

	newClaims := &domain.Claims{
		UserID:  claims.UserID,
		TokenID: generateRandomID(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenString, err = token.SignedString([]byte(s.cfg.JWT.SecKey))
	if err != nil {
		return "", fmt.Errorf("error signing refreshed token: %w", err)
	}

	return tokenString, nil
}
