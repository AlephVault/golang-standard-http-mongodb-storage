package validation

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
)

var (
	rxMongoName       = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]+$")
	rxMethodName      = rxMongoName
	rxMongoIndexEntry = regexp.MustCompile("^[#~@-]?[a-zA-Z][a-zA-Z0-9_-]*$")
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

// makeValidator creates the validator we need for this all.
func makeValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("mdb-name", regexFunction(rxMongoName))
	v.RegisterValidation("mdb-index-entry", regexFunction(rxMongoIndexEntry))
	v.RegisterValidation("method-name", regexFunction(rxMethodName))
	return v
}

var (
	currentValidator = makeValidator()
)

// Validator returns the only validator we need and will be using.
func Validator() *validator.Validate {
	return currentValidator
}
