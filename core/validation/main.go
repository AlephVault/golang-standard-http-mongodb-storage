package validation

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
	"standard-http-mongodb-storage/core/dsl"
	"strconv"
	"strings"
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

// requiredIifValidator is used like this: require_iif=Foo {value}
// where the list/slice/map must have elements if and only if the
// field Foo (which is int-like) has a certain value.
func requiredIifValidator(fl validator.FieldLevel) bool {
	params := strings.Fields(fl.Param()) // Get the parameters and split by space

	if len(params) != 2 {
		panic("require_iif only allows a parameter like this: SomeIntLikeField someIntLikeValue")
	}

	fieldName, paramValueStr := params[0], params[1]

	// Use reflection to get the field value dynamically
	rv := reflect.ValueOf(fl.Parent().Interface())
	field := rv.FieldByName(fieldName)

	// Make sure field exists and is an integer type
	if !field.IsValid() {
		panic("require_iif requires a valid int-like field to be specified")
	} else if field.Kind() != reflect.Int && field.Kind() != reflect.Int64 &&
		field.Kind() != reflect.Uint && field.Kind() != reflect.Uint64 {
		panic("require_iif requires the kind to be [u]int ir [u]int64")
	}

	// Convert the value and understand a priori.
	paramValue, err := strconv.Atoi(paramValueStr)
	if err != nil {
		if !errors.Is(err, strconv.ErrRange) {
			panic("require_iif requires the comparison value to be a valid integer number")
		} else {
			return true
		}
	}

	// Get the actual field value.
	actualFieldValue := int(field.Int())

	// Determine the length of the map, slice, or array
	var length int
	switch fl.Field().Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		length = fl.Field().Len()
	default:
		return false
	}

	if actualFieldValue == paramValue {
		return length != 0
	}

	return length == 0
}

// makeValidator creates the validator we need for this all.
func makeValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("mdb-name", regexFunction(rxMongoName))
	v.RegisterValidation("mdb-index-entry", regexFunction(rxMongoIndexEntry))
	v.RegisterValidation("method-name", regexFunction(rxMethodName))
	v.RegisterValidation("verbs", dsl.ValidateVerbs)
	v.RegisterValidation("required_iif", requiredIifValidator)
	return v
}

var (
	currentValidator = makeValidator()
)

// Validator returns the only validator we need and will be using.
func Validator() *validator.Validate {
	return currentValidator
}
