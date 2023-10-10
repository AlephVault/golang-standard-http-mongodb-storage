package validation

import (
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

	switch fl.Field().Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		// Nothing here
	default:
		panic("require_iif can only be applied to lists, slices and maps")
	}

	if len(params) != 2 {
		panic("require_iif only allows a parameter like this: SomeIntLikeField someIntLikeValue")
	}

	fieldName, paramValueStr := params[0], params[1]

	// Use reflection to get the field value dynamically
	rv := reflect.ValueOf(fl.Parent().Interface())
	field := rv.FieldByName(fieldName)

	// Make sure field exists and is an integer type
	var actualFieldValue int = 0
	var paramValue int = 0
	if !field.IsValid() {
		panic("require_iif requires a valid int-like field to be specified")
	} else if field.Kind() == reflect.Int || field.Kind() == reflect.Int64 {
		paramValue, err := strconv.Atoi(paramValueStr)
		if err != nil || paramValue < 0 {
			panic("require_iif requires the comparison value to be a valid non-negative integer number")
		}
		actualFieldValue = int(field.Int())
	} else if field.Kind() == reflect.Uint || field.Kind() == reflect.Uint64 {
		paramValue, err := strconv.ParseUint(paramValueStr, 10, 64)
		if err != nil {
			panic("require_iif requires the comparison value to be a valid non-negative integer number")
		} else if paramValue > (^uint64(0) >> 1) {
			return true
		}
		actualFieldValue = int(field.Uint())
	} else {
		panic("require_iif requires a valid int-like field to be specified")
	}

	// Determine the length of the map, slice, or array
	var length int = fl.Field().Len()

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
