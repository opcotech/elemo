package validate

import "github.com/go-playground/validator/v10"

var (
	validate = validator.New() // the validator instance
)

// Validator returns the validator instance.
func Validator() *validator.Validate {
	return validate
}

// Struct validates the given struct.
func Struct(s any) error {
	return validate.Struct(s)
}

// Var validates the given variable.
func Var(field any, tag string) error {
	return validate.Var(field, tag)
}
