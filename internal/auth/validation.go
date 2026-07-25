package auth

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationErrors(err error) map[string]string {
	errorMap := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		fieldName := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			errorMap[fieldName] = fmt.Sprintf("%s is required", HumanizeFieldName(fieldName))
		case "alphanum":
			errorMap[fieldName] = fmt.Sprintf("%s must contain only alphanumeric characters", HumanizeFieldName(fieldName))
		case "min":
			errorMap[fieldName] = fmt.Sprintf("%s must be at least %s characters long", HumanizeFieldName(fieldName), e.Param())
		case "max":
			errorMap[fieldName] = fmt.Sprintf("%s must not exceed %s characters", HumanizeFieldName(fieldName), e.Param())
		case "email":
			errorMap[fieldName] = "Please enter a valid email address"
		default:
			errorMap[fieldName] = e.Error()
		}
	}
	return errorMap
}

// HumanizeFieldName converts snake_case field names to readable format
func HumanizeFieldName(field string) string {
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
	case "project_name":
		return "Project name"
	case "project_desc":
		return "Project description"
	case "refresh_token":
		return "Refresh token"
	default:
		return field
	}
}
