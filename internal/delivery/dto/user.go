package dto

import "time"

type UserResponse struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Status     string       `json:"status"`
	IsVerified bool         `json:"is_verified"`
	Email      *EmailResponse `json:"email,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

type EmailResponse struct {
	Address    string `json:"address"`
	IsPrimary  bool   `json:"is_primary"`
	IsVerified bool   `json:"is_verified"`
}

type UpdateUserRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=100"`
	NewPassword string `json:"new_password" validate:"omitempty,min=6,max=72"`
}
