package mapper

import (
	"tsuskills-user/internal/delivery/dto"
	"tsuskills-user/internal/domain"
)

func UserToRespUser(user domain.User, email domain.Email) dto.RespUser {
	return dto.RespUser{
		Name:       user.Name,
		Status:     string(user.Status),
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
		Email: dto.RespEmail{
			Addr:       email.Addr,
			IsPrimary:  email.IsPrimary,
			IsVerified: email.IsPrimary,
		},
	}
}
