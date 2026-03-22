package mapper

import (
	"tsuskills-user/internal/delivery/dto"
	"tsuskills-user/internal/domain"
)

func LoginReqToDomain(req dto.LoginRequest) domain.LoginRequest {
	return domain.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
}

func RegReqToDomain(req dto.RegistrationRequest) domain.RegistrationRequest {
	return domain.RegistrationRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
}

func UserToResponse(user *domain.User, email *domain.Email) dto.UserResponse {
	resp := dto.UserResponse{
		ID:         user.ID.String(),
		Name:       user.Name,
		Status:     string(user.Status),
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

	if email != nil && email.Addr != "" {
		resp.Email = &dto.EmailResponse{
			Address:    email.Addr,
			IsPrimary:  email.IsPrimary,
			IsVerified: email.IsVerified,
		}
	}

	return resp
}
