package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	PasswordHash string
	Status       StatusUser // active/blocked/deleted
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type Email struct {
	ID         int64
	UserID     uuid.UUID
	Addr       string
	IsPrimary  bool
	IsVerified bool
	VerifiedAt *time.Time
}

type StatusUser string

const (
	StatusActive  StatusUser = "active"
	StatusBlocked StatusUser = "blocked"
	StatusDeleted StatusUser = "deleted"
)
