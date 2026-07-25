package auth

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func formatValidationErrors(err error) map[string]string {
	errorMap := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		fieldName := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			errorMap[fieldName] = fmt.Sprintf("%s is required", humanizeFieldName(fieldName))
		case "alphanum":
			errorMap[fieldName] = fmt.Sprintf("%s must contain only alphanumeric characters", humanizeFieldName(fieldName))
		case "min":
			errorMap[fieldName] = fmt.Sprintf("%s must be at least %s characters long", humanizeFieldName(fieldName), e.Param())
		case "max":
			errorMap[fieldName] = fmt.Sprintf("%s must not exceed %s characters", humanizeFieldName(fieldName), e.Param())
		case "email":
			errorMap[fieldName] = "Please enter a valid email address"
		default:
			errorMap[fieldName] = e.Error()
		}
	}
	return errorMap
}

func humanizeFieldName(field string) string {
	switch field {
	case "username":
		return "Username"
	case "password":
		return "Password"
	case "email":
		return "Email"
	case "old_password":
		return "Old password"
	case "new_password":
		return "New password"
	default:
		return field
	}
}
