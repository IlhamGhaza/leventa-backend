package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	v := validator.New()
	return &Validator{validate: v}
}

// Validate validates a struct and returns a map of field errors.
func (v *Validator) Validate(data interface{}) map[string]string {
	errs := v.validate.Struct(data)
	if errs == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range errs.(validator.ValidationErrors) {
		field := toSnakeCase(err.Field())
		errors[field] = formatError(err)
	}
	return errors
}

func formatError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", toSnakeCase(err.Field()))
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", err.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters", err.Param())
	case "uuid4", "uuid":
		return "Must be a valid UUID"
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", err.Param())
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", err.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", err.Param())
	default:
		return fmt.Sprintf("Failed validation: %s", err.Tag())
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
