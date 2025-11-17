package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// validateRequest validates a struct and returns user-friendly error messages.
// Returns true if validation passes, false otherwise.
// If validation fails, error response is written to the ResponseWriter.
func validateRequest(w http.ResponseWriter, req interface{}) bool {
	if err := validate.Struct(req); err != nil {
		var errorMessages []string

		for _, err := range err.(validator.ValidationErrors) {
			var message string
			switch err.Tag() {
			case "required":
				message = fmt.Sprintf("%s is required", getFieldName(err.Field()))
			case "min":
				if err.Type().Kind().String() == "string" {
					message = fmt.Sprintf("%s must be at least %s characters", getFieldName(err.Field()), err.Param())
				} else {
					message = fmt.Sprintf("%s must be at least %s", getFieldName(err.Field()), err.Param())
				}
			case "max":
				if err.Type().Kind().String() == "string" {
					message = fmt.Sprintf("%s must be at most %s characters", getFieldName(err.Field()), err.Param())
				} else {
					message = fmt.Sprintf("%s must be at most %s", getFieldName(err.Field()), err.Param())
				}
			case "url":
				message = fmt.Sprintf("%s must be a valid URL", getFieldName(err.Field()))
			case "oneof":
				message = fmt.Sprintf("%s must be one of: %s", getFieldName(err.Field()), err.Param())
			default:
				message = fmt.Sprintf("%s is invalid", getFieldName(err.Field()))
			}
			errorMessages = append(errorMessages, message)
		}

		http.Error(w, strings.Join(errorMessages, "; "), http.StatusBadRequest)
		return false
	}
	return true
}

// getFieldName converts struct field name to a more user-friendly name
func getFieldName(field string) string {
	// Convert camelCase/PascalCase to Title Case with spaces
	var result strings.Builder
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteString(" ")
		}
		result.WriteRune(r)
	}
	return result.String()
}

