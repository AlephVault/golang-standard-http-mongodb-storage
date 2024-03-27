package app

import (
	"errors"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/auth"
	"github.com/AlephVault/golang-standard-http-mongodb-storage/core/responses"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"regexp"
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
func checkId(ctx echo.Context, arg string, raiseNotFoundOnError bool) (primitive.ObjectID, bool, error) {
	idParam := ctx.Param(arg)
	if id, err := primitive.ObjectIDFromHex(idParam); err != nil {
		if raiseNotFoundOnError {
			return primitive.NilObjectID, false, responses.NotFound(ctx)
		}
		return primitive.NilObjectID, false, nil
	} else {
		return id, true, err
	}
}

// authenticate performs an authentication and permissions check.
func authenticate(ctx echo.Context, collection *mongo.Collection, key, permission string) (bool, error) {
	token := ctx.Request().Header.Get("Authorization")
	if token == "" {
		return false, responses.AuthMissing(ctx)
	}
	if !strings.HasPrefix(token, "Bearer ") {
		return false, responses.AuthBadScheme(ctx)
	}

	token = token[7:]
	var tokenRecord auth.AuthToken
	if result := collection.FindOne(ctx.Request().Context(), bson.M{
		"api-key": token, "valid_until": bson.M{
			"$not": bson.M{"$lt": time.Now()},
		},
	}); result.Err() != nil {
		return false, responses.AuthNotFound(ctx)
	} else if err := result.Decode(&tokenRecord); err != nil {
		return false, responses.InternalError(ctx)
	}

	hasPermission := false
	if globalPermissions, ok := tokenRecord.Permissions["*"]; ok {
		if globalPermissionsArray, ok := globalPermissions.(primitive.A); ok {
			hasPermission = checkPermission(permission, globalPermissionsArray)
		}
	}
	if !hasPermission {
		if localPermissions, ok := tokenRecord.Permissions[key]; ok {
			if localPermissionsArray, ok := localPermissions.(primitive.A); ok {
				hasPermission = checkPermission(permission, localPermissionsArray)
			}
		}
	}
	if !hasPermission {
		return false, responses.AuthForbidden(ctx)
	}

	return true, nil
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
