package responses

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

// AuthMissing dumps a simple "missing header" message
// response (401) in the gin context.
func AuthMissing(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code": "authorization:missing-header",
	})
}

// AuthBadScheme dumps a simple "bad scheme" message
// response (400) in the gin context.
func AuthBadScheme(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code": "authorization:bad-scheme",
	})
}

// AuthSyntaxError dumps a simple "syntax error" message
// response (400) in the gin context.
func AuthSyntaxError(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code": "authorization:syntax-error",
	})
}

// AuthNotFound dumps a simple "not found" message
// response (401, for auth) in the gin context.
func AuthNotFound(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code": "authorization:not-found",
	})
}

// AuthForbidden dumps a simple "forbidden" message
// response (403) in the gin context.
func AuthForbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"code": "authorization:forbidden",
	})
}

// NotFound dumps a simple "not found" message
// response (404) in the gin context.
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"code": "not-found",
	})
}

// MethodNotAllowed dumps a simple "method not allowed"
// message response (405) in the gin context.
func MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"code": "method-not-allowed",
	})
}

// InternalError dumps a simple "internal error"
// message response (500) in the gin context.
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code": "internal-error",
	})
}

// OkWith dumps a 200 message with a given body.
// Most likely, the body will be a gin.H instance.
func OkWith(c *gin.Context, value any) {
	c.JSON(http.StatusOK, value)
}

// Ok dumps a simple "ok" message response (200)
// in the gin context
func Ok(c *gin.Context) {
	OkWith(c, gin.H{
		"code": "ok",
	})
}

// Created dumps a simple "created" message with the
// id of the created object.
func Created(c *gin.Context, id primitive.ObjectID) {
	c.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}

// UnexpectedFormat dumps a simple "unexpected format"
// message response (400) in the gin context.
func UnexpectedFormat(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code": "format:unexpected",
	})
}

// InvalidFormat dumps an "invalid format" message
// response (400) in the gin context, with the errors
// that must flow to the user..
func InvalidFormat(c *gin.Context, errors any) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":   "format:invalid",
		"errors": errors,
	})
}

// AlreadyExists dumps a simple "already exists" message
// response (409) in the gin context.
func AlreadyExists(c *gin.Context) {
	c.JSON(http.StatusConflict, gin.H{
		"code": "already-exists",
	})
}

// DuplicateKey dumps a "duplicate key" message response
// (409) with the attempted key combination.
func DuplicateKey(c *gin.Context, key any) {
	c.JSON(http.StatusConflict, gin.H{
		"code": "duplicate-key",
		"key":  key,
	})
}
