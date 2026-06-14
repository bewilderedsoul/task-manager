// Package validate wraps go-playground/validator and turns its errors into a
// flat, client-friendly field->message map.
package validate

import (
	"github.com/go-playground/validator/v10"
)

var v = validator.New()

// FieldErrors maps a field name to a human-readable validation message.
type FieldErrors map[string]string

// Struct validates s using its `validate` struct tags. It returns nil when the
// value is valid, or a FieldErrors describing each failing field.
func Struct(s any) FieldErrors {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	out := FieldErrors{}
	for _, fe := range err.(validator.ValidationErrors) {
		out[fe.Field()] = message(fe)
	}
	return out
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "is invalid"
	}
}
