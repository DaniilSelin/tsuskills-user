package domain

import (
	"errors"
)

var (
	//общие ошибки
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInternal     = errors.New("internal server error")

	// ошибки пакета security
	ErrGetHashPswd   = errors.New("error hashing password")
	ErrGenerateToken = errors.New("failed to generate token")
	ErrByScript      = errors.New("byscript error")

	// ошибки на уровне БД (repository)
	ErrDB = errors.New("DB error")

	// ошибки слоя бизнесс логики (service)
	ErrExpiredToken       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// коды ошибок слоя бизнесс логики
type ErrorCode string

const (
	CodeOK                 ErrorCode = ""
	CodeInternal           ErrorCode = "INTERNAL_ERROR"
	CodeInvalidLimit       ErrorCode = "INVALID_LIMIT"
	CodeInvalidRequestBody ErrorCode = "INVALID_REQUEST_BODY"
)
