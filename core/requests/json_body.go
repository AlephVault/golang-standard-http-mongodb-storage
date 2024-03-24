package requests

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"standard-http-mongodb-storage/core/responses"
	"strings"
)

// ReadJSONBody reads a JSON body and, on error, aborts.
func ReadJSONBody(context echo.Context, validator_ *validator.Validate, body any) (bool, error) {
	if body == nil {
		panic("the body must not be nil")
	}

	if !strings.Contains(strings.ToLower(context.Request().Header.Get("Content-Type")), "application/json") {
		fmt.Println("lalalala 1")
		return false, responses.UnexpectedFormat(context)
	}

	if err := (&echo.DefaultBinder{}).BindBody(context, &body); err != nil {
		fmt.Println("lalalala 2")
		return false, responses.UnexpectedFormat(context)
	}

	if validator_ != nil {
		if err := validator_.Struct(body); err != nil {
			return false, responses.InvalidFormat(context, err.(validator.ValidationErrors))
		}
	}

	return true, nil
}
