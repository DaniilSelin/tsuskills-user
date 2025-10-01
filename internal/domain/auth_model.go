package domain

import (
	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Email string
	Pswd  string
}

type LoginResponse struct {
	Token string
}

type AuthResponse struct {
	UserID int
}

type RegistrationRequest struct {
	Name  string
	Email string
	Pswd  string
	Role  string
}

type RegistrationResponse struct {
	UserID int
	Token  string
}

type Claims struct {
	UserID  int
	TokenID string
	jwt.RegisteredClaims
}
