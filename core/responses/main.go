package responses

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"net/http"
	"strings"
)

// AuthMissing dumps a simple "missing header" message
// response (401) in the gin context.
func AuthMissing(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, echo.Map{
		"code": "authorization:missing-header",
	})
}

// AuthBadScheme dumps a simple "bad scheme" message
// response (400) in the gin context.
func AuthBadScheme(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, echo.Map{
		"code": "authorization:bad-scheme",
	})
}

// AuthSyntaxError dumps a simple "syntax error" message
// response (400) in the gin context.
func AuthSyntaxError(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, echo.Map{
		"code": "authorization:syntax-error",
	})
}

// AuthNotFound dumps a simple "not found" message
// response (401, for auth) in the gin context.
func AuthNotFound(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, echo.Map{
		"code": "authorization:not-found",
	})
}

// AuthForbidden dumps a simple "forbidden" message
// response (403) in the gin context.
func AuthForbidden(c echo.Context) error {
	return c.JSON(http.StatusForbidden, echo.Map{
		"code": "authorization:forbidden",
	})
}

// NotFound dumps a simple "not found" message
// response (404) in the gin context.
func NotFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, echo.Map{
		"code": "not-found",
	})
}

// MethodNotAllowed dumps a simple "method not allowed"
// message response (405) in the gin context.
func MethodNotAllowed(c echo.Context) error {
	return c.JSON(http.StatusMethodNotAllowed, echo.Map{
		"code": "method-not-allowed",
	})
}

// InternalError dumps a simple "internal error"
// message response (500) in the gin context.
func InternalError(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, echo.Map{
		"code": "internal-error",
	})
}

// OkWith dumps a 200 message with a given body.
// Most likely, the body will be a echo.Map instance.
func OkWith(c echo.Context, value any) error {
	return c.JSON(http.StatusOK, value)
}

// Ok dumps a simple "ok" message response (200)
// in the gin context
func Ok(c echo.Context) error {
	return OkWith(c, echo.Map{
		"code": "ok",
	})
}

// Created dumps a simple "created" message with the
// id of the created object.
func Created(c echo.Context, id primitive.ObjectID) error {
	return c.JSON(http.StatusCreated, echo.Map{
		"id": id,
	})
}

// UnexpectedFormat dumps a simple "unexpected format"
// message response (400) in the gin context.
func UnexpectedFormat(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, echo.Map{
		"code": "format:unexpected",
	})
}

// InvalidFormat dumps an "invalid format" message
// response (400) in the gin context, with the errors
// that must flow to the user.
func InvalidFormat(c echo.Context, errors validator.ValidationErrors) error {
	errorMessages := make([]string, len(errors))
	for index, value := range errors {
		errorMessages[index] = fmt.Sprintf(
			"%s: %s", strings.SplitN(value.Namespace(), ".", 2)[1], value.Tag(),
		)
	}
	return c.JSON(http.StatusBadRequest, echo.Map{
		"code":   "format:invalid",
		"errors": errorMessages,
	})
}

// AlreadyExists dumps a simple "already exists" message
// response (409) in the gin context.
func AlreadyExists(c echo.Context) error {
	return c.JSON(http.StatusConflict, echo.Map{
		"code": "already-exists",
	})
}

// DuplicateKey dumps a "duplicate key" message response
// (409) with the attempted key combination.
func DuplicateKey(c echo.Context) error {
	return c.JSON(http.StatusConflict, echo.Map{
		"code": "duplicate-key",
	})
}

// FindOneOperationError is a helper to render an error
// for a MongoDB's FindOne operation on a collection.
func FindOneOperationError(c echo.Context, err error, logger ...*slog.Logger) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return NotFound(c)
	} else {
		if len(logger) > 0 {
			logger[0].Info("An error occurred: " + err.Error())
		}
		return InternalError(c)
	}
}
