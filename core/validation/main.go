package validation

import "github.com/go-playground/validator/v10"

// Validate uses the validator framework, actually.
// The validation is run against the class, so no
// further schema is needed.
func Validate(value any) error {
	return (validator.New()).Struct(value)
}
