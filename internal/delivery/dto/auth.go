package dto

type JWTToken struct {
	Token string `json:"jwt_token" validate:"required"` // Валидация для вида, потом поменять required на jwt_format
}

type LoginRequest struct {
	Email string `json:"email" validate:"required,email"`
	Pswd  string `json:"pswd" validate:"required,min=3,max=25"`
}

type RegistrationRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=25,ascii"`
	Email string `json:"email" validate:"required,email"`
	Pswd  string `json:"pswd" validate:"required,min=3,max=25"`
}
