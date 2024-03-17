package app

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"regexp"
	"standard-http-mongodb-storage/core/auth"
	"standard-http-mongodb-storage/core/responses"
	"strings"
	"time"
)

// checkPermission checks whether a permission is in the list of permissions.
func checkPermission(permission string, existingPermissions []any) bool {
	for _, existingPermission := range existingPermissions {
		if perm, ok := existingPermission.(string); ok && (perm == "*" || perm == permission) {
			return true
		}
	}
	return false
}

// checkId ensures the :id is valid.
func checkId(ctx *gin.Context, arg string, raiseNotFoundOnError bool) (primitive.ObjectID, bool) {
	idParam := ctx.Param(arg)
	if id, err := primitive.ObjectIDFromHex(idParam); err != nil {
		if raiseNotFoundOnError {
			responses.NotFound(ctx)
		}
		return primitive.NilObjectID, false
	} else {
		return id, true
	}
}

// authenticate performs an authentication and permissions check.
func authenticate(ctx *gin.Context, collection *mongo.Collection, key, permission string) bool {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		responses.AuthMissing(ctx)
		return false
	}
	if !strings.HasPrefix(token, "Bearer ") {
		responses.AuthBadScheme(ctx)
		return false
	}

	token = token[7:]
	var tokenRecord auth.AuthToken
	if result := collection.FindOne(ctx, bson.M{
		"api-key": token, "valid_until": bson.M{
			"$not": bson.M{"$lt": time.Now()},
		},
	}); result.Err() != nil {
		responses.AuthNotFound(ctx)
		return false
	} else if err := result.Decode(&tokenRecord); err != nil {
		responses.InternalError(ctx)
		return false
	}

	hasPermission := false
	if globalPermissions, ok := tokenRecord.Permissions["*"]; ok {
		if globalPermissionsArray, ok := globalPermissions.([]any); ok {
			hasPermission = checkPermission(permission, globalPermissionsArray)
		}
	}
	if !hasPermission {
		if localPermissions, ok := tokenRecord.Permissions[key]; ok {
			if localPermissionsArray, ok := localPermissions.([]any); ok {
				hasPermission = checkPermission(permission, localPermissionsArray)
			}
		}
	}
	if !hasPermission {
		responses.AuthForbidden(ctx)
		return false
	}

	return true
}

var rxDuplicateError = regexp.MustCompile(`E11000|E11001|E12582|16460`)

// Tells whether an error is a Duplicate Key error.
func isDuplicateKeyError(err error) bool {
	var writeException mongo.WriteException
	if errors.As(err, &writeException) {
		for _, we := range writeException.WriteErrors {
			code := we.Code
			if code == 11000 || code == 11001 || code == 12582 || code == 16460 {
				return true
			}
		}
	}

	// Fallback to checking the error message directly if it's not a WriteException
	return rxDuplicateError.MatchString(err.Error())
}
