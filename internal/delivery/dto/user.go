package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserCreateDTO — структура для создания вакансии
type ReqCreateUser struct {
	Name  string `json:"title" validate:"required,min=1,max=255"`
	Email string `json:"activity_type" validate:"email"`
}

type RespUser struct {
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Email      RespEmail `json:"email"`
	IsVerified bool      `json:"isVerified"`
	CreatedAt  time.Time `json:"createdAt"`
}

type RespEmail struct {
	Addr       string `json:"address"`
	IsPrimary  bool   `json:"isPrimare"`
	IsVerified bool   `json:"isVerified"`
}

type UserID struct {
	ID uuid.UUID `json:"id" validate:"required,uuid4"`
}
