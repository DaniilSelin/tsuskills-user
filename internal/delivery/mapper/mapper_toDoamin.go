package mapper

import (
	"tsuskills-user/internal/delivery/dto"
	"tsuskills-user/internal/domain"

	"github.com/google/uuid"
)

func LoginReqToLoginModel(req dto.LoginRequest) domain.LoginRequest {
	return domain.LoginRequest{
		Email: req.Email,
		Pswd:  req.Pswd,
	}
}

func RegRequestToUser(req dto.RegistrationRequest) domain.RegistrationRequest {
	return domain.RegistrationRequest{
		Name:  req.Name,
		Email: req.Email,
		Pswd:  req.Pswd,
	}
}

func UserIdToUUID(req dto.UserID) uuid.UUID {
	return req.ID
}
