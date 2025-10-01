package validator

import (
	"fmt"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateStruct валидирует структуру и возвращает читаемую ошибку
func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var messages []string
	for _, validationErr := range err.(validator.ValidationErrors) {
		field := validationErr.Field()
		tag := validationErr.Tag()
		param := validationErr.Param()

		message := getValidationMessage(field, tag, param)
		messages = append(messages, message)
	}

	return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}

// getValidationMessage возвращает человекочитаемое сообщение об ошибке
func getValidationMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}
