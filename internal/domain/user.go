package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	Name       string
	HashPswd   string
	Status     StatusUser // active/blocked/deleted
	IsVerified bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
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

var (
	Active  StatusUser = "active"
	Blocked StatusUser = "blocked"
	Deleted StatusUser = "deleted"
)

// на случай нескольких способов авториазации
type Credentials struct {
	UserID            uuid.UUID
	PasswordHash      []byte // argon2id
	PasswordUpdatedAt *time.Time
	TOTPEnabled       bool
	TOTPSecretEnc     []byte
	RecoveryCodesHash [][]byte
	WebAuthnEnabled   bool
}

// для работы с refresh tokens
type Session struct {
	ID           int64
	UserID       uuid.UUID
	RefreshHash  []byte
	IssuedAt     time.Time
	ExpiresAt    time.Time
	LastSeenAt   *time.Time
	IP           string
	UserAgent    string
	IsRevoked    bool
	ReplacedByID *int64
}
