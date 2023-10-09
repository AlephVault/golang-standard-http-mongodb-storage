package validation

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
)

var (
	rxMongoName = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]+$")
)

// regexFunction creates a new regex-validator function.
func regexFunction(regex *regexp.Regexp) func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()
		switch field.Kind() {
		case reflect.String:
			return regex.MatchString(field.String())
		default:
			return false
		}
	}
}

// Validate uses the validator framework, actually.
// The validation is run against the class, so no
// further schema is needed.
func Validate(value any) error {
	v := validator.New()
	v.RegisterValidation("mdb-name", regexFunction(rxMongoName))
	return (validator.New()).Struct(value)
}
