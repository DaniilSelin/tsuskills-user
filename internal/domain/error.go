package domain

import (
	"errors"
)

// repository-level errors
var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
	ErrInternal     = errors.New("internal server error")
)

// auth / security errors
var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrExpiredToken       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrHashPassword       = errors.New("error hashing password")
	ErrGenerateToken      = errors.New("failed to generate token")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
)

// service-layer error codes
type ErrorCode string

const (
	CodeOK                 ErrorCode = ""
	CodeInternal           ErrorCode = "INTERNAL_ERROR"
	CodeNotFound           ErrorCode = "NOT_FOUND"
	CodeConflict           ErrorCode = "CONFLICT"
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeInvalidRequestBody ErrorCode = "INVALID_REQUEST_BODY"
	CodeInvalidLimit       ErrorCode = "INVALID_LIMIT"
	CodeForbidden          ErrorCode = "FORBIDDEN"
)
