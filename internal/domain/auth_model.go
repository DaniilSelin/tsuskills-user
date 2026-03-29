package domain

import (
	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Email    string
	Password string
}

type RegistrationRequest struct {
	Name     string
	Email    string
	Password string
}

type Claims struct {
	UserID  string `json:"user_id"`
	TokenID string `json:"token_id"`
	jwt.RegisteredClaims
}
