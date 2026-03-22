package validator

import (
	"fmt"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var messages []string
	for _, ve := range err.(validator.ValidationErrors) {
		messages = append(messages, formatMessage(ve.Field(), ve.Tag(), ve.Param()))
	}

	return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}

func formatMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}
